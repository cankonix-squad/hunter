//go:build integration

package httpapi_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/collector"
	"github.com/cankonix/cankonix-oi/services/api/internal/httpapi"
	"github.com/cankonix/cankonix-oi/services/api/internal/identity"
	"github.com/cankonix/cankonix-oi/services/api/internal/platform/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	User         struct {
		Email string   `json:"email"`
		Roles []string `json:"roles"`
	} `json:"user"`
}

func TestAuthorizationAndAPIContracts(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL or DATABASE_URL is required for integration tests")
	}

	ctx := context.Background()
	password := randomCredential(t)
	jwtSecret := randomCredential(t) + randomCredential(t)
	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open integration database: %v", err)
	}
	t.Cleanup(db.Close)

	identityService := identity.New(db, jwtSecret, 15*time.Minute, time.Hour)
	if err := identityService.BootstrapAdmin(ctx, "admin.contract@example.test", "Contract Admin", password); err != nil {
		t.Fatalf("bootstrap administrator: %v", err)
	}
	createUser(t, ctx, db, "operator.contract@example.test", "Contract Operator", password, "operator")
	createUser(t, ctx, db, "viewer.contract@example.test", "Contract Viewer", password, "")

	runner := collector.NewRunner(db, collector.NewRegistry(collector.FixtureAdapter{}), "contract-test")
	server := httptest.NewServer(httpapi.New(db, identityService, runner))
	t.Cleanup(server.Close)

	client := server.Client()

	t.Run("public health contract and security headers", func(t *testing.T) {
		resp := request(t, client, http.MethodGet, server.URL+"/healthz", "", nil)
		assertStatus(t, resp, http.StatusOK)
		defer resp.Body.Close()
		if resp.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("content type = %q", resp.Header.Get("Content-Type"))
		}
		if resp.Header.Get("X-Content-Type-Options") != "nosniff" || resp.Header.Get("X-Frame-Options") != "DENY" {
			t.Fatal("security headers are incomplete")
		}
		if resp.Header.Get("Cache-Control") != "no-store" || resp.Header.Get("X-Request-ID") == "" {
			t.Fatal("cache or request ID contract is incomplete")
		}
		var body map[string]string
		decode(t, resp, &body)
		if body["status"] != "ok" {
			t.Fatalf("health status = %q", body["status"])
		}
	})

	t.Run("protected endpoint rejects missing and invalid tokens", func(t *testing.T) {
		resp := request(t, client, http.MethodGet, server.URL+"/api/v1/sources", "", nil)
		assertAPIError(t, resp, http.StatusUnauthorized, "authentication_required")

		resp = request(t, client, http.MethodGet, server.URL+"/api/v1/sources", "not-a-token", nil)
		assertAPIError(t, resp, http.StatusUnauthorized, "invalid_access_token")
	})

	admin := login(t, client, server.URL, "admin.contract@example.test", password)
	operator := login(t, client, server.URL, "operator.contract@example.test", password)
	viewer := login(t, client, server.URL, "viewer.contract@example.test", password)

	t.Run("login response follows token contract", func(t *testing.T) {
		if admin.TokenType != "Bearer" || admin.AccessToken == "" || admin.RefreshToken == "" {
			t.Fatal("login response is missing bearer token fields")
		}
		if admin.User.Email != "admin.contract@example.test" || !contains(admin.User.Roles, "administrator") {
			t.Fatalf("unexpected administrator principal: %+v", admin.User)
		}
	})

	t.Run("authenticated read is allowed for roleless user", func(t *testing.T) {
		resp := request(t, client, http.MethodGet, server.URL+"/api/v1/sources", viewer.AccessToken, nil)
		assertStatus(t, resp, http.StatusOK)
		defer resp.Body.Close()
		var body struct {
			Items []map[string]any `json:"items"`
		}
		decode(t, resp, &body)
		if body.Items == nil {
			t.Fatal("sources contract must return an items array")
		}
	})

	t.Run("RBAC separates viewer operator and administrator", func(t *testing.T) {
		createBody := map[string]any{
			"name": "Contract Source", "base_url": "https://contract-source.example.test",
			"timezone": "Asia/Jakarta", "adapter_name": "fixture", "adapter_version": "v1", "policy": map[string]any{},
		}
		resp := request(t, client, http.MethodPost, server.URL+"/api/v1/admin/sources", viewer.AccessToken, createBody)
		assertAPIError(t, resp, http.StatusForbidden, "forbidden")

		resp = request(t, client, http.MethodPost, server.URL+"/api/v1/admin/sources", operator.AccessToken, createBody)
		assertAPIError(t, resp, http.StatusForbidden, "forbidden")

		resp = request(t, client, http.MethodPost, server.URL+"/api/v1/admin/sources", admin.AccessToken, createBody)
		assertStatus(t, resp, http.StatusCreated)
		var source struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		decode(t, resp, &source)
		resp.Body.Close()
		if source.ID == "" || source.Status != "disabled" {
			t.Fatalf("unexpected source contract: %+v", source)
		}

		updateBody := map[string]any{"status": "active", "policy_version": "contract-v1", "policy": map[string]any{}}
		resp = request(t, client, http.MethodPatch, server.URL+"/api/v1/admin/sources/"+source.ID, admin.AccessToken, updateBody)
		assertStatus(t, resp, http.StatusOK)
		resp.Body.Close()

		queueBody := map[string]any{"source_id": source.ID, "idempotency_key": "contract-source-run-1"}
		resp = request(t, client, http.MethodPost, server.URL+"/api/v1/admin/source-runs", viewer.AccessToken, queueBody)
		assertAPIError(t, resp, http.StatusForbidden, "forbidden")

		resp = request(t, client, http.MethodPost, server.URL+"/api/v1/admin/source-runs", operator.AccessToken, queueBody)
		assertStatus(t, resp, http.StatusAccepted)
		var run struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		decode(t, resp, &run)
		resp.Body.Close()
		if run.ID == "" || run.Status != "queued" {
			t.Fatalf("unexpected source-run contract: %+v", run)
		}

		resp = request(t, client, http.MethodGet, server.URL+"/api/v1/operations/dead-letters", viewer.AccessToken, nil)
		assertAPIError(t, resp, http.StatusForbidden, "forbidden")
		resp = request(t, client, http.MethodGet, server.URL+"/api/v1/operations/dead-letters", operator.AccessToken, nil)
		assertStatus(t, resp, http.StatusOK)
		var deadLetters struct {
			Items []map[string]any `json:"items"`
		}
		decode(t, resp, &deadLetters)
		resp.Body.Close()
		if deadLetters.Items == nil {
			t.Fatal("dead-letter contract must return an items array")
		}
	})

	t.Run("validation errors use safe stable envelope", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/alert-rules", strings.NewReader(`{"unknown":true}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+viewer.AccessToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "contract-request-id")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.Header.Get("X-Request-ID") != "contract-request-id" {
			t.Fatal("request ID was not preserved")
		}
		assertAPIError(t, resp, http.StatusBadRequest, "invalid_json")

		resp = request(t, client, http.MethodGet, server.URL+"/api/v1/opportunities?page=0", viewer.AccessToken, nil)
		assertAPIError(t, resp, http.StatusBadRequest, "invalid_page")
	})
}

