package collector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/alerts"
	"github.com/cankonix/cankonix-oi/services/api/internal/classification"
	"github.com/cankonix/cankonix-oi/services/api/internal/opportunities"
	"github.com/cankonix/cankonix-oi/services/api/internal/scoring"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Run struct {
	ID       string `json:"id"`
	SourceID string `json:"source_id"`
	Status   string `json:"status"`
	Attempts int    `json:"attempts"`
}

type policy struct {
	MaxRetries              int `json:"max_retries"`
	CircuitFailureThreshold int `json:"circuit_failure_threshold"`
}

type Runner struct {
	db       *pgxpool.Pool
	registry *Registry
	owner    string
}

func NewRunner(db *pgxpool.Pool, registry *Registry, owner string) *Runner {
	return &Runner{db: db, registry: registry, owner: owner}
}

func (r *Runner) Queue(ctx context.Context, sourceID, idempotencyKey string) (Run, error) {
	var run Run
	err := r.db.QueryRow(ctx, `INSERT INTO source_runs(source_id, idempotency_key, policy_version)
		SELECT id, $2, policy_version FROM sources
		WHERE id = $1 AND status = 'active' AND (circuit_open_until IS NULL OR circuit_open_until <= now())
		ON CONFLICT (idempotency_key) DO UPDATE SET idempotency_key = excluded.idempotency_key
		RETURNING id::text, source_id::text, status::text, attempts`, sourceID, idempotencyKey).
		Scan(&run.ID, &run.SourceID, &run.Status, &run.Attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, errors.New("source is not active or its circuit is open")
	}
	if err != nil {
		return Run{}, fmt.Errorf("queue source run: %w", err)
	}
	return run, nil
}

