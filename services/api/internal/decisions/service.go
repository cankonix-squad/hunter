package decisions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Decision struct {
	OpportunityID string    `json:"opportunity_id"`
	Decision      string    `json:"decision"`
	Reason        string    `json:"reason"`
	DecidedBy     string    `json:"decided_by"`
	DeciderName   string    `json:"decider_name"`
	DecidedAt     time.Time `json:"decided_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
type UpdateInput struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}
type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) Get(ctx context.Context, opportunityID string) (*Decision, error) {
	if !validUUID(opportunityID) {
		return nil, pgx.ErrNoRows
	}
	var exists bool
	if err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM opportunities WHERE id=$1)`, opportunityID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, pgx.ErrNoRows
	}
	var item Decision
	err := s.db.QueryRow(ctx, `SELECT d.opportunity_id::text,d.decision::text,d.reason,d.decided_by::text,u.display_name,d.decided_at,d.updated_at
		FROM opportunity_decisions d JOIN users u ON u.id=d.decided_by WHERE d.opportunity_id=$1`, opportunityID).
		Scan(&item.OpportunityID, &item.Decision, &item.Reason, &item.DecidedBy, &item.DeciderName, &item.DecidedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get decision: %w", err)
	}
	return &item, nil
}

func (s *Service) Update(ctx context.Context, opportunityID, actorID string, input UpdateInput) (Decision, error) {
	input.Decision = strings.TrimSpace(input.Decision)
	input.Reason = strings.TrimSpace(input.Reason)
	if input.Decision != "pursue" && input.Decision != "pass" {
		return Decision{}, errors.New("decision must be pursue or pass")
	}
	if len([]rune(input.Reason)) < 3 || len([]rune(input.Reason)) > 2000 {
		return Decision{}, errors.New("reason must contain between 3 and 2000 characters")
	}
	if !validUUID(opportunityID) {
		return Decision{}, pgx.ErrNoRows
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Decision{}, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM opportunities WHERE id=$1)`, opportunityID).Scan(&exists); err != nil {
		return Decision{}, err
	}
	if !exists {
		return Decision{}, pgx.ErrNoRows
	}
	var previous string
	_ = tx.QueryRow(ctx, `SELECT decision::text FROM opportunity_decisions WHERE opportunity_id=$1`, opportunityID).Scan(&previous)
	_, err = tx.Exec(ctx, `INSERT INTO opportunity_decisions(opportunity_id,decision,reason,decided_by) VALUES($1,$2,$3,$4)
		ON CONFLICT(opportunity_id) DO UPDATE SET decision=excluded.decision,reason=excluded.reason,decided_by=excluded.decided_by,updated_at=now()`, opportunityID, input.Decision, input.Reason, actorID)
	if err != nil {
		return Decision{}, fmt.Errorf("update decision: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata) VALUES($1,'opportunity.decision_changed','opportunity',$2,
		jsonb_build_object('previous_decision',$3::text,'decision',$4::text,'reason',$5::text))`, actorID, opportunityID, previous, input.Decision, input.Reason)
	if err != nil {
		return Decision{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Decision{}, err
	}
	item, err := s.Get(ctx, opportunityID)
	if err != nil || item == nil {
		return Decision{}, err
	}
	return *item, nil
}
func validUUID(value string) bool { var id pgtype.UUID; return id.Scan(value) == nil && id.Valid }
