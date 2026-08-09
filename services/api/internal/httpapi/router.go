package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/collector"
	"github.com/cankonix/cankonix-oi/services/api/internal/identity"
	"github.com/cankonix/cankonix-oi/services/api/internal/sources"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	db        *pgxpool.Pool
	identity  *identity.Service
	sources   *sources.Service
	collector *collector.Runner
	startedAt time.Time
}

func New(db *pgxpool.Pool, identityService *identity.Service, collectorRunner *collector.Runner) http.Handler {
	s := &Server{db: db, identity: identityService, sources: sources.New(db), collector: collectorRunner, startedAt: time.Now().UTC()}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /readyz", s.ready)
	mux.HandleFunc("GET /api/v1/system", s.system)
	mux.HandleFunc("POST /api/v1/auth/login", s.login)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", s.logout)
	mux.Handle("GET /api/v1/sources", s.authenticate(http.HandlerFunc(s.listSources)))
	mux.Handle("POST /api/v1/admin/sources", s.authenticate(requireRoles("administrator")(http.HandlerFunc(s.createSource))))
	mux.Handle("PATCH /api/v1/admin/sources/{id}", s.authenticate(requireRoles("administrator")(http.HandlerFunc(s.updateSource))))
	mux.Handle("POST /api/v1/admin/source-runs", s.authenticate(requireRoles("administrator", "operator")(http.HandlerFunc(s.queueSourceRun))))
	return requestID(recoverer(securityHeaders(mux)))
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