func createUser(t *testing.T, ctx context.Context, db *pgxpool.Pool, email, name, password, role string) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(ctx, `WITH created AS (
		INSERT INTO users(email,display_name,password_hash) VALUES ($1,$2,$3) RETURNING id
	) INSERT INTO user_roles(user_id,role_id)
	SELECT created.id,roles.id FROM created JOIN roles ON roles.name=nullif($4,'')`, email, name, string(hash), role)
	if err != nil {
		t.Fatalf("create %s user: %v", role, err)
	}
}

func login(t *testing.T, client *http.Client, baseURL, email, password string) loginResponse {
	t.Helper()
	resp := request(t, client, http.MethodPost, baseURL+"/api/v1/auth/login", "", map[string]string{
		"email": email, "password": password,
	})
	assertStatus(t, resp, http.StatusOK)
	defer resp.Body.Close()
	var body loginResponse
	decode(t, resp, &body)
	return body
}

func randomCredential(t *testing.T) string {
	t.Helper()
	value := make([]byte, 24)
	if _, err := rand.Read(value); err != nil {
		t.Fatalf("generate test credential: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(value)
}

func request(t *testing.T, client *http.Client, method, url, token string, body any) *http.Response {
	t.Helper()
	var payload *bytes.Reader
	if body == nil {
		payload = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		payload = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		defer resp.Body.Close()
		var body any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		t.Fatalf("status = %d, want %d, body = %#v", resp.StatusCode, expected, body)
	}
}

func assertAPIError(t *testing.T, resp *http.Response, status int, code string) {
	t.Helper()
	assertStatus(t, resp, status)
	defer resp.Body.Close()
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("error content type = %q", resp.Header.Get("Content-Type"))
	}
	var body apiError
	decode(t, resp, &body)
	if body.Error.Code != code || body.Error.Message == "" {
		t.Fatalf("error = %+v, want code %q and safe message", body.Error, code)
	}
}

func decode(t *testing.T, resp *http.Response, dst any) {
	t.Helper()
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func contains(items []string, expected string) bool {
	for _, item := range items {
		if item == expected {
			return true
		}
	}
	return false
}
