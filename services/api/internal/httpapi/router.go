package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/assignments"
	"github.com/cankonix/cankonix-oi/services/api/internal/collector"
	"github.com/cankonix/cankonix-oi/services/api/internal/identity"
	"github.com/cankonix/cankonix-oi/services/api/internal/opportunities"
	"github.com/cankonix/cankonix-oi/services/api/internal/savedviews"
	"github.com/cankonix/cankonix-oi/services/api/internal/sources"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	db            *pgxpool.Pool
	identity      *identity.Service
	sources       *sources.Service
	collector     *collector.Runner
	opportunities *opportunities.Service
	savedViews    *savedviews.Service
	assignments   *assignments.Service
	startedAt     time.Time
}

func New(db *pgxpool.Pool, identityService *identity.Service, collectorRunner *collector.Runner) http.Handler {
	s := &Server{db: db, identity: identityService, sources: sources.New(db), collector: collectorRunner,
		opportunities: opportunities.New(db), savedViews: savedviews.New(db), assignments: assignments.New(db), startedAt: time.Now().UTC()}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /readyz", s.ready)
	mux.HandleFunc("GET /api/v1/system", s.system)
	mux.HandleFunc("POST /api/v1/auth/login", s.login)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", s.logout)
	mux.Handle("GET /api/v1/sources", s.authenticate(http.HandlerFunc(s.listSources)))
	mux.Handle("GET /api/v1/opportunities", s.authenticate(http.HandlerFunc(s.listOpportunities)))
	mux.Handle("GET /api/v1/opportunities/review", s.authenticate(http.HandlerFunc(s.listReviewOpportunities)))
	mux.Handle("GET /api/v1/opportunities/{id}", s.authenticate(http.HandlerFunc(s.getOpportunity)))
	mux.Handle("GET /api/v1/saved-views", s.authenticate(http.HandlerFunc(s.listSavedViews)))
	mux.Handle("POST /api/v1/saved-views", s.authenticate(http.HandlerFunc(s.createSavedView)))
	mux.Handle("DELETE /api/v1/saved-views/{id}", s.authenticate(http.HandlerFunc(s.deleteSavedView)))
	mux.Handle("GET /api/v1/team", s.authenticate(http.HandlerFunc(s.listTeam)))
	mux.Handle("GET /api/v1/opportunities/{id}/assignment", s.authenticate(http.HandlerFunc(s.getAssignment)))
	mux.Handle("PATCH /api/v1/opportunities/{id}/assignment", s.authenticate(requireRoles("administrator", "operator", "business_development", "bid_manager")(http.HandlerFunc(s.updateAssignment))))
	mux.Handle("POST /api/v1/admin/sources", s.authenticate(requireRoles("administrator")(http.HandlerFunc(s.createSource))))
	mux.Handle("PATCH /api/v1/admin/sources/{id}", s.authenticate(requireRoles("administrator")(http.HandlerFunc(s.updateSource))))
	mux.Handle("POST /api/v1/admin/source-runs", s.authenticate(requireRoles("administrator", "operator")(http.HandlerFunc(s.queueSourceRun))))
	return requestID(recoverer(securityHeaders(mux)))
}

func (s *Server) listTeam(w http.ResponseWriter, r *http.Request) {
	items, err := s.identity.ListTeam(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Team is temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) getAssignment(w http.ResponseWriter, r *http.Request) {
	item, err := s.assignments.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Assignment is temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"assignment": item})
}

func (s *Server) updateAssignment(w http.ResponseWriter, r *http.Request) {
	var input assignments.UpdateInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	p, _ := principalFrom(r)
	item, err := s.assignments.Update(r.Context(), r.PathValue("id"), p.UserID, input)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_assignment", "Assignee must be an active team member")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"assignment": item})
}

