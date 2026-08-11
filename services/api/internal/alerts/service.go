package alerts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Filters struct {
	SourceID       string `json:"source_id,omitempty"`
	Relevance      string `json:"relevance"`
	MinimumScoreBP *int   `json:"min_score_bp,omitempty"`
}

type Rule struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Enabled    bool      `json:"enabled"`
	EventTypes []string  `json:"event_types"`
	Filters    Filters   `json:"filters"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Input struct {
	Name       string   `json:"name"`
	Enabled    *bool    `json:"enabled,omitempty"`
	EventTypes []string `json:"event_types"`
	Filters    Filters  `json:"filters"`
}

type Event struct {
	ID              int64     `json:"id"`
	RuleID          string    `json:"rule_id"`
	RuleName        string    `json:"rule_name"`
	OpportunityID   string    `json:"opportunity_id"`
	PackageIdentity string    `json:"package_identity"`
	Title           string    `json:"title"`
	EventType       string    `json:"event_type"`
	ChangedFields   []string  `json:"changed_fields"`
	DeliveryStatus  string    `json:"delivery_status"`
	CreatedAt       time.Time `json:"created_at"`
	AvailableAfter  time.Time `json:"available_after"`
}

type Preferences struct {
	QuietHoursEnabled bool      `json:"quiet_hours_enabled"`
	QuietHoursStart   string    `json:"quiet_hours_start"`
	QuietHoursEnd     string    `json:"quiet_hours_end"`
	Timezone          string    `json:"timezone"`
	UpdatedAt         time.Time `json:"updated_at"`
}
type PreferencesInput struct {
	QuietHoursEnabled bool   `json:"quiet_hours_enabled"`
	QuietHoursStart   string `json:"quiet_hours_start"`
	QuietHoursEnd     string `json:"quiet_hours_end"`
	Timezone          string `json:"timezone"`
}

type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) List(ctx context.Context, userID string) ([]Rule, error) {
	rows, err := s.db.Query(ctx, `SELECT id::text,name,enabled,event_types,filters,created_at,updated_at FROM alert_rules WHERE user_id=$1 ORDER BY lower(name)`, userID)
	if err != nil {
		return nil, fmt.Errorf("list alert rules: %w", err)
	}
	defer rows.Close()
	items := make([]Rule, 0)
	for rows.Next() {
		item, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Create(ctx context.Context, userID string, input Input) (Rule, error) {
	if err := validate(&input); err != nil {
		return Rule{}, err
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	raw, _ := json.Marshal(input.Filters)
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Rule{}, err
	}
	defer tx.Rollback(ctx)
	var item Rule
	item.Filters = input.Filters
	err = tx.QueryRow(ctx, `INSERT INTO alert_rules(user_id,name,enabled,event_types,filters) VALUES($1,$2,$3,$4,$5) RETURNING id::text,name,enabled,event_types,created_at,updated_at`, userID, input.Name, enabled, input.EventTypes, raw).Scan(&item.ID, &item.Name, &item.Enabled, &item.EventTypes, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return Rule{}, fmt.Errorf("create alert rule: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata) VALUES($1,'alert_rule.created','alert_rule',$2,jsonb_build_object('name',$3::text))`, userID, item.ID, item.Name); err != nil {
		return Rule{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Rule{}, err
	}
	return item, nil
}

