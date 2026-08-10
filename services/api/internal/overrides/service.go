package overrides

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

type Override struct {
	OpportunityID  string    `json:"opportunity_id"`
	Recommendation *string   `json:"recommendation,omitempty"`
	ScoreBP        *int      `json:"score_bp,omitempty"`
	Reason         string    `json:"reason"`
	OverriddenBy   string    `json:"overridden_by"`
	OverriderName  string    `json:"overrider_name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
type UpdateInput struct {
	Recommendation *string `json:"recommendation"`
	ScoreBP        *int    `json:"score_bp"`
	Reason         string  `json:"reason"`
	Clear          bool    `json:"clear"`
}
type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }
func (s *Service) Get(ctx context.Context, id string) (*Override, error) {
	if !validUUID(id) {
		return nil, pgx.ErrNoRows
	}
	var exists bool
	if err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM opportunities WHERE id=$1)`, id).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, pgx.ErrNoRows
	}
	var x Override
	err := s.db.QueryRow(ctx, `SELECT o.opportunity_id::text,o.recommendation,o.score_bp,o.reason,o.overridden_by::text,u.display_name,o.created_at,o.updated_at FROM opportunity_overrides o JOIN users u ON u.id=o.overridden_by WHERE o.opportunity_id=$1`, id).Scan(&x.OpportunityID, &x.Recommendation, &x.ScoreBP, &x.Reason, &x.OverriddenBy, &x.OverriderName, &x.CreatedAt, &x.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get override: %w", err)
	}
	return &x, nil
}
func (s *Service) Update(ctx context.Context, id, actor string, in UpdateInput) (*Override, error) {
	if !validUUID(id) {
		return nil, pgx.ErrNoRows
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM opportunities WHERE id=$1)`, id).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, pgx.ErrNoRows
	}
	if in.Clear {
		_, err = tx.Exec(ctx, `DELETE FROM opportunity_overrides WHERE opportunity_id=$1`, id)
		if err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id) VALUES($1,'opportunity.override_cleared','opportunity',$2)`, actor, id)
		}
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if len([]rune(in.Reason)) < 3 || len([]rune(in.Reason)) > 2000 {
		return nil, errors.New("reason is required")
	}
	if in.Recommendation == nil && in.ScoreBP == nil {
		return nil, errors.New("recommendation or score is required")
	}
	if in.Recommendation != nil && *in.Recommendation != "relevant" && *in.Recommendation != "excluded" && *in.Recommendation != "needs_review" {
		return nil, errors.New("invalid recommendation")
	}
	if in.ScoreBP != nil && (*in.ScoreBP < 0 || *in.ScoreBP > 10000) {
		return nil, errors.New("invalid score")
	}
	_, err = tx.Exec(ctx, `INSERT INTO opportunity_overrides(opportunity_id,recommendation,score_bp,reason,overridden_by) VALUES($1,$2,$3,$4,$5) ON CONFLICT(opportunity_id) DO UPDATE SET recommendation=excluded.recommendation,score_bp=excluded.score_bp,reason=excluded.reason,overridden_by=excluded.overridden_by,updated_at=now()`, id, in.Recommendation, in.ScoreBP, in.Reason, actor)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata) VALUES($1,'opportunity.override_updated','opportunity',$2,jsonb_build_object('recommendation',$3::text,'score_bp',$4::integer,'reason',$5::text))`, actor, id, in.Recommendation, in.ScoreBP, in.Reason)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}
func validUUID(v string) bool { var id pgtype.UUID; return id.Scan(v) == nil && id.Valid }
