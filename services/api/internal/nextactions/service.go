package nextactions

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

type NextAction struct {
	OpportunityID string    `json:"opportunity_id"`
	Action        string    `json:"action"`
	DueAt         time.Time `json:"due_at"`
	UpdatedBy     string    `json:"updated_by"`
	UpdaterName   string    `json:"updater_name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
type UpdateInput struct {
	Action *string    `json:"action"`
	DueAt  *time.Time `json:"due_at"`
}
type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) Get(ctx context.Context, opportunityID string) (*NextAction, error) {
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
	var item NextAction
	err := s.db.QueryRow(ctx, `SELECT n.opportunity_id::text,n.action,n.due_at,n.updated_by::text,u.display_name,n.created_at,n.updated_at
		FROM opportunity_next_actions n JOIN users u ON u.id=n.updated_by WHERE n.opportunity_id=$1`, opportunityID).
		Scan(&item.OpportunityID, &item.Action, &item.DueAt, &item.UpdatedBy, &item.UpdaterName, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get next action: %w", err)
	}
	return &item, nil
}

func (s *Service) Update(ctx context.Context, opportunityID, actorID string, input UpdateInput) (*NextAction, error) {
	if !validUUID(opportunityID) {
		return nil, pgx.ErrNoRows
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM opportunities WHERE id=$1)`, opportunityID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, pgx.ErrNoRows
	}
	if input.Action == nil && input.DueAt == nil {
		_, err = tx.Exec(ctx, `DELETE FROM opportunity_next_actions WHERE opportunity_id=$1`, opportunityID)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id) VALUES($1,'opportunity.next_action_cleared','opportunity',$2)`, actorID, opportunityID)
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if input.Action == nil || input.DueAt == nil {
		return nil, errors.New("action and due_at are required together")
	}
	action := strings.TrimSpace(*input.Action)
	if len([]rune(action)) < 3 || len([]rune(action)) > 500 {
		return nil, errors.New("action must contain between 3 and 500 characters")
	}
	due := input.DueAt.UTC()
	_, err = tx.Exec(ctx, `INSERT INTO opportunity_next_actions(opportunity_id,action,due_at,updated_by) VALUES($1,$2,$3,$4)
		ON CONFLICT(opportunity_id) DO UPDATE SET action=excluded.action,due_at=excluded.due_at,updated_by=excluded.updated_by,updated_at=now()`, opportunityID, action, due, actorID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata) VALUES($1,'opportunity.next_action_updated','opportunity',$2,jsonb_build_object('action',$3::text,'due_at',$4::timestamptz))`, actorID, opportunityID, action, due)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.Get(ctx, opportunityID)
}
func validUUID(value string) bool { var id pgtype.UUID; return id.Scan(value) == nil && id.Valid }
