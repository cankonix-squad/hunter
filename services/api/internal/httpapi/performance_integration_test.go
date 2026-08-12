//go:build integration

package httpapi_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/collector"
	"github.com/cankonix/cankonix-oi/services/api/internal/httpapi"
	"github.com/cankonix/cankonix-oi/services/api/internal/identity"
	"github.com/cankonix/cankonix-oi/services/api/internal/opportunities"
	"github.com/cankonix/cankonix-oi/services/api/internal/platform/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	performanceRows     = 10000
	performanceRequests = 40
	performanceWorkers  = 8
	performanceP95      = 750 * time.Millisecond
)

func TestOpportunityPaginationLoadBudget(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL or DATABASE_URL is required for integration tests")
	}

	ctx := context.Background()
	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open performance database: %v", err)
	}
	t.Cleanup(db.Close)

	sourceID := seedPerformanceData(t, ctx, db)
	service := opportunities.New(db)

	first, err := service.List(ctx, opportunities.Query{SourceID: sourceID, Page: 1, PageSize: 100, MinimumScoreBP: -1})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if first.Total != performanceRows || len(first.Items) != 100 || first.TotalPages != 100 {
		t.Fatalf("first page contract: total=%d items=%d pages=%d", first.Total, len(first.Items), first.TotalPages)
	}

	last, err := service.List(ctx, opportunities.Query{SourceID: sourceID, Page: 100, PageSize: 100, MinimumScoreBP: -1})
	if err != nil {
		t.Fatalf("list last page: %v", err)
	}
	if len(last.Items) != 100 || last.Page != 100 {
		t.Fatalf("last page contract: page=%d items=%d", last.Page, len(last.Items))
	}

	review, err := service.List(ctx, opportunities.Query{SourceID: sourceID, Page: 1, PageSize: 100, MinimumScoreBP: -1, NeedsReview: "true"})
	if err != nil {
		t.Fatalf("list review page: %v", err)
	}
	if review.Total != performanceRows/5 || len(review.Items) != 100 {
		t.Fatalf("review page contract: total=%d items=%d", review.Total, len(review.Items))
	}

	search, err := service.List(ctx, opportunities.Query{SourceID: sourceID, Page: 1, PageSize: 20, MinimumScoreBP: -1, Search: "Load Package 09999"})
	if err != nil {
		t.Fatalf("search performance data: %v", err)
	}
	if search.Total != 1 || len(search.Items) != 1 || search.Items[0].PackageIdentity != "LOAD-09999" {
		t.Fatalf("search contract: total=%d items=%d", search.Total, len(search.Items))
	}

	password := randomCredential(t)
	identityService := identity.New(db, randomCredential(t)+randomCredential(t), 15*time.Minute, time.Hour)
	if err := identityService.BootstrapAdmin(ctx, "admin.performance@example.test", "Performance Admin", password); err != nil {
		t.Fatalf("bootstrap performance user: %v", err)
	}
	runner := collector.NewRunner(db, collector.NewRegistry(collector.FixtureAdapter{}), "performance-test")
	server := httptest.NewServer(httpapi.New(db, identityService, runner))
	t.Cleanup(server.Close)
	client := server.Client()
	admin := login(t, client, server.URL, "admin.performance@example.test", password)

	response := opportunityPage(t, client, server.URL+"/api/v1/opportunities?source_id="+sourceID+"&page_size=1000", admin.AccessToken)
	if response.PageSize != 100 || len(response.Items) != 100 {
		t.Fatalf("page size cap: page_size=%d items=%d", response.PageSize, len(response.Items))
	}

	durations := concurrentFeedLoad(t, client, server.URL, sourceID, admin.AccessToken)
	p95 := percentile(durations, 95)
	if p95 > performanceP95 {
		t.Fatalf("feed p95 %s exceeds budget %s (%d requests, %d workers)", p95, performanceP95, performanceRequests, performanceWorkers)
	}
	t.Logf("10k-row feed: requests=%d workers=%d p50=%s p95=%s max=%s", performanceRequests,
		performanceWorkers, percentile(durations, 50), p95, durations[len(durations)-1])
}

