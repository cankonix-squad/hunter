package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleResult struct {
	Eligible int
	Queued   int
}

type Scheduler struct {
	db  *pgxpool.Pool
	now func() time.Time
}

func NewScheduler(db *pgxpool.Pool) *Scheduler {
	return &Scheduler{db: db, now: time.Now}
}

func (s *Scheduler) Tick(ctx context.Context) (ScheduleResult, error) {
	rows, err := s.db.Query(ctx, `SELECT s.id::text, s.policy_version,
		coalesce((s.policy->>'schedule_interval_minutes')::integer, 360)
		FROM sources s
		WHERE s.status = 'active'
		  AND coalesce((s.policy->>'schedule_enabled')::boolean, false)
		  AND coalesce((s.policy->>'schedule_interval_minutes')::integer, 360) BETWEEN 5 AND 10080
		  AND (s.circuit_open_until IS NULL OR s.circuit_open_until <= now())
		  AND NOT EXISTS (SELECT 1 FROM source_runs pending WHERE pending.source_id=s.id
		                  AND pending.status IN ('queued','running'))
		  AND coalesce((SELECT max(previous.created_at) FROM source_runs previous WHERE previous.source_id=s.id),
		               '-infinity'::timestamptz)
		      <= now() - make_interval(mins => coalesce((s.policy->>'schedule_interval_minutes')::integer, 360))`)
	if err != nil {
		return ScheduleResult{}, fmt.Errorf("list due sources: %w", err)
	}
	defer rows.Close()

	result := ScheduleResult{}
	for rows.Next() {
		var sourceID, policyVersion string
		var intervalMinutes int64
		if err := rows.Scan(&sourceID, &policyVersion, &intervalMinutes); err != nil {
			return result, fmt.Errorf("scan due source: %w", err)
		}
		result.Eligible++
		key := scheduleKey(sourceID, policyVersion, intervalMinutes, s.now().UTC())
		tag, err := s.db.Exec(ctx, `INSERT INTO source_runs(source_id,idempotency_key,policy_version)
			VALUES ($1,$2,$3) ON CONFLICT (idempotency_key) DO NOTHING`, sourceID, key, policyVersion)
		if err != nil {
			return result, fmt.Errorf("queue scheduled source: %w", err)
		}
		result.Queued += int(tag.RowsAffected())
	}
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("iterate due sources: %w", err)
	}
	return result, nil
}

func scheduleKey(sourceID, policyVersion string, intervalMinutes int64, now time.Time) string {
	seconds := intervalMinutes * 60
	bucket := now.Unix() / seconds
	return fmt.Sprintf("scheduled:%s:%s:%d", sourceID, policyVersion, bucket)
}