func (s *Server) listSavedViews(w http.ResponseWriter, r *http.Request) {
	p, _ := principalFrom(r)
	items, err := s.savedViews.List(r.Context(), p.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Saved views are temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) createSavedView(w http.ResponseWriter, r *http.Request) {
	var input savedviews.CreateInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	p, _ := principalFrom(r)
	item, err := s.savedViews.Create(r.Context(), p.UserID, input)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_saved_view", "Saved view filters are invalid or its name is already in use")
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) deleteSavedView(w http.ResponseWriter, r *http.Request) {
	p, _ := principalFrom(r)
	err := s.savedViews.Delete(r.Context(), p.UserID, r.PathValue("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "saved_view_not_found", "Saved view was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Saved view could not be deleted")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listOpportunities(w http.ResponseWriter, r *http.Request) {
	page, err := positiveInt(r.URL.Query().Get("page"), 1)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_page", "page must be a positive integer")
		return
	}
	pageSize, err := positiveInt(r.URL.Query().Get("page_size"), 20)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_page_size", "page_size must be a positive integer")
		return
	}
	minimumScore, err := boundedInt(r.URL.Query().Get("min_score"), -1, 0, 10000)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_min_score", "min_score must be between 0 and 10000")
		return
	}
	result, err := s.opportunities.List(r.Context(), opportunities.Query{
		Search: r.URL.Query().Get("q"), Status: r.URL.Query().Get("status"),
		SourceID: r.URL.Query().Get("source_id"), Page: page, PageSize: pageSize,
		Relevance: r.URL.Query().Get("relevance"), MinimumScoreBP: minimumScore,
		NeedsReview: r.URL.Query().Get("needs_review"),
	})
	if err != nil {
		if strings.HasPrefix(err.Error(), "invalid") {
			writeError(w, http.StatusBadRequest, "invalid_filter", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Opportunities are temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) listReviewOpportunities(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	query.Set("needs_review", "true")
	r.URL.RawQuery = query.Encode()
	s.listOpportunities(w, r)
}

func (s *Server) getOpportunity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var parsedID pgtype.UUID
	if err := parsedID.Scan(id); err != nil || !parsedID.Valid {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	item, err := s.opportunities.Get(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Opportunity is temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func positiveInt(raw string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, errors.New("not a positive integer")
	}
	return value, nil
}

func boundedInt(raw string, fallback, minimum, maximum int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < minimum || value > maximum {
		return 0, errors.New("out of range")
	}
	return value, nil
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 2*time.Second)
	defer cancel()
	if err := s.db.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not_ready", "dependency": "postgres"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
}

func (s *Server) system(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":       "CANKONIX Opportunity Intelligence",
		"version":    "0.1.0",
		"started_at": s.startedAt,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON with known fields")
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must contain one JSON value")
		return errors.New("multiple JSON values")
	}
	return nil
}

func withPrincipal(ctx context.Context, p identity.Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, p)
}

func principalFrom(r *http.Request) (identity.Principal, bool) {
	p, ok := r.Context().Value(principalContextKey{}).(identity.Principal)
	return p, ok
}

func (s *Server) listSources(w http.ResponseWriter, r *http.Request) {
	items, err := s.sources.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Sources are temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) createSource(w http.ResponseWriter, r *http.Request) {
	var body sources.CreateInput
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	p, _ := principalFrom(r)
	item, err := s.sources.Create(r.Context(), p.UserID, body)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_source", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) updateSource(w http.ResponseWriter, r *http.Request) {
	var body sources.UpdateInput
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	p, _ := principalFrom(r)
	item, err := s.sources.Update(r.Context(), p.UserID, r.PathValue("id"), body)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_source", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}

type queueRunRequest struct {
	SourceID       string `json:"source_id"`
	IdempotencyKey string `json:"idempotency_key"`
}

func (s *Server) queueSourceRun(w http.ResponseWriter, r *http.Request) {
	var body queueRunRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	if strings.TrimSpace(body.SourceID) == "" || strings.TrimSpace(body.IdempotencyKey) == "" {
		writeError(w, http.StatusUnprocessableEntity, "invalid_source_run", "source_id and idempotency_key are required")
		return
	}
	run, err := s.collector.Queue(r.Context(), body.SourceID, body.IdempotencyKey)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_source_run", err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, run)
}
