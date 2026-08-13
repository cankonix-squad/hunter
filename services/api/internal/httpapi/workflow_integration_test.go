//go:build integration

package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/analytics"
	"github.com/cankonix/cankonix-oi/services/api/internal/collector"
	"github.com/cankonix/cankonix-oi/services/api/internal/httpapi"
	"github.com/cankonix/cankonix-oi/services/api/internal/identity"
	"github.com/cankonix/cankonix-oi/services/api/internal/platform/database"
)

func TestAuthenticatedOpportunityWorkflowEndToEnd(t *testing.T) {
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

	password := randomCredential(t)
	email := "workflow.operator@example.test"
	createUser(t, ctx, db, email, "Workflow Operator", password, "operator")
	var operatorID string
	if err := db.QueryRow(ctx, `SELECT id::text FROM users WHERE email=$1`, email).Scan(&operatorID); err != nil {
		t.Fatalf("load workflow operator: %v", err)
	}

	var opportunityID string
	err = db.QueryRow(ctx, `WITH source AS (
		INSERT INTO sources(name,base_url,adapter_name,adapter_version,status,policy_version,policy)
		VALUES ('Workflow Acceptance Source',$1,'spse-fixture','v1','active','workflow-v1','{}')
		RETURNING id
	) INSERT INTO opportunities(source_id,package_identity,title,agency,procurement_type,
		lifecycle_status,currency,canonical_url,current_fingerprint)
	SELECT id,'WORKFLOW-ACCEPTANCE-1','Workflow Acceptance Opportunity','CANKONIX Test',
		'Jasa Lainnya','registration_open','IDR',$2,'workflow-acceptance-fingerprint' FROM source
	RETURNING id::text`, "https://workflow.example.test", "https://example.test/workflow-acceptance").Scan(&opportunityID)
	if err != nil {
		t.Fatalf("seed workflow opportunity: %v", err)
	}

	identityService := identity.New(db, randomCredential(t)+randomCredential(t), 15*time.Minute, time.Hour)
	runner := collector.NewRunner(db, collector.NewRegistry(collector.FixtureAdapter{}), "workflow-test")
	server := httptest.NewServer(httpapi.New(db, identityService, runner))
	t.Cleanup(server.Close)
	client := server.Client()
	operator := login(t, client, server.URL, email, password)
	baselinePipeline := workflowPipeline(t, client, server.URL, operator.AccessToken)

	resp := request(t, client, http.MethodGet, server.URL+"/api/v1/opportunities?status=active&q=Workflow+Acceptance", operator.AccessToken, nil)
	assertStatus(t, resp, http.StatusOK)
	var feed struct {
		Total int `json:"total"`
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	decode(t, resp, &feed)
	resp.Body.Close()
	if feed.Total != 1 || len(feed.Items) != 1 || feed.Items[0].ID != opportunityID {
		t.Fatalf("workflow feed = %+v, want seeded opportunity %s", feed, opportunityID)
	}

	resp = request(t, client, http.MethodGet, server.URL+"/api/v1/opportunities/"+opportunityID, operator.AccessToken, nil)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	resp = request(t, client, http.MethodPatch, server.URL+"/api/v1/opportunities/"+opportunityID+"/assignment", operator.AccessToken, map[string]any{"assignee_user_id": operatorID})
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	resp = request(t, client, http.MethodPatch, server.URL+"/api/v1/opportunities/"+opportunityID+"/shortlist", operator.AccessToken, map[string]any{"shortlisted": true})
	assertStatus(t, resp, http.StatusOK)
	var shortlistState struct {
		Shortlisted bool `json:"shortlisted"`
	}
	decode(t, resp, &shortlistState)
	resp.Body.Close()
	if !shortlistState.Shortlisted {
		t.Fatal("opportunity was not shortlisted")
	}

	resp = request(t, client, http.MethodPatch, server.URL+"/api/v1/opportunities/"+opportunityID+"/decision", operator.AccessToken, map[string]any{
		"decision": "pursue", "reason": "Matches the acceptance scope",
	})
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	resp = request(t, client, http.MethodPost, server.URL+"/api/v1/opportunities/"+opportunityID+"/comments", operator.AccessToken, map[string]any{
		"body": "Validate delivery capacity",
	})
	assertStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	resp = request(t, client, http.MethodPatch, server.URL+"/api/v1/opportunities/"+opportunityID+"/readiness", operator.AccessToken, map[string]any{
		"code": "scope_validated", "completed": true, "note": "Scope reviewed",
	})
	assertStatus(t, resp, http.StatusOK)
	var checklist struct {
		Completed int `json:"completed"`
		Total     int `json:"total"`
	}
	decode(t, resp, &checklist)
	resp.Body.Close()
	if checklist.Completed != 1 || checklist.Total != 5 {
		t.Fatalf("readiness = %+v, want 1/5", checklist)
	}

	resp = request(t, client, http.MethodPatch, server.URL+"/api/v1/opportunities/"+opportunityID+"/override", operator.AccessToken, map[string]any{
		"recommendation": "relevant", "score_bp": nil, "reason": "Reviewed by operator", "clear": false,
	})
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	dueAt := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)
	resp = request(t, client, http.MethodPatch, server.URL+"/api/v1/opportunities/"+opportunityID+"/next-action", operator.AccessToken, map[string]any{
		"action": "Validate technical scope", "due_at": dueAt,
	})
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	mutatedPipeline := workflowPipeline(t, client, server.URL, operator.AccessToken)
	if mutatedPipeline.Total != baselinePipeline.Total || mutatedPipeline.Active != baselinePipeline.Active ||
		mutatedPipeline.Assigned != baselinePipeline.Assigned+1 || mutatedPipeline.Shortlisted != baselinePipeline.Shortlisted+1 ||
		mutatedPipeline.Pursue != baselinePipeline.Pursue+1 || mutatedPipeline.Undecided != baselinePipeline.Undecided-1 ||
		mutatedPipeline.ActionsDueSevenDays != baselinePipeline.ActionsDueSevenDays+1 {
		t.Fatalf("mutated pipeline = %+v, baseline = %+v", mutatedPipeline, baselinePipeline)
	}

	resp = request(t, client, http.MethodGet, server.URL+"/api/v1/opportunities/"+opportunityID+"/activity", operator.AccessToken, nil)
	assertStatus(t, resp, http.StatusOK)
	var activity struct {
		Items []struct {
			Action string `json:"action"`
		} `json:"items"`
	}
	decode(t, resp, &activity)
	resp.Body.Close()
	for _, expected := range []string{
		"opportunity.assigned",
		"opportunity.shortlisted",
		"opportunity.decision_changed",
		"opportunity.next_action_updated",
		"opportunity.comment_created",
		"opportunity.readiness_updated",
		"opportunity.override_updated",
	} {
		found := false
		for _, item := range activity.Items {
			if item.Action == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("activity missing %q: %+v", expected, activity.Items)
		}
	}

	resp = request(t, client, http.MethodPatch, server.URL+"/api/v1/opportunities/"+opportunityID+"/assignment", operator.AccessToken, map[string]any{"assignee_user_id": nil})
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	resp = request(t, client, http.MethodPatch, server.URL+"/api/v1/opportunities/"+opportunityID+"/shortlist", operator.AccessToken, map[string]any{"shortlisted": false})
	assertStatus(t, resp, http.StatusOK)
	decode(t, resp, &shortlistState)
	resp.Body.Close()
	if shortlistState.Shortlisted {
		t.Fatal("opportunity remained shortlisted after removal")
	}

	resp = request(t, client, http.MethodPatch, server.URL+"/api/v1/opportunities/"+opportunityID+"/next-action", operator.AccessToken, map[string]any{"action": nil, "due_at": nil})
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	resp = request(t, client, http.MethodPatch, server.URL+"/api/v1/opportunities/"+opportunityID+"/override", operator.AccessToken, map[string]any{
		"recommendation": nil, "score_bp": nil, "reason": "", "clear": true,
	})
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	for _, endpoint := range []string{"assignment", "decision", "next-action", "override"} {
		resp = request(t, client, http.MethodGet, server.URL+"/api/v1/opportunities/"+opportunityID+"/"+endpoint, operator.AccessToken, nil)
		assertStatus(t, resp, http.StatusOK)
		resp.Body.Close()
	}
	for table, expected := range map[string]int{
		"opportunity_assignments":     0,
		"opportunity_shortlist":       0,
		"opportunity_next_actions":    0,
		"opportunity_overrides":       0,
		"opportunity_decisions":       1,
		"opportunity_comments":        1,
		"opportunity_readiness_items": 1,
	} {
		var count int
		query := "SELECT count(*) FROM " + table + " WHERE opportunity_id=$1"
		if err := db.QueryRow(ctx, query, opportunityID).Scan(&count); err != nil {
			t.Fatalf("count workflow state in %s: %v", table, err)
		}
		if count != expected {
			t.Fatalf("workflow state %s count=%d, want %d", table, count, expected)
		}
	}
	reversedPipeline := workflowPipeline(t, client, server.URL, operator.AccessToken)
	if reversedPipeline.Total != baselinePipeline.Total || reversedPipeline.Active != baselinePipeline.Active ||
		reversedPipeline.Assigned != baselinePipeline.Assigned || reversedPipeline.Shortlisted != baselinePipeline.Shortlisted ||
		reversedPipeline.Pursue != baselinePipeline.Pursue+1 || reversedPipeline.Undecided != baselinePipeline.Undecided-1 ||
		reversedPipeline.ActionsDueSevenDays != baselinePipeline.ActionsDueSevenDays {
		t.Fatalf("reversed pipeline = %+v, baseline = %+v", reversedPipeline, baselinePipeline)
	}

	resp = request(t, client, http.MethodGet, server.URL+"/api/v1/opportunities/"+opportunityID+"/activity", operator.AccessToken, nil)
	assertStatus(t, resp, http.StatusOK)
	decode(t, resp, &activity)
	resp.Body.Close()
	for _, expected := range []string{
		"opportunity.unassigned",
		"opportunity.unshortlisted",
		"opportunity.next_action_cleared",
		"opportunity.override_cleared",
	} {
		found := false
		for _, item := range activity.Items {
			if item.Action == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("activity missing reversal %q: %+v", expected, activity.Items)
		}
	}
}

func workflowPipeline(t *testing.T, client *http.Client, baseURL, token string) analytics.Pipeline {
	t.Helper()
	resp := request(t, client, http.MethodGet, baseURL+"/api/v1/analytics/pipeline", token, nil)
	assertStatus(t, resp, http.StatusOK)
	defer resp.Body.Close()
	var pipeline analytics.Pipeline
	decode(t, resp, &pipeline)
	return pipeline
}
