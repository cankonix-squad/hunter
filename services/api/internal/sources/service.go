package sources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Source struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	BaseURL             string          `json:"base_url"`
	Timezone            string          `json:"timezone"`
	AdapterName         string          `json:"adapter_name"`
	AdapterVersion      string          `json:"adapter_version"`
	Status              string          `json:"status"`
	Policy              json.RawMessage `json:"policy"`
	HealthScore         int16           `json:"health_score"`
	PolicyVersion       string          `json:"policy_version"`
	ConsecutiveFailures int             `json:"consecutive_failures"`
	CircuitOpenUntil    *time.Time      `json:"circuit_open_until,omitempty"`
	LastSuccessAt       *time.Time      `json:"last_success_at,omitempty"`
	LastFailureAt       *time.Time      `json:"last_failure_at,omitempty"`
	LatestRunStatus     string          `json:"latest_run_status,omitempty"`
	LatestRunMetrics    json.RawMessage `json:"latest_run_metrics"`
	LatestRunError      string          `json:"latest_run_error,omitempty"`
	LatestRunAt         *time.Time      `json:"latest_run_at,omitempty"`
	LagSeconds          *int64          `json:"lag_seconds,omitempty"`
	PendingRuns         int             `json:"pending_runs"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type CreateInput struct {
	Name           string          `json:"name"`
	BaseURL        string          `json:"base_url"`
	Timezone       string          `json:"timezone"`
	AdapterName    string          `json:"adapter_name"`
	AdapterVersion string          `json:"adapter_version"`
	Policy         json.RawMessage `json:"policy"`
}

type UpdateInput struct {
	Status        string          `json:"status"`
	PolicyVersion string          `json:"policy_version"`
	Policy        json.RawMessage `json:"policy"`
}

type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) List(ctx context.Context) ([]Source, error) {
	rows, err := s.db.Query(ctx, `SELECT s.id::text, s.name, s.base_url, s.timezone, s.adapter_name, s.adapter_version,
		s.status::text, s.policy, s.health_score, s.policy_version, s.consecutive_failures,
		s.circuit_open_until, s.last_success_at, s.last_failure_at,
		coalesce(latest.status::text,''), coalesce(latest.metrics,'{}'::jsonb), coalesce(latest.safe_error,''),
		latest.created_at,
		CASE WHEN s.last_success_at IS NULL THEN NULL ELSE extract(epoch FROM now()-s.last_success_at)::bigint END,
		(SELECT count(*)::integer FROM source_runs pending WHERE pending.source_id=s.id AND pending.status IN ('queued','running')),
		s.created_at, s.updated_at
		FROM sources s LEFT JOIN LATERAL (SELECT status,metrics,safe_error,created_at FROM source_runs
		WHERE source_id=s.id ORDER BY created_at DESC LIMIT 1) latest ON true ORDER BY s.name`)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	defer rows.Close()
	items := make([]Source, 0)
	for rows.Next() {
		var item Source
		if err := rows.Scan(&item.ID, &item.Name, &item.BaseURL, &item.Timezone, &item.AdapterName,
			&item.AdapterVersion, &item.Status, &item.Policy, &item.HealthScore, &item.PolicyVersion,
			&item.ConsecutiveFailures, &item.CircuitOpenUntil, &item.LastSuccessAt, &item.LastFailureAt,
			&item.LatestRunStatus, &item.LatestRunMetrics, &item.LatestRunError, &item.LatestRunAt,
			&item.LagSeconds, &item.PendingRuns, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Create(ctx context.Context, actorID string, in CreateInput) (Source, error) {
	if err := validate(&in); err != nil {
		return Source{}, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Source{}, fmt.Errorf("begin create source: %w", err)
	}
	defer tx.Rollback(ctx)
	var item Source
	err = tx.QueryRow(ctx, `INSERT INTO sources(name, base_url, timezone, adapter_name, adapter_version, policy)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text, name, base_url, timezone, adapter_name, adapter_version, status::text,
		policy, health_score, created_at, updated_at`, in.Name, in.BaseURL, in.Timezone, in.AdapterName, in.AdapterVersion, in.Policy).
		Scan(&item.ID, &item.Name, &item.BaseURL, &item.Timezone, &item.AdapterName, &item.AdapterVersion,
			&item.Status, &item.Policy, &item.HealthScore, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Source{}, errors.New("source was not created")
		}
		return Source{}, fmt.Errorf("insert source: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id, action, object_type, object_id, metadata)
		VALUES ($1, 'source.created', 'source', $2, jsonb_build_object('name', $3::text, 'base_url', $4::text))`, actorID, item.ID, item.Name, item.BaseURL); err != nil {
		return Source{}, fmt.Errorf("audit source creation: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Source{}, fmt.Errorf("commit source: %w", err)
	}
	return item, nil
}

