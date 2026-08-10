package assignments

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Assignment struct {
	OpportunityID string    `json:"opportunity_id"`
	AssigneeID    string    `json:"assignee_user_id"`
	AssigneeName  string    `json:"assignee_name"`
	AssigneeEmail string    `json:"assignee_email"`
	AssignedBy    string    `json:"assigned_by"`
	AssignedAt    time.Time `json:"assigned_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UpdateInput struct {
	AssigneeUserID *string `json:"assignee_user_id"`
}
type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) Get(ctx context.Context, opportunityID string) (*Assignment, error) {
	var item Assignment
	err := s.db.QueryRow(ctx, `SELECT a.opportunity_id::text,a.assignee_user_id::text,u.display_name,u.email,
		a.assigned_by::text,a.assigned_at,a.updated_at FROM opportunity_assignments a
		JOIN users u ON u.id=a.assignee_user_id WHERE a.opportunity_id=$1`, opportunityID).
		Scan(&item.OpportunityID, &item.AssigneeID, &item.AssigneeName, &item.AssigneeEmail,
			&item.AssignedBy, &item.AssignedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get assignment: %w", err)
	}
	return &item, nil
}

func (s *Service) Update(ctx context.Context, opportunityID, actorID string, input UpdateInput) (*Assignment, error) {
	if !validUUID(opportunityID) {
		return nil, pgx.ErrNoRows
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM opportunities WHERE id=$1)`, opportunityID).Scan(&exists); err != nil || !exists {
		if err != nil {
			return nil, err
		}
		return nil, pgx.ErrNoRows
	}
	if input.AssigneeUserID == nil {
		tag, err := tx.Exec(ctx, `DELETE FROM opportunity_assignments WHERE opportunity_id=$1`, opportunityID)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() > 0 {
			_, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id) VALUES($1,'opportunity.unassigned','opportunity',$2)`, actorID, opportunityID)
			if err != nil {
				return nil, err
			}
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if !validUUID(*input.AssigneeUserID) {
		return nil, errors.New("assignee_user_id must be a UUID or null")
	}
	var active bool
	if err := tx.QueryRow(ctx, `SELECT active FROM users WHERE id=$1`, *input.AssigneeUserID).Scan(&active); errors.Is(err, pgx.ErrNoRows) || !active {
		return nil, errors.New("assignee must be an active user")
	} else if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO opportunity_assignments(opportunity_id,assignee_user_id,assigned_by)
		VALUES($1,$2,$3) ON CONFLICT(opportunity_id) DO UPDATE SET assignee_user_id=excluded.assignee_user_id,
		assigned_by=excluded.assigned_by,updated_at=now()`, opportunityID, *input.AssigneeUserID, actorID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata)
		VALUES($1,'opportunity.assigned','opportunity',$2,jsonb_build_object('assignee_user_id',$3::text))`, actorID, opportunityID, *input.AssigneeUserID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.Get(ctx, opportunityID)
}

func validUUID(value string) bool { var id pgtype.UUID; return id.Scan(value) == nil && id.Valid }
