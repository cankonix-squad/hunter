package notifications

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Delivery struct {
	ID              int64      `json:"id"`
	AlertEventID    int64      `json:"alert_event_id"`
	Channel         string     `json:"channel"`
	Destination     string     `json:"destination"`
	Status          string     `json:"status"`
	Attempts        int        `json:"attempts"`
	ReplayCount     int        `json:"replay_count"`
	SafeError       string     `json:"safe_error,omitempty"`
	ResponseCode    *int       `json:"response_code,omitempty"`
	EventType       string     `json:"event_type"`
	OpportunityID   string     `json:"opportunity_id"`
	PackageIdentity string     `json:"package_identity"`
	Title           string     `json:"title"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LastReplayedAt  *time.Time `json:"last_replayed_at,omitempty"`
}
type Service struct{ db *pgxpool.Pool }

func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }
func (s *Service) ListDeadLetters(ctx context.Context) ([]Delivery, error) {
	rows, err := s.db.Query(ctx, `SELECT d.id,d.alert_event_id,d.channel,d.destination,d.status,d.attempts,d.replay_count,coalesce(d.safe_error,''),d.response_code,e.event_type,o.id::text,o.package_identity,o.title,d.updated_at,d.last_replayed_at FROM notification_deliveries d JOIN alert_events e ON e.id=d.alert_event_id JOIN opportunities o ON o.id=e.opportunity_id WHERE d.status='dead_letter' ORDER BY d.updated_at DESC,d.id DESC LIMIT 200`)
	if err != nil {
		return nil, fmt.Errorf("list dead letters: %w", err)
	}
	defer rows.Close()
	items := make([]Delivery, 0)
	for rows.Next() {
		var x Delivery
		if err := rows.Scan(&x.ID, &x.AlertEventID, &x.Channel, &x.Destination, &x.Status, &x.Attempts, &x.ReplayCount, &x.SafeError, &x.ResponseCode, &x.EventType, &x.OpportunityID, &x.PackageIdentity, &x.Title, &x.UpdatedAt, &x.LastReplayedAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func (s *Service) Replay(ctx context.Context, id int64, actorID string) (Delivery, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Delivery{}, fmt.Errorf("reset dead-letter delivery: %w", err)
	}
	defer tx.Rollback(ctx)
	var x Delivery
	err = tx.QueryRow(ctx, `UPDATE notification_deliveries d SET status='queued',attempts=0,next_attempt_at=now(),lease_owner=NULL,lease_expires_at=NULL,safe_error=NULL,response_code=NULL,replay_count=replay_count+1,last_replayed_at=now(),last_replayed_by=$2,updated_at=now() FROM alert_events e,opportunities o WHERE d.id=$1 AND d.status='dead_letter' AND e.id=d.alert_event_id AND o.id=e.opportunity_id RETURNING d.id,d.alert_event_id,d.channel,d.destination,d.status,d.attempts,d.replay_count,coalesce(d.safe_error,''),d.response_code,e.event_type,o.id::text,o.package_identity,o.title,d.updated_at,d.last_replayed_at`, id, actorID).Scan(&x.ID, &x.AlertEventID, &x.Channel, &x.Destination, &x.Status, &x.Attempts, &x.ReplayCount, &x.SafeError, &x.ResponseCode, &x.EventType, &x.OpportunityID, &x.PackageIdentity, &x.Title, &x.UpdatedAt, &x.LastReplayedAt)
	if err != nil {
		return Delivery{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata)VALUES($1,'notification_delivery.replayed','notification_delivery',$2,jsonb_build_object('channel',$3::text,'replay_count',$4::integer))`, actorID, strconv.FormatInt(id, 10), x.Channel, x.ReplayCount); err != nil {
		return Delivery{}, fmt.Errorf("audit dead-letter replay: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Delivery{}, err
	}
	return x, nil
}
