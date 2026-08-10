package shortlist

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type State struct {
	OpportunityID string     `json:"opportunity_id"`
	Shortlisted   bool       `json:"shortlisted"`
	ShortlistedBy *string    `json:"shortlisted_by,omitempty"`
	ActorName     string     `json:"actor_name,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}
type UpdateInput struct {
	Shortlisted bool `json:"shortlisted"`
}
type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) Get(ctx context.Context, opportunityID string) (State, error) {
	if !validUUID(opportunityID) {
		return State{}, pgx.ErrNoRows
	}
	var state State
	err := s.db.QueryRow(ctx, `SELECT o.id::text,sl.shortlisted_by::text,coalesce(u.display_name,''),sl.updated_at
		FROM opportunities o LEFT JOIN opportunity_shortlist sl ON sl.opportunity_id=o.id
		LEFT JOIN users u ON u.id=sl.shortlisted_by WHERE o.id=$1`, opportunityID).
		Scan(&state.OpportunityID, &state.ShortlistedBy, &state.ActorName, &state.UpdatedAt)
	if err != nil {
		return State{}, err
	}
	state.Shortlisted = state.ShortlistedBy != nil
	return state, nil
}

func (s *Service) Update(ctx context.Context, opportunityID, actorID string, input UpdateInput) (State, error) {
	if !validUUID(opportunityID) {
		return State{}, pgx.ErrNoRows
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return State{}, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM opportunities WHERE id=$1)`, opportunityID).Scan(&exists); err != nil {
		return State{}, err
	}
	if !exists {
		return State{}, pgx.ErrNoRows
	}
	action := "opportunity.shortlisted"
	if input.Shortlisted {
		_, err = tx.Exec(ctx, `INSERT INTO opportunity_shortlist(opportunity_id,shortlisted_by) VALUES($1,$2)
			ON CONFLICT(opportunity_id) DO UPDATE SET shortlisted_by=excluded.shortlisted_by,updated_at=now()`, opportunityID, actorID)
	} else {
		action = "opportunity.unshortlisted"
		_, err = tx.Exec(ctx, `DELETE FROM opportunity_shortlist WHERE opportunity_id=$1`, opportunityID)
	}
	if err != nil {
		return State{}, fmt.Errorf("update shortlist: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id) VALUES($1,$2,'opportunity',$3)`, actorID, action, opportunityID)
	if err != nil {
		return State{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return State{}, err
	}
	return s.Get(ctx, opportunityID)
}

func validUUID(value string) bool { var id pgtype.UUID; return id.Scan(value) == nil && id.Valid }