func (s *Service) Update(ctx context.Context, actorID, sourceID string, in UpdateInput) (Source, error) {
	if in.Status != "active" && in.Status != "paused" && in.Status != "degraded" && in.Status != "disabled" {
		return Source{}, errors.New("status must be active, paused, degraded, or disabled")
	}
	if strings.TrimSpace(in.PolicyVersion) == "" {
		return Source{}, errors.New("policy_version is required")
	}
	var object map[string]any
	if json.Unmarshal(in.Policy, &object) != nil || object == nil {
		return Source{}, errors.New("policy must be a JSON object")
	}
	if err := validateSchedulePolicy(object); err != nil {
		return Source{}, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Source{}, fmt.Errorf("begin update source: %w", err)
	}
	defer tx.Rollback(ctx)
	var item Source
	err = tx.QueryRow(ctx, `UPDATE sources SET status = $2::source_status, policy_version = $3,
		policy = $4, updated_at = now() WHERE id = $1
		RETURNING id::text, name, base_url, timezone, adapter_name, adapter_version, status::text,
		policy, health_score, created_at, updated_at`, sourceID, in.Status, in.PolicyVersion, in.Policy).
		Scan(&item.ID, &item.Name, &item.BaseURL, &item.Timezone, &item.AdapterName, &item.AdapterVersion,
			&item.Status, &item.Policy, &item.HealthScore, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Source{}, errors.New("source not found")
	}
	if err != nil {
		return Source{}, fmt.Errorf("update source: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id, action, object_type, object_id, metadata)
		VALUES ($1, 'source.updated', 'source', $2, jsonb_build_object('status', $3::text, 'policy_version', $4::text))`,
		actorID, item.ID, in.Status, in.PolicyVersion)
	if err != nil {
		return Source{}, fmt.Errorf("audit source update: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Source{}, fmt.Errorf("commit source update: %w", err)
	}
	return item, nil
}

func validate(in *CreateInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.BaseURL = strings.TrimRight(strings.TrimSpace(in.BaseURL), "/")
	in.Timezone = strings.TrimSpace(in.Timezone)
	in.AdapterName = strings.TrimSpace(in.AdapterName)
	in.AdapterVersion = strings.TrimSpace(in.AdapterVersion)
	if in.Name == "" || in.BaseURL == "" || in.AdapterName == "" || in.AdapterVersion == "" {
		return errors.New("name, base_url, adapter_name, and adapter_version are required")
	}
	parsed, err := url.ParseRequestURI(in.BaseURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return errors.New("base_url must be an absolute HTTPS URL")
	}
	if in.Timezone == "" {
		in.Timezone = "Asia/Jakarta"
	}
	if _, err := time.LoadLocation(in.Timezone); err != nil {
		return errors.New("timezone is invalid")
	}
	if len(in.Policy) == 0 {
		in.Policy = json.RawMessage(`{}`)
	}
	var object map[string]any
	if json.Unmarshal(in.Policy, &object) != nil || object == nil {
		return errors.New("policy must be a JSON object")
	}
	if err := validateSchedulePolicy(object); err != nil {
		return err
	}
	return nil
}

func validateSchedulePolicy(policy map[string]any) error {
	if enabled, exists := policy["schedule_enabled"]; exists {
		if _, ok := enabled.(bool); !ok {
			return errors.New("policy.schedule_enabled must be a boolean")
		}
	}
	if interval, exists := policy["schedule_interval_minutes"]; exists {
		minutes, ok := interval.(float64)
		if !ok || minutes != float64(int64(minutes)) || minutes < 5 || minutes > 10080 {
			return errors.New("policy.schedule_interval_minutes must be an integer between 5 and 10080")
		}
	}
	return nil
}
