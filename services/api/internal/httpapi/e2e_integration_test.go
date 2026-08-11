//go:build integration

package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/collector"
	"github.com/cankonix/cankonix-oi/services/api/internal/httpapi"
	"github.com/cankonix/cankonix-oi/services/api/internal/identity"
	"github.com/cankonix/cankonix-oi/services/api/internal/opportunities"
	"github.com/cankonix/cankonix-oi/services/api/internal/platform/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDiscoverNormalizeClassifyScoreReviewEndToEnd(t *testing.T) {
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
		t.Fatalf("open integration database: %v", err)
	}
	t.Cleanup(db.Close)

	var sourceID string
	err = db.QueryRow(ctx, `INSERT INTO sources(name,base_url,adapter_name,adapter_version,status,policy_version,policy)
		VALUES ('E2E Fixture Source','https://e2e-fixture.example.test','spse-fixture','v1','active','e2e-v1','{}')
		RETURNING id::text`).Scan(&sourceID)
	if err != nil {
		t.Fatalf("create E2E source: %v", err)
	}

	runner := collector.NewRunner(db, collector.NewRegistry(collector.FixtureAdapter{}), "e2e-test")
	firstRun, err := runner.Queue(ctx, sourceID, "e2e-discover-1")
	if err != nil {
		t.Fatalf("queue first discovery: %v", err)
	}
	processed, err := runner.ExecuteOne(ctx)
	if err != nil || !processed {
		t.Fatalf("execute first discovery: processed=%v error=%v", processed, err)
	}
	assertRun(t, ctx, db, firstRun.ID, "succeeded", 1, 2)

	assertCount(t, ctx, db, `SELECT count(*) FROM source_snapshots WHERE source_id=$1`, sourceID, 2)
	assertCount(t, ctx, db, `SELECT count(*) FROM opportunities WHERE source_id=$1`, sourceID, 2)
	assertCount(t, ctx, db, `SELECT count(*) FROM opportunity_versions v JOIN opportunities o ON o.id=v.opportunity_id WHERE o.source_id=$1`, sourceID, 2)
	assertCount(t, ctx, db, `SELECT count(*) FROM opportunity_classifications c JOIN opportunities o ON o.id=c.opportunity_id WHERE o.source_id=$1`, sourceID, 2)
	assertCount(t, ctx, db, `SELECT count(*) FROM opportunity_scores s JOIN opportunities o ON o.id=s.opportunity_id WHERE o.source_id=$1`, sourceID, 2)
	assertCount(t, ctx, db, `SELECT count(*) FROM score_dimensions d JOIN opportunity_scores s ON s.id=d.score_id JOIN opportunities o ON o.id=s.opportunity_id WHERE o.source_id=$1`, sourceID, 10)

	password := randomCredential(t)
	identityService := identity.New(db, randomCredential(t)+randomCredential(t), 15*time.Minute, time.Hour)
	if err := identityService.BootstrapAdmin(ctx, "admin.e2e@example.test", "E2E Admin", password); err != nil {
		t.Fatalf("bootstrap E2E administrator: %v", err)
	}
	server := httptest.NewServer(httpapi.New(db, identityService, runner))
	t.Cleanup(server.Close)
	client := server.Client()
	admin := login(t, client, server.URL, "admin.e2e@example.test", password)

	relevant := opportunityPage(t, client, server.URL+"/api/v1/opportunities?source_id="+sourceID+"&relevance=relevant&min_score=7000", admin.AccessToken)
	if relevant.Total != 1 || len(relevant.Items) != 1 {
		t.Fatalf("relevant feed total=%d items=%d, want 1", relevant.Total, len(relevant.Items))
	}
	relevantItem := relevant.Items[0]
	if relevantItem.PackageIdentity != "TEST-SPSE-2026-0001" || relevantItem.Classification == nil || relevantItem.Score == nil {
		t.Fatalf("unexpected relevant opportunity: %+v", relevantItem)
	}
	if relevantItem.Classification.Disposition != "relevant" || relevantItem.Classification.CategoryCode != "data-ai" {
		t.Fatalf("unexpected relevant classification: %+v", relevantItem.Classification)
	}
	if relevantItem.Score.TotalScoreBP != 8000 || relevantItem.Score.NeedsReview {
		t.Fatalf("unexpected relevant score: %+v", relevantItem.Score)
	}

	detail := opportunityDetail(t, client, server.URL, relevantItem.ID, admin.AccessToken)
	if detail.VersionCount != 1 || len(detail.History) != 1 {
		t.Fatalf("detail history count=%d history=%d", detail.VersionCount, len(detail.History))
	}
	if detail.Classification == nil || len(detail.Classification.Evidence) != 1 {
		t.Fatalf("classification evidence = %+v", detail.Classification)
	}
	if detail.Score == nil || len(detail.Score.Dimensions) != 5 {
		t.Fatalf("score dimensions = %+v", detail.Score)
	}

	review := opportunityPage(t, client, server.URL+"/api/v1/opportunities/review?source_id="+sourceID, admin.AccessToken)
	if review.Total != 1 || len(review.Items) != 1 {
		t.Fatalf("review total=%d items=%d, want 1", review.Total, len(review.Items))
	}
	if review.Items[0].PackageIdentity != "TEST-SPSE-2026-0002" || review.Items[0].Classification == nil || review.Items[0].Score == nil {
		t.Fatalf("unexpected review opportunity: %+v", review.Items[0])
	}
	if review.Items[0].Classification.Disposition != "needs_review" || !review.Items[0].Score.NeedsReview {
		t.Fatalf("ambiguous opportunity did not enter review: %+v", review.Items[0])
	}

	secondRun, err := runner.Queue(ctx, sourceID, "e2e-discover-2")
	if err != nil {
		t.Fatalf("queue repeated discovery: %v", err)
	}
	processed, err = runner.ExecuteOne(ctx)
	if err != nil || !processed {
		t.Fatalf("execute repeated discovery: processed=%v error=%v", processed, err)
	}
	assertRun(t, ctx, db, secondRun.ID, "succeeded", 1, 0)

	assertCount(t, ctx, db, `SELECT count(*) FROM source_snapshots WHERE source_id=$1`, sourceID, 2)
	assertCount(t, ctx, db, `SELECT count(*) FROM opportunity_versions v JOIN opportunities o ON o.id=v.opportunity_id WHERE o.source_id=$1`, sourceID, 2)
	assertCount(t, ctx, db, `SELECT count(*) FROM opportunity_classifications c JOIN opportunities o ON o.id=c.opportunity_id WHERE o.source_id=$1`, sourceID, 2)
	assertCount(t, ctx, db, `SELECT count(*) FROM opportunity_scores s JOIN opportunities o ON o.id=s.opportunity_id WHERE o.source_id=$1`, sourceID, 2)
	assertCount(t, ctx, db, `SELECT count(*) FROM score_dimensions d JOIN opportunity_scores s ON s.id=d.score_id JOIN opportunities o ON o.id=s.opportunity_id WHERE o.source_id=$1`, sourceID, 10)

	review = opportunityPage(t, client, server.URL+"/api/v1/opportunities/review?source_id="+sourceID, admin.AccessToken)
	if review.Total != 1 {
		t.Fatalf("review total after repeated discovery = %d, want 1", review.Total)
	}
}

