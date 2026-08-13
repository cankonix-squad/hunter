//go:build integration

package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

	dueAt := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)
	resp = request(t, client, http.MethodPatch, server.URL+"/api/v1/opportunities/"+opportunityID+"/next-action", operator.AccessToken, map[string]any{
		"action": "Validate technical scope", "due_at": dueAt,
	})
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	resp = request(t, client, http.MethodGet, server.URL+"/api/v1/opportunities/"+opportunityID+"/activity", operator.AccessToken, nil)
	assertStatus(t, resp, http.StatusOK)
	var activity struct {
		Items []struct {
			Action string `json:"action"`
		} `json:"items"`
	}
	decode(t, resp, &activity)
	resp.Body.Close()
	for _, expected := range []string{"opportunity.shortlisted", "opportunity.decision_changed", "opportunity.next_action_updated"} {
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
}
