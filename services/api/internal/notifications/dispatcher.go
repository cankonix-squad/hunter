package notifications

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dispatcher struct {
	db     *pgxpool.Pool
	sender EmailSender
	owner  string
}

func NewDispatcher(db *pgxpool.Pool, sender EmailSender, owner string) *Dispatcher {
	return &Dispatcher{db: db, sender: sender, owner: owner}
}

func (d *Dispatcher) RunOne(ctx context.Context) (bool, error) {
	if _, err := d.db.Exec(ctx, `UPDATE notification_deliveries SET status='failed',next_attempt_at=now(),lease_owner=NULL,lease_expires_at=NULL,safe_error='delivery lease expired',updated_at=now() WHERE status='running' AND lease_expires_at<now()`); err != nil {
		return false, fmt.Errorf("release expired delivery leases: %w", err)
	}
	if _, err := d.db.Exec(ctx, `INSERT INTO notification_deliveries(alert_event_id,channel,destination)
	SELECT e.id,'email',u.email FROM alert_events e JOIN users u ON u.id=e.user_id WHERE e.delivery_status='pending' AND e.available_after<=now() AND u.active ON CONFLICT DO NOTHING`); err != nil {
		return false, fmt.Errorf("enqueue email deliveries: %w", err)
	}
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	var id int64
	var to, title, eventType, canonicalURL string
	var attempts int
	err = tx.QueryRow(ctx, `WITH candidate AS(SELECT d.id FROM notification_deliveries d WHERE d.channel='email' AND d.status IN('queued','failed') AND d.next_attempt_at<=now() ORDER BY d.created_at FOR UPDATE SKIP LOCKED LIMIT 1)
	UPDATE notification_deliveries d SET status='running',attempts=attempts+1,lease_owner=$1,lease_expires_at=now()+interval '2 minutes',updated_at=now() FROM candidate,alert_events e,opportunities o WHERE d.id=candidate.id AND e.id=d.alert_event_id AND o.id=e.opportunity_id RETURNING d.id,d.destination,o.title,e.event_type,o.canonical_url,d.attempts`, d.owner).Scan(&id, &to, &title, &eventType, &canonicalURL, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return false, err
	}
	message := Message{To: to, Subject: "CANKONIX alert: " + title, Text: fmt.Sprintf("%s\n\nEvent: %s\nOfficial source: %s\n\nVerify procurement details on the official source before making a decision.", title, eventType, canonicalURL)}
	if err = d.sender.Send(message); err != nil {
		return true, d.fail(ctx, id, attempts, err)
	}
	return true, d.complete(ctx, id)
}

func (d *Dispatcher) complete(ctx context.Context, id int64) error {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var eventID int64
	err = tx.QueryRow(ctx, `UPDATE notification_deliveries SET status='delivered',delivered_at=now(),lease_owner=NULL,lease_expires_at=NULL,safe_error=NULL,updated_at=now() WHERE id=$1 RETURNING alert_event_id`, id).Scan(&eventID)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE alert_events SET delivery_status='delivered' WHERE id=$1`, eventID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (d *Dispatcher) fail(ctx context.Context, id int64, attempts int, sendErr error) error {
	status := "failed"
	if attempts >= 5 {
		status = "dead_letter"
	}
	backoff := time.Duration(math.Min(math.Pow(2, float64(attempts-1))*30, 3600)) * time.Second
	safe := sendErr.Error()
	if len(safe) > 500 {
		safe = safe[:500]
	}
	_, err := d.db.Exec(ctx, `UPDATE notification_deliveries SET status=$2,next_attempt_at=now()+$3::interval,lease_owner=NULL,lease_expires_at=NULL,safe_error=$4,updated_at=now() WHERE id=$1`, id, status, fmt.Sprintf("%f seconds", backoff.Seconds()), safe)
	return err
}
