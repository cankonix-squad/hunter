package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/activity"
	"github.com/cankonix/cankonix-oi/services/api/internal/alerts"
	"github.com/cankonix/cankonix-oi/services/api/internal/analytics"
	"github.com/cankonix/cankonix-oi/services/api/internal/assignments"
	"github.com/cankonix/cankonix-oi/services/api/internal/collector"
	"github.com/cankonix/cankonix-oi/services/api/internal/comments"
	"github.com/cankonix/cankonix-oi/services/api/internal/decisions"
	"github.com/cankonix/cankonix-oi/services/api/internal/identity"
	"github.com/cankonix/cankonix-oi/services/api/internal/nextactions"
	"github.com/cankonix/cankonix-oi/services/api/internal/notifications"
	"github.com/cankonix/cankonix-oi/services/api/internal/opportunities"
	"github.com/cankonix/cankonix-oi/services/api/internal/overrides"
	"github.com/cankonix/cankonix-oi/services/api/internal/readiness"
	"github.com/cankonix/cankonix-oi/services/api/internal/savedviews"
	"github.com/cankonix/cankonix-oi/services/api/internal/shortlist"
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
	comments      *comments.Service
	shortlist     *shortlist.Service
	decisions     *decisions.Service
	nextActions   *nextactions.Service
	readiness     *readiness.Service
	overrides     *overrides.Service
	activity      *activity.Service
	analytics     *analytics.Service
	alerts        *alerts.Service
	notifications *notifications.Service
	startedAt     time.Time
}

func New(db *pgxpool.Pool, identityService *identity.Service, collectorRunner *collector.Runner) http.Handler {
	s := &Server{db: db, identity: identityService, sources: sources.New(db), collector: collectorRunner,
		opportunities: opportunities.New(db), savedViews: savedviews.New(db), assignments: assignments.New(db), comments: comments.New(db), shortlist: shortlist.New(db), decisions: decisions.New(db), nextActions: nextactions.New(db), readiness: readiness.New(db), overrides: overrides.New(db), activity: activity.New(db), analytics: analytics.New(db), alerts: alerts.New(db), notifications: notifications.NewService(db), startedAt: time.Now().UTC()}
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
	mux.Handle("GET /api/v1/opportunities/{id}/comments", s.authenticate(http.HandlerFunc(s.listComments)))
	mux.Handle("POST /api/v1/opportunities/{id}/comments", s.authenticate(http.HandlerFunc(s.createComment)))
	mux.Handle("GET /api/v1/opportunities/{id}/shortlist", s.authenticate(http.HandlerFunc(s.getShortlist)))
	mux.Handle("PATCH /api/v1/opportunities/{id}/shortlist", s.authenticate(http.HandlerFunc(s.updateShortlist)))
	mux.Handle("GET /api/v1/opportunities/{id}/decision", s.authenticate(http.HandlerFunc(s.getDecision)))
	mux.Handle("PATCH /api/v1/opportunities/{id}/decision", s.authenticate(requireRoles("administrator", "operator", "business_development", "bid_manager", "executive")(http.HandlerFunc(s.updateDecision))))
	mux.Handle("GET /api/v1/opportunities/{id}/next-action", s.authenticate(http.HandlerFunc(s.getNextAction)))
	mux.Handle("PATCH /api/v1/opportunities/{id}/next-action", s.authenticate(requireRoles("administrator", "operator", "business_development", "bid_manager")(http.HandlerFunc(s.updateNextAction))))
	mux.Handle("GET /api/v1/opportunities/{id}/readiness", s.authenticate(http.HandlerFunc(s.getReadiness)))
	mux.Handle("PATCH /api/v1/opportunities/{id}/readiness", s.authenticate(requireRoles("administrator", "operator", "business_development", "bid_manager", "technical_lead")(http.HandlerFunc(s.updateReadiness))))
	mux.Handle("GET /api/v1/opportunities/{id}/override", s.authenticate(http.HandlerFunc(s.getOverride)))
	mux.Handle("PATCH /api/v1/opportunities/{id}/override", s.authenticate(requireRoles("administrator", "operator", "business_development", "bid_manager", "executive")(http.HandlerFunc(s.updateOverride))))
	mux.Handle("GET /api/v1/opportunities/{id}/activity", s.authenticate(http.HandlerFunc(s.listActivity)))
	mux.Handle("GET /api/v1/analytics/pipeline", s.authenticate(http.HandlerFunc(s.pipelineAnalytics)))
	mux.Handle("GET /api/v1/alert-rules", s.authenticate(http.HandlerFunc(s.listAlertRules)))
	mux.Handle("POST /api/v1/alert-rules", s.authenticate(http.HandlerFunc(s.createAlertRule)))
	mux.Handle("PUT /api/v1/alert-rules/{id}", s.authenticate(http.HandlerFunc(s.updateAlertRule)))
	mux.Handle("DELETE /api/v1/alert-rules/{id}", s.authenticate(http.HandlerFunc(s.deleteAlertRule)))
	mux.Handle("GET /api/v1/alerts", s.authenticate(http.HandlerFunc(s.listAlerts)))
	mux.Handle("GET /api/v1/notification-preferences", s.authenticate(http.HandlerFunc(s.getNotificationPreferences)))
	mux.Handle("PUT /api/v1/notification-preferences", s.authenticate(http.HandlerFunc(s.updateNotificationPreferences)))
	mux.Handle("GET /api/v1/operations/dead-letters", s.authenticate(requireRoles("administrator", "operator")(http.HandlerFunc(s.listDeadLetters))))
	mux.Handle("POST /api/v1/operations/dead-letters/{id}/replay", s.authenticate(requireRoles("administrator", "operator")(http.HandlerFunc(s.replayDeadLetter))))
	mux.Handle("POST /api/v1/admin/sources", s.authenticate(requireRoles("administrator")(http.HandlerFunc(s.createSource))))
	mux.Handle("PATCH /api/v1/admin/sources/{id}", s.authenticate(requireRoles("administrator")(http.HandlerFunc(s.updateSource))))
	mux.Handle("POST /api/v1/admin/source-runs", s.authenticate(requireRoles("administrator", "operator")(http.HandlerFunc(s.queueSourceRun))))
	return requestID(recoverer(securityHeaders(mux)))
}

