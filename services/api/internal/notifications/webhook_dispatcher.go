package notifications

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookDispatcher struct {
	db                 *pgxpool.Pool
	sender             WebhookSender
	owner, destination string
}

func NewWebhookDispatcher(db *pgxpool.Pool, sender WebhookSender, owner, destination string) *WebhookDispatcher {
	return &WebhookDispatcher{db: db, sender: sender, owner: owner, destination: destination}
}
func (d *WebhookDispatcher) RunOne(ctx context.Context) (bool, error) {
	if _, err := d.db.Exec(ctx, `UPDATE notification_deliveries SET status='failed',next_attempt_at=now(),lease_owner=NULL,lease_expires_at=NULL,safe_error='delivery lease expired',updated_at=now() WHERE channel='webhook' AND status='running' AND lease_expires_at<now()`); err != nil {
		return false, err
	}
	if _, err := d.db.Exec(ctx, `INSERT INTO notification_deliveries(alert_event_id,channel,destination)SELECT id,'webhook',$1 FROM alert_events WHERE delivery_status IN('pending','delivered') AND available_after<=now() ON CONFLICT DO NOTHING`, d.destination); err != nil {
		return false, fmt.Errorf("enqueue webhook deliveries: %w", err)
	}
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	var deliveryID int64
	var attempts int
	var payload WebhookPayload
	err = tx.QueryRow(ctx, `WITH candidate AS(SELECT id FROM notification_deliveries WHERE channel='webhook' AND status IN('queued','failed') AND next_attempt_at<=now() ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1) UPDATE notification_deliveries d SET status='running',attempts=attempts+1,lease_owner=$1,lease_expires_at=now()+interval '2 minutes',updated_at=now() FROM candidate,alert_events e,opportunities o WHERE d.id=candidate.id AND e.id=d.alert_event_id AND o.id=e.opportunity_id RETURNING d.id,e.id,e.event_type,o.id::text,o.package_identity,o.title,e.changed_fields,o.canonical_url,e.created_at,d.attempts`, d.owner).Scan(&deliveryID, &payload.ID, &payload.EventType, &payload.OpportunityID, &payload.PackageIdentity, &payload.Title, &payload.ChangedFields, &payload.CanonicalURL, &payload.OccurredAt, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return false, err
	}
	code, sendErr := d.sender.Send(payload)
	if sendErr != nil {
		return true, failDelivery(ctx, d.db, deliveryID, attempts, sendErr, code)
	}
	return true, completeDelivery(ctx, d.db, deliveryID, code)
}