func (s *Service) Update(ctx context.Context, userID, id string, input Input) (Rule, error) {
	if err := validate(&input); err != nil {
		return Rule{}, err
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	raw, _ := json.Marshal(input.Filters)
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Rule{}, err
	}
	defer tx.Rollback(ctx)
	var item Rule
	item.Filters = input.Filters
	err = tx.QueryRow(ctx, `UPDATE alert_rules SET name=$3,enabled=$4,event_types=$5,filters=$6,updated_at=now() WHERE id=$1 AND user_id=$2 RETURNING id::text,name,enabled,event_types,created_at,updated_at`, id, userID, input.Name, enabled, input.EventTypes, raw).Scan(&item.ID, &item.Name, &item.Enabled, &item.EventTypes, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return Rule{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata) VALUES($1,'alert_rule.updated','alert_rule',$2,jsonb_build_object('name',$3::text,'enabled',$4::boolean))`, userID, item.ID, item.Name, item.Enabled); err != nil {
		return Rule{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Rule{}, err
	}
	return item, nil
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	var parsed pgtype.UUID
	if err := parsed.Scan(id); err != nil || !parsed.Valid {
		return pgx.ErrNoRows
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var name string
	if err = tx.QueryRow(ctx, `DELETE FROM alert_rules WHERE id=$1 AND user_id=$2 RETURNING name`, id, userID).Scan(&name); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata) VALUES($1,'alert_rule.deleted','alert_rule',$2,jsonb_build_object('name',$3::text))`, userID, id, name); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) ListEvents(ctx context.Context, userID string) ([]Event, error) {
	rows, err := s.db.Query(ctx, `SELECT e.id,e.rule_id::text,r.name,e.opportunity_id::text,o.package_identity,o.title,e.event_type,e.changed_fields,e.delivery_status,e.created_at,e.available_after FROM alert_events e JOIN alert_rules r ON r.id=e.rule_id JOIN opportunities o ON o.id=e.opportunity_id WHERE e.user_id=$1 ORDER BY e.created_at DESC,e.id DESC LIMIT 200`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Event, 0)
	for rows.Next() {
		var x Event
		if err := rows.Scan(&x.ID, &x.RuleID, &x.RuleName, &x.OpportunityID, &x.PackageIdentity, &x.Title, &x.EventType, &x.ChangedFields, &x.DeliveryStatus, &x.CreatedAt, &x.AvailableAfter); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}

func (s *Service) GetPreferences(ctx context.Context, userID string) (Preferences, error) {
	var x Preferences
	err := s.db.QueryRow(ctx, `SELECT quiet_hours_enabled,quiet_hours_start::text,quiet_hours_end::text,timezone,updated_at FROM notification_preferences WHERE user_id=$1`, userID).Scan(&x.QuietHoursEnabled, &x.QuietHoursStart, &x.QuietHoursEnd, &x.Timezone, &x.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Preferences{QuietHoursStart: "22:00:00", QuietHoursEnd: "07:00:00", Timezone: "Asia/Jakarta"}, nil
	}
	return x, err
}
func (s *Service) UpdatePreferences(ctx context.Context, userID string, input PreferencesInput) (Preferences, error) {
	if err := validatePreferences(input); err != nil {
		return Preferences{}, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Preferences{}, err
	}
	defer tx.Rollback(ctx)
	var x Preferences
	err = tx.QueryRow(ctx, `INSERT INTO notification_preferences(user_id,quiet_hours_enabled,quiet_hours_start,quiet_hours_end,timezone)VALUES($1,$2,$3::time,$4::time,$5) ON CONFLICT(user_id)DO UPDATE SET quiet_hours_enabled=excluded.quiet_hours_enabled,quiet_hours_start=excluded.quiet_hours_start,quiet_hours_end=excluded.quiet_hours_end,timezone=excluded.timezone,updated_at=now() RETURNING quiet_hours_enabled,quiet_hours_start::text,quiet_hours_end::text,timezone,updated_at`, userID, input.QuietHoursEnabled, input.QuietHoursStart, input.QuietHoursEnd, input.Timezone).Scan(&x.QuietHoursEnabled, &x.QuietHoursStart, &x.QuietHoursEnd, &x.Timezone, &x.UpdatedAt)
	if err != nil {
		return Preferences{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata)VALUES($1::uuid,'notification_preferences.updated','notification_preferences',$1::text,jsonb_build_object('quiet_hours_enabled',$2::boolean,'timezone',$3::text))`, userID, input.QuietHoursEnabled, input.Timezone)
	if err != nil {
		return Preferences{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Preferences{}, err
	}
	return x, nil
}

func Match(ctx context.Context, tx pgx.Tx, opportunityID string, versionID int64, eventType string, changedFields []string) error {
	_, err := tx.Exec(ctx, `INSERT INTO alert_events(rule_id,user_id,opportunity_id,opportunity_version_id,event_type,changed_fields,available_after)
	SELECT matched.id,matched.user_id,matched.opportunity_id,$2,$3,$4,
	CASE WHEN p.quiet_hours_enabled AND ((p.quiet_hours_start<p.quiet_hours_end AND (now() AT TIME ZONE p.timezone)::time>=p.quiet_hours_start AND (now() AT TIME ZONE p.timezone)::time<p.quiet_hours_end) OR (p.quiet_hours_start>p.quiet_hours_end AND ((now() AT TIME ZONE p.timezone)::time>=p.quiet_hours_start OR (now() AT TIME ZONE p.timezone)::time<p.quiet_hours_end))) THEN ((CASE WHEN p.quiet_hours_start>p.quiet_hours_end AND (now() AT TIME ZONE p.timezone)::time>=p.quiet_hours_start THEN (now() AT TIME ZONE p.timezone)::date+1 ELSE (now() AT TIME ZONE p.timezone)::date END)+p.quiet_hours_end) AT TIME ZONE p.timezone ELSE now() END
	FROM (
	SELECT DISTINCT ON(r.user_id) r.id,r.user_id,o.id AS opportunity_id FROM alert_rules r CROSS JOIN opportunities o
	LEFT JOIN LATERAL(SELECT disposition FROM opportunity_classifications c WHERE c.opportunity_id=o.id AND c.opportunity_fingerprint=o.current_fingerprint ORDER BY evaluated_at DESC LIMIT 1)c ON true
	LEFT JOIN LATERAL(SELECT total_score_bp FROM opportunity_scores s WHERE s.opportunity_id=o.id AND s.opportunity_fingerprint=o.current_fingerprint ORDER BY evaluated_at DESC LIMIT 1)s ON true
	WHERE o.id=$1 AND r.enabled AND $3=ANY(r.event_types) AND (coalesce(r.filters->>'source_id','')='' OR r.filters->>'source_id'=o.source_id::text) AND (coalesce(r.filters->>'relevance','all')='all' OR r.filters->>'relevance'=c.disposition) AND (NOT(r.filters?'min_score_bp') OR s.total_score_bp >= (r.filters->>'min_score_bp')::integer)
	ORDER BY r.user_id,(CASE WHEN r.filters?'source_id' THEN 1 ELSE 0 END + CASE WHEN coalesce(r.filters->>'relevance','all')<>'all' THEN 1 ELSE 0 END + CASE WHEN r.filters?'min_score_bp' THEN 1 ELSE 0 END) DESC,r.created_at,r.id
	) matched LEFT JOIN notification_preferences p ON p.user_id=matched.user_id ON CONFLICT DO NOTHING`, opportunityID, versionID, eventType, changedFields)
	if err != nil {
		return fmt.Errorf("match alert rules: %w", err)
	}
	return nil
}

func validate(input *Input) error {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len([]rune(input.Name)) > 80 {
		return errors.New("name must contain between 1 and 80 characters")
	}
	seen := map[string]bool{}
	for _, event := range input.EventTypes {
		if event != "opportunity.new" && event != "opportunity.material_change" {
			return errors.New("event_types contains an invalid event")
		}
		if seen[event] {
			return errors.New("event_types must be unique")
		}
		seen[event] = true
	}
	if len(seen) < 1 {
		return errors.New("event_types is required")
	}
	if input.Filters.Relevance == "" {
		input.Filters.Relevance = "all"
	}
	if input.Filters.Relevance != "all" && input.Filters.Relevance != "relevant" && input.Filters.Relevance != "excluded" && input.Filters.Relevance != "needs_review" {
		return errors.New("filters.relevance is invalid")
	}
	if input.Filters.MinimumScoreBP != nil && (*input.Filters.MinimumScoreBP < 0 || *input.Filters.MinimumScoreBP > 10000) {
		return errors.New("filters.min_score_bp must be between 0 and 10000")
	}
	if input.Filters.SourceID != "" {
		var id pgtype.UUID
		if err := id.Scan(input.Filters.SourceID); err != nil || !id.Valid {
			return errors.New("filters.source_id must be a UUID")
		}
	}
	return nil
}

func validatePreferences(input PreferencesInput) error {
	if input.QuietHoursStart == input.QuietHoursEnd {
		return errors.New("quiet hours start and end must differ")
	}
	for _, value := range []string{input.QuietHoursStart, input.QuietHoursEnd} {
		if _, err := time.Parse("15:04", value); err != nil {
			return errors.New("quiet hours must use HH:MM")
		}
	}
	if input.Timezone == "" {
		return errors.New("timezone is required")
	}
	if _, err := time.LoadLocation(input.Timezone); err != nil {
		return errors.New("timezone is invalid")
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scanRule(row scanner) (Rule, error) {
	var x Rule
	var raw []byte
	if err := row.Scan(&x.ID, &x.Name, &x.Enabled, &x.EventTypes, &raw, &x.CreatedAt, &x.UpdatedAt); err != nil {
		return Rule{}, err
	}
	if err := json.Unmarshal(raw, &x.Filters); err != nil {
		return Rule{}, err
	}
	return x, nil
}