func seedPerformanceData(t *testing.T, ctx context.Context, db *pgxpool.Pool) string {
	t.Helper()
	var sourceID string
	err := db.QueryRow(ctx, `INSERT INTO sources(name,base_url,adapter_name,adapter_version,status,policy_version,policy)
		VALUES ('Performance Source','https://performance.example.test','spse-fixture','v1','active','performance-v1','{}')
		RETURNING id::text`).Scan(&sourceID)
	if err != nil {
		t.Fatalf("create performance source: %v", err)
	}

	_, err = db.Exec(ctx, `INSERT INTO opportunities(source_id,package_identity,title,agency,procurement_type,
		lifecycle_status,budget_minor,currency,registration_deadline,canonical_url,source_timestamp,current_fingerprint,
		first_seen_at,last_seen_at)
		SELECT $1,format('LOAD-%s',lpad(g::text,5,'0')),format('Load Package %s',lpad(g::text,5,'0')),
		'Performance Agency','Jasa Lainnya','registration_open',g::bigint*1000000,'IDR',
		now()+((g%365)+1)*interval '1 day',format('https://performance.example.test/package/%s',g),now(),
		format('performance-%s',g),now()-g*interval '1 second',now()-g*interval '1 second'
		FROM generate_series(1,$2) g`, sourceID, performanceRows)
	if err != nil {
		t.Fatalf("seed opportunities: %v", err)
	}

	_, err = db.Exec(ctx, `INSERT INTO opportunity_classifications(opportunity_id,opportunity_fingerprint,
		taxonomy_version,rule_version,disposition,category_code,confidence_bp,reason_code,explanation,auto_rejected)
		SELECT id,current_fingerprint,'cankonix-it-v1','deterministic-v1',
		CASE WHEN right(package_identity,1) IN ('0','5') THEN 'needs_review' ELSE 'relevant' END,
		CASE WHEN right(package_identity,1) IN ('0','5') THEN NULL ELSE 'data-ai' END,
		CASE WHEN right(package_identity,1) IN ('0','5') THEN 5000 ELSE 9000 END,
		CASE WHEN right(package_identity,1) IN ('0','5') THEN 'no_deterministic_match' ELSE 'include_data' END,
		'Performance fixture classification',false FROM opportunities WHERE source_id=$1`, sourceID)
	if err != nil {
		t.Fatalf("seed classifications: %v", err)
	}

	_, err = db.Exec(ctx, `INSERT INTO opportunity_scores(opportunity_id,opportunity_fingerprint,classification_id,
		scoring_version,total_score_bp,confidence_bp,needs_review,reason_code,explanation,evaluated_at)
		SELECT o.id,o.current_fingerprint,c.id,'explainable-v1',
		CASE WHEN c.disposition='needs_review' THEN 5600 ELSE 8000 END,c.confidence_bp,
		c.disposition='needs_review',CASE WHEN c.disposition='needs_review' THEN 'low_confidence' ELSE 'score_ready' END,
		'Performance fixture score',now() FROM opportunities o JOIN opportunity_classifications c ON c.opportunity_id=o.id
		WHERE o.source_id=$1`, sourceID)
	if err != nil {
		t.Fatalf("seed scores: %v", err)
	}
	if _, err := db.Exec(ctx, `ANALYZE opportunities; ANALYZE opportunity_classifications; ANALYZE opportunity_scores`); err != nil {
		t.Fatalf("analyze performance data: %v", err)
	}
	return sourceID
}

func concurrentFeedLoad(t *testing.T, client *http.Client, baseURL, sourceID, token string) []time.Duration {
	t.Helper()
	jobs := make(chan int)
	durations := make(chan time.Duration, performanceRequests)
	errors := make(chan error, performanceRequests)
	var workers sync.WaitGroup
	for worker := 0; worker < performanceWorkers; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for requestNumber := range jobs {
				page := requestNumber%100 + 1
				url := fmt.Sprintf("%s/api/v1/opportunities?source_id=%s&page=%d&page_size=100", baseURL, sourceID, page)
				started := time.Now()
				resp, err := client.Do(authenticatedRequest(t, url, token))
				elapsed := time.Since(started)
				if err != nil {
					errors <- err
					continue
				}
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					errors <- fmt.Errorf("page %d returned HTTP %d", page, resp.StatusCode)
					continue
				}
				durations <- elapsed
			}
		}()
	}
	for requestNumber := 0; requestNumber < performanceRequests; requestNumber++ {
		jobs <- requestNumber
	}
	close(jobs)
	workers.Wait()
	close(durations)
	close(errors)
	for err := range errors {
		t.Error(err)
	}
	if t.Failed() {
		t.FailNow()
	}
	result := make([]time.Duration, 0, performanceRequests)
	for duration := range durations {
		result = append(result, duration)
	}
	if len(result) != performanceRequests {
		t.Fatalf("completed requests=%d, want %d", len(result), performanceRequests)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func authenticatedRequest(t *testing.T, url, token string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func percentile(sorted []time.Duration, percentile int) time.Duration {
	index := (len(sorted)*percentile + 99) / 100
	if index < 1 {
		index = 1
	}
	return sorted[index-1]
}