func (s *Server) listAlertRules(w http.ResponseWriter, r *http.Request) {
	p, _ := principalFrom(r)
	items, err := s.alerts.List(r.Context(), p.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Alert rules are temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (s *Server) createAlertRule(w http.ResponseWriter, r *http.Request) {
	var input alerts.Input
	if decodeJSON(w, r, &input) != nil {
		return
	}
	p, _ := principalFrom(r)
	item, err := s.alerts.Create(r.Context(), p.UserID, input)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_alert_rule", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}
func (s *Server) updateAlertRule(w http.ResponseWriter, r *http.Request) {
	var input alerts.Input
	if decodeJSON(w, r, &input) != nil {
		return
	}
	p, _ := principalFrom(r)
	item, err := s.alerts.Update(r.Context(), p.UserID, r.PathValue("id"), input)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "alert_rule_not_found", "Alert rule was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_alert_rule", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}
func (s *Server) deleteAlertRule(w http.ResponseWriter, r *http.Request) {
	p, _ := principalFrom(r)
	err := s.alerts.Delete(r.Context(), p.UserID, r.PathValue("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "alert_rule_not_found", "Alert rule was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Alert rule could not be deleted")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (s *Server) listAlerts(w http.ResponseWriter, r *http.Request) {
	p, _ := principalFrom(r)
	items, err := s.alerts.ListEvents(r.Context(), p.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Alerts are temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (s *Server) getNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	p, _ := principalFrom(r)
	item, err := s.alerts.GetPreferences(r.Context(), p.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Notification preferences are temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, item)
}
func (s *Server) updateNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	var input alerts.PreferencesInput
	if decodeJSON(w, r, &input) != nil {
		return
	}
	p, _ := principalFrom(r)
	item, err := s.alerts.UpdatePreferences(r.Context(), p.UserID, input)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_notification_preferences", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}
func (s *Server) listDeadLetters(w http.ResponseWriter, r *http.Request) {
	items, err := s.notifications.ListDeadLetters(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Dead letters are temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (s *Server) replayDeadLetter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusNotFound, "delivery_not_found", "Dead-letter delivery was not found")
		return
	}
	p, _ := principalFrom(r)
	item, err := s.notifications.Replay(r.Context(), id, p.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "delivery_not_found", "Dead-letter delivery was not found")
		return
	}
	if err != nil {
		slog.Error("dead-letter replay failed", "delivery_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Dead-letter delivery could not be replayed")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) pipelineAnalytics(w http.ResponseWriter, r *http.Request) {
	item, err := s.analytics.Pipeline(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Pipeline analytics are temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) listActivity(w http.ResponseWriter, r *http.Request) {
	items, err := s.activity.List(r.Context(), r.PathValue("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Activity is temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) getOverride(w http.ResponseWriter, r *http.Request) {
	item, err := s.overrides.Get(r.Context(), r.PathValue("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Override is temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"override": item})
}
func (s *Server) updateOverride(w http.ResponseWriter, r *http.Request) {
	var input overrides.UpdateInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	p, _ := principalFrom(r)
	item, err := s.overrides.Update(r.Context(), r.PathValue("id"), p.UserID, input)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_override", "Override requires a valid recommendation or score and a mandatory reason")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"override": item})
}

func (s *Server) getReadiness(w http.ResponseWriter, r *http.Request) {
	item, err := s.readiness.Get(r.Context(), r.PathValue("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Readiness is temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, item)
}
func (s *Server) updateReadiness(w http.ResponseWriter, r *http.Request) {
	var input readiness.UpdateInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	p, _ := principalFrom(r)
	item, err := s.readiness.Update(r.Context(), r.PathValue("id"), p.UserID, input)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_readiness", "Readiness item or note is invalid")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) getNextAction(w http.ResponseWriter, r *http.Request) {
	item, err := s.nextActions.Get(r.Context(), r.PathValue("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Next action is temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"next_action": item})
}
func (s *Server) updateNextAction(w http.ResponseWriter, r *http.Request) {
	var input nextactions.UpdateInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	p, _ := principalFrom(r)
	item, err := s.nextActions.Update(r.Context(), r.PathValue("id"), p.UserID, input)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_next_action", "Action and RFC3339 due_at are required together; send both null to clear")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"next_action": item})
}

func (s *Server) getDecision(w http.ResponseWriter, r *http.Request) {
	item, err := s.decisions.Get(r.Context(), r.PathValue("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Decision is temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"decision": item})
}
func (s *Server) updateDecision(w http.ResponseWriter, r *http.Request) {
	var input decisions.UpdateInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	p, _ := principalFrom(r)
	item, err := s.decisions.Update(r.Context(), r.PathValue("id"), p.UserID, input)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_decision", "Decision must be pursue or pass with a reason of 3 to 2000 characters")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"decision": item})
}

func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	items, err := s.comments.List(r.Context(), r.PathValue("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Comments are temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) createComment(w http.ResponseWriter, r *http.Request) {
	var input comments.CreateInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	p, _ := principalFrom(r)
	item, err := s.comments.Create(r.Context(), r.PathValue("id"), p.UserID, input)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_comment", "Comment must contain between 1 and 4000 characters")
		return
	}
	writeJSON(w, http.StatusCreated, item)
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
		NeedsReview: r.URL.Query().Get("needs_review"), Shortlisted: r.URL.Query().Get("shortlisted"),
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

func (s *Server) getShortlist(w http.ResponseWriter, r *http.Request) {
	item, err := s.shortlist.Get(r.Context(), r.PathValue("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Shortlist is temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, item)
}
func (s *Server) updateShortlist(w http.ResponseWriter, r *http.Request) {
	var input shortlist.UpdateInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	p, _ := principalFrom(r)
	item, err := s.shortlist.Update(r.Context(), r.PathValue("id"), p.UserID, input)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "opportunity_not_found", "Opportunity was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Shortlist could not be updated")
		return
	}
	writeJSON(w, http.StatusOK, item)
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