func (r *Runner) ExecuteOne(ctx context.Context) (bool, error) {
	run, err := r.claim(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := r.execute(ctx, run); err != nil {
		slog.Warn("source run failed", "run_id", run.ID, "source_id", run.SourceID, "error", err)
		if failureErr := r.fail(ctx, run, err); failureErr != nil {
			return true, failureErr
		}
		return true, nil
	}
	return true, nil
}

func (r *Runner) claim(ctx context.Context) (Run, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Run{}, fmt.Errorf("begin run claim: %w", err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `UPDATE source_runs SET status = 'failed', lease_owner = NULL, lease_expires_at = NULL,
		next_attempt_at = now(), safe_error = 'worker lease expired', error_code = 'lease_expired'
		WHERE status = 'running' AND lease_expires_at < now()`)
	if err != nil {
		return Run{}, fmt.Errorf("release expired leases: %w", err)
	}
	var run Run
	err = tx.QueryRow(ctx, `WITH candidate AS (
		SELECT sr.id FROM source_runs sr JOIN sources s ON s.id = sr.source_id
		WHERE sr.status IN ('queued', 'failed') AND sr.next_attempt_at <= now()
		  AND s.status = 'active' AND (s.circuit_open_until IS NULL OR s.circuit_open_until <= now())
		ORDER BY sr.created_at FOR UPDATE OF sr SKIP LOCKED LIMIT 1
	)
	UPDATE source_runs sr SET status = 'running', attempts = attempts + 1, started_at = coalesce(started_at, now()),
		lease_owner = $1, lease_expires_at = now() + interval '2 minutes'
	FROM candidate WHERE sr.id = candidate.id
	RETURNING sr.id::text, sr.source_id::text, sr.status::text, sr.attempts`, r.owner).
		Scan(&run.ID, &run.SourceID, &run.Status, &run.Attempts)
	if err != nil {
		return Run{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Run{}, fmt.Errorf("commit run claim: %w", err)
	}
	return run, nil
}

func (r *Runner) execute(ctx context.Context, run Run) error {
	var source Source
	err := r.db.QueryRow(ctx, `SELECT id::text, name, base_url, adapter_name, adapter_version, policy
		FROM sources WHERE id = $1`, run.SourceID).Scan(&source.ID, &source.Name, &source.BaseURL,
		&source.AdapterName, &source.AdapterVersion, &source.Policy)
	if err != nil {
		return fmt.Errorf("load source: %w", err)
	}
	adapter, err := r.registry.Resolve(source.AdapterName, source.AdapterVersion)
	if err != nil {
		return err
	}
	candidates, err := adapter.Discover(ctx, source)
	if err != nil {
		return fmt.Errorf("discover: %w", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin snapshot write: %w", err)
	}
	defer tx.Rollback(ctx)
	inserted := 0
	for _, candidate := range candidates {
		normalized, err := opportunities.Normalize(candidate.Metadata)
		if err != nil {
			return fmt.Errorf("normalize package %s: %w", candidate.PackageIdentity, err)
		}
		fingerprint, err := candidate.Fingerprint()
		if err != nil {
			return err
		}
		metadata, err := json.Marshal(candidate.Metadata)
		if err != nil {
			return fmt.Errorf("encode safe metadata: %w", err)
		}
		var snapshotID string
		err = tx.QueryRow(ctx, `INSERT INTO source_snapshots(source_id, source_run_id, package_identity,
			fingerprint, canonical_url, source_timestamp, safe_metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (source_id, package_identity, fingerprint) DO UPDATE SET package_identity = excluded.package_identity
			RETURNING id::text`, source.ID, run.ID, candidate.PackageIdentity, fingerprint,
			candidate.CanonicalURL, candidate.SourceTimestamp, metadata).Scan(&snapshotID)
		if err != nil {
			return fmt.Errorf("insert snapshot: %w", err)
		}
		var previous opportunities.Normalized
		previousErr := tx.QueryRow(ctx, `SELECT title,agency,procurement_type,lifecycle_status,budget_minor,currency,registration_deadline
			FROM opportunities WHERE source_id=$1 AND package_identity=$2`, source.ID, candidate.PackageIdentity).
			Scan(&previous.Title, &previous.Agency, &previous.ProcurementType, &previous.LifecycleStatus,
				&previous.BudgetMinor, &previous.Currency, &previous.RegistrationDeadline)
		isNew := errors.Is(previousErr, pgx.ErrNoRows)
		if previousErr != nil && !isNew {
			return fmt.Errorf("load previous opportunity: %w", previousErr)
		}
		changed := opportunities.ChangedFields(nil, normalized)
		if !isNew {
			changed = opportunities.ChangedFields(&previous, normalized)
		}
		var opportunityID string
		err = tx.QueryRow(ctx, `INSERT INTO opportunities(source_id, package_identity, title, agency,
			procurement_type, lifecycle_status, budget_minor, currency, registration_deadline,
			canonical_url, source_timestamp, current_fingerprint)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (source_id, package_identity) DO UPDATE SET
				title = excluded.title, agency = excluded.agency, procurement_type = excluded.procurement_type,
				lifecycle_status = excluded.lifecycle_status, budget_minor = excluded.budget_minor,
				currency = excluded.currency, registration_deadline = excluded.registration_deadline,
				canonical_url = excluded.canonical_url, source_timestamp = excluded.source_timestamp,
				last_seen_at = now(), current_fingerprint = excluded.current_fingerprint, updated_at = now()
			RETURNING id::text`, source.ID, candidate.PackageIdentity, normalized.Title, normalized.Agency,
			normalized.ProcurementType, normalized.LifecycleStatus, normalized.BudgetMinor, normalized.Currency,
			normalized.RegistrationDeadline, candidate.CanonicalURL, candidate.SourceTimestamp, fingerprint).Scan(&opportunityID)
		if err != nil {
			return fmt.Errorf("upsert opportunity: %w", err)
		}
		normalizedJSON, err := json.Marshal(normalized)
		if err != nil {
			return fmt.Errorf("encode normalized opportunity: %w", err)
		}
		var versionID int64
		err = tx.QueryRow(ctx, `INSERT INTO opportunity_versions(opportunity_id, snapshot_id, fingerprint,
			lifecycle_status, normalized_data, observed_at)
			SELECT $1, $2, $3, $4, $5, observed_at FROM source_snapshots WHERE id = $2
			ON CONFLICT DO NOTHING RETURNING id`, opportunityID, snapshotID, fingerprint, normalized.LifecycleStatus, normalizedJSON).Scan(&versionID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("insert opportunity version: %w", err)
		}
		classificationResult := classification.DefaultEngine().Evaluate(classification.Input{
			Title: normalized.Title, ProcurementType: normalized.ProcurementType,
		})
		classificationID, err := persistClassification(ctx, tx, opportunityID, fingerprint, classificationResult)
		if err != nil {
			return err
		}
		evaluatedAt := time.Now().UTC()
		if candidate.SourceTimestamp != nil {
			evaluatedAt = candidate.SourceTimestamp.UTC()
		}
		score := scoring.Evaluate(scoring.Input{Classification: classificationResult,
			BudgetMinor: normalized.BudgetMinor, Currency: normalized.Currency,
			LifecycleStatus: normalized.LifecycleStatus, RegistrationDeadline: normalized.RegistrationDeadline,
			EvaluatedAt: evaluatedAt})
		if err := persistScore(ctx, tx, opportunityID, fingerprint, classificationID, score); err != nil {
			return err
		}
		if versionID != 0 {
			inserted++
			eventType := ""
			if isNew {
				eventType = "opportunity.new"
			} else if len(changed) > 0 {
				eventType = "opportunity.material_change"
			}
			if eventType != "" {
				if err := alerts.Match(ctx, tx, opportunityID, versionID, eventType, changed); err != nil {
					return err
				}
			}
		}
	}
	runMetrics := map[string]int{"candidates": len(candidates), "new_snapshots": inserted}
	if provider, ok := adapter.(MetricsProvider); ok {
		for key, value := range provider.DiscoveryMetrics() {
			runMetrics[key] = value
		}
	}
	metrics, _ := json.Marshal(runMetrics)
	_, err = tx.Exec(ctx, `UPDATE source_runs SET status = 'succeeded', finished_at = now(), lease_owner = NULL,
		lease_expires_at = NULL, metrics = $2, error_code = NULL, safe_error = NULL WHERE id = $1`, run.ID, metrics)
	if err != nil {
		return fmt.Errorf("complete source run: %w", err)
	}
	_, err = tx.Exec(ctx, `UPDATE sources SET consecutive_failures = 0, circuit_open_until = NULL,
		last_success_at = now(), health_score = least(100, health_score + 5), updated_at = now() WHERE id = $1`, source.ID)
	if err != nil {
		return fmt.Errorf("update source success: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit snapshots: %w", err)
	}
	slog.Info("source run succeeded", "run_id", run.ID, "candidates", len(candidates), "new_snapshots", inserted)
	return nil
}

func persistClassification(ctx context.Context, tx pgx.Tx, opportunityID, fingerprint string, result classification.Result) (int64, error) {
	var classificationID int64
	err := tx.QueryRow(ctx, `INSERT INTO opportunity_classifications(opportunity_id, opportunity_fingerprint,
		taxonomy_version, rule_version, disposition, category_code, confidence_bp, reason_code, explanation, auto_rejected)
		VALUES ($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9,$10)
		ON CONFLICT (opportunity_id, opportunity_fingerprint, taxonomy_version, rule_version) DO UPDATE SET
			disposition = excluded.disposition, category_code = excluded.category_code,
			confidence_bp = excluded.confidence_bp, reason_code = excluded.reason_code,
			explanation = excluded.explanation, auto_rejected = excluded.auto_rejected, evaluated_at = now()
		RETURNING id`, opportunityID, fingerprint, result.TaxonomyVersion, result.RuleVersion,
		result.Disposition, result.Category, result.ConfidenceBP, result.ReasonCode, result.Explanation,
		result.AutoRejected).Scan(&classificationID)
	if err != nil {
		return 0, fmt.Errorf("persist opportunity classification: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM classification_evidence WHERE classification_id = $1`, classificationID); err != nil {
		return 0, fmt.Errorf("replace classification evidence: %w", err)
	}
	for _, evidence := range result.Evidence {
		_, err := tx.Exec(ctx, `INSERT INTO classification_evidence(classification_id, rule_code, kind,
			category_code, keyword, source_field, excerpt) VALUES ($1,$2,$3,NULLIF($4,''),$5,$6,$7)`,
			classificationID, evidence.RuleCode, evidence.Kind, evidence.Category, evidence.Keyword,
			evidence.SourceField, evidence.Excerpt)
		if err != nil {
			return 0, fmt.Errorf("persist classification evidence: %w", err)
		}
	}
	return classificationID, nil
}

func persistScore(ctx context.Context, tx pgx.Tx, opportunityID, fingerprint string, classificationID int64, result scoring.Result) error {
	var scoreID int64
	err := tx.QueryRow(ctx, `INSERT INTO opportunity_scores(opportunity_id, opportunity_fingerprint,
		classification_id, scoring_version, total_score_bp, confidence_bp, needs_review,
		reason_code, explanation, evaluated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (opportunity_id, opportunity_fingerprint, scoring_version) DO UPDATE SET
			classification_id = excluded.classification_id, total_score_bp = excluded.total_score_bp,
			confidence_bp = excluded.confidence_bp, needs_review = excluded.needs_review,
			reason_code = excluded.reason_code, explanation = excluded.explanation,
			evaluated_at = excluded.evaluated_at
		RETURNING id`, opportunityID, fingerprint, classificationID, result.Version, result.TotalScoreBP,
		result.ConfidenceBP, result.NeedsReview, result.ReasonCode, result.Explanation, result.EvaluatedAt).Scan(&scoreID)
	if err != nil {
		return fmt.Errorf("persist opportunity score: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM score_dimensions WHERE score_id = $1`, scoreID); err != nil {
		return fmt.Errorf("replace score dimensions: %w", err)
	}
	for _, dimension := range result.Dimensions {
		_, err := tx.Exec(ctx, `INSERT INTO score_dimensions(score_id, dimension_code, weight_bp,
			raw_score_bp, contribution, reason_code, explanation, evidence)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, scoreID, dimension.Code, dimension.WeightBP,
			dimension.RawScoreBP, dimension.Contribution, dimension.ReasonCode,
			dimension.Explanation, dimension.Evidence)
		if err != nil {
			return fmt.Errorf("persist score dimension: %w", err)
		}
	}
	return nil
}

func (r *Runner) fail(ctx context.Context, run Run, runErr error) error {
	var raw json.RawMessage
	var failures int
	if err := r.db.QueryRow(ctx, `SELECT policy, consecutive_failures FROM sources WHERE id = $1`, run.SourceID).Scan(&raw, &failures); err != nil {
		return fmt.Errorf("load retry policy: %w", err)
	}
	p := policy{MaxRetries: 5, CircuitFailureThreshold: 5}
	_ = json.Unmarshal(raw, &p)
	if p.MaxRetries < 0 {
		p.MaxRetries = 0
	}
	if p.CircuitFailureThreshold < 1 {
		p.CircuitFailureThreshold = 5
	}
	status, errorCode := "failed", "collector_error"
	var permanent interface{ Permanent() bool }
	if errors.As(runErr, &permanent) && permanent.Permanent() {
		status, errorCode = "dead_letter", "source_policy_error"
	} else if run.Attempts > p.MaxRetries {
		status = "dead_letter"
	}
	backoff := retryBackoff(run.Attempts)
	safeError := runErr.Error()
	if len(safeError) > 500 {
		safeError = safeError[:500]
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin persist failure: %w", err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `UPDATE source_runs SET status = $2::run_status, next_attempt_at = now() + $3::interval,
		finished_at = CASE WHEN $2 = 'dead_letter' THEN now() ELSE NULL END, lease_owner = NULL,
		lease_expires_at = NULL, error_code = $4, safe_error = $5 WHERE id = $1`,
		run.ID, status, fmt.Sprintf("%f seconds", backoff.Seconds()), errorCode, safeError)
	if err != nil {
		return fmt.Errorf("persist run failure: %w", err)
	}
	_, err = tx.Exec(ctx, `UPDATE sources SET consecutive_failures = consecutive_failures + 1, last_failure_at = now(),
		health_score = greatest(0, health_score - 20),
		circuit_open_until = CASE WHEN consecutive_failures + 1 >= $1 THEN now() + interval '15 minutes' ELSE circuit_open_until END,
		updated_at = now() WHERE id = $2`, p.CircuitFailureThreshold, run.SourceID)
	if err != nil {
		return fmt.Errorf("persist source failure: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit run failure: %w", err)
	}
	return nil
}

func retryBackoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	return time.Duration(math.Min(math.Pow(2, float64(attempt-1))*30, 1800)) * time.Second
}