func assertRun(t *testing.T, ctx context.Context, db *pgxpool.Pool, runID, expectedStatus string, expectedAttempts, expectedSnapshots int) {
	t.Helper()
	var status string
	var attempts, snapshots int
	if err := db.QueryRow(ctx, `SELECT status::text,attempts,coalesce((metrics->>'new_snapshots')::integer,0) FROM source_runs WHERE id=$1`, runID).
		Scan(&status, &attempts, &snapshots); err != nil {
		t.Fatalf("load source run: %v", err)
	}
	if status != expectedStatus || attempts != expectedAttempts || snapshots != expectedSnapshots {
		t.Fatalf("run status=%s attempts=%d snapshots=%d", status, attempts, snapshots)
	}
}

func assertCount(t *testing.T, ctx context.Context, db *pgxpool.Pool, query, sourceID string, expected int) {
	t.Helper()
	var count int
	if err := db.QueryRow(ctx, query, sourceID).Scan(&count); err != nil {
		t.Fatalf("count E2E records: %v", err)
	}
	if count != expected {
		t.Fatalf("count = %d, want %d for %s", count, expected, query)
	}
}

func opportunityPage(t *testing.T, client *http.Client, url, token string) opportunities.Page {
	t.Helper()
	resp := request(t, client, http.MethodGet, url, token, nil)
	assertStatus(t, resp, http.StatusOK)
	defer resp.Body.Close()
	var page opportunities.Page
	decode(t, resp, &page)
	return page
}

func opportunityDetail(t *testing.T, client *http.Client, baseURL, id, token string) opportunities.Item {
	t.Helper()
	resp := request(t, client, http.MethodGet, baseURL+"/api/v1/opportunities/"+id, token, nil)
	assertStatus(t, resp, http.StatusOK)
	defer resp.Body.Close()
	var item opportunities.Item
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		t.Fatalf("decode opportunity detail: %v", err)
	}
	return item
}
