package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Event struct {
	ID         int64           `json:"id"`
	Action     string          `json:"action"`
	ActorID    *string         `json:"actor_id,omitempty"`
	ActorName  string          `json:"actor_name,omitempty"`
	Metadata   json.RawMessage `json:"metadata"`
	OccurredAt time.Time       `json:"occurred_at"`
}
type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }
func (s *Service) List(ctx context.Context, opportunityID string) ([]Event, error) {
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
	rows, err := s.db.Query(ctx, `SELECT a.id,a.action,a.actor_id::text,coalesce(u.display_name,''),a.metadata,a.occurred_at FROM audit_logs a LEFT JOIN users u ON u.id=a.actor_id WHERE a.object_type='opportunity' AND a.object_id=$1 ORDER BY a.occurred_at DESC,a.id DESC LIMIT 200`, opportunityID)
	if err != nil {
		return nil, fmt.Errorf("list activity: %w", err)
	}
	defer rows.Close()
	items := make([]Event, 0)
	for rows.Next() {
		var x Event
		if err := rows.Scan(&x.ID, &x.Action, &x.ActorID, &x.ActorName, &x.Metadata, &x.OccurredAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func validUUID(v string) bool { var id pgtype.UUID; return id.Scan(v) == nil && id.Valid }
