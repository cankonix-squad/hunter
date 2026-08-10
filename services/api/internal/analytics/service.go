package analytics

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pipeline struct {
	Total               int `json:"total"`
	Active              int `json:"active"`
	Shortlisted         int `json:"shortlisted"`
	Assigned            int `json:"assigned"`
	Pursue              int `json:"pursue"`
	Pass                int `json:"pass"`
	Undecided           int `json:"undecided"`
	OverdueActions      int `json:"overdue_actions"`
	ActionsDueSevenDays int `json:"actions_due_seven_days"`
}
type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }
func (s *Service) Pipeline(ctx context.Context) (Pipeline, error) {
	var x Pipeline
	err := s.db.QueryRow(ctx, `SELECT
	count(*)::integer,
	count(*) FILTER(WHERE o.lifecycle_status IN('announced','registration_open'))::integer,
	count(*) FILTER(WHERE sl.opportunity_id IS NOT NULL)::integer,
	count(*) FILTER(WHERE a.opportunity_id IS NOT NULL)::integer,
	count(*) FILTER(WHERE d.decision='pursue')::integer,
	count(*) FILTER(WHERE d.decision='pass')::integer,
	count(*) FILTER(WHERE d.opportunity_id IS NULL)::integer,
	count(*) FILTER(WHERE n.due_at<now())::integer,
	count(*) FILTER(WHERE n.due_at>=now() AND n.due_at<now()+interval '7 days')::integer
	FROM opportunities o LEFT JOIN opportunity_shortlist sl ON sl.opportunity_id=o.id
	LEFT JOIN opportunity_assignments a ON a.opportunity_id=o.id LEFT JOIN opportunity_decisions d ON d.opportunity_id=o.id
	LEFT JOIN opportunity_next_actions n ON n.opportunity_id=o.id`).Scan(&x.Total, &x.Active, &x.Shortlisted, &x.Assigned, &x.Pursue, &x.Pass, &x.Undecided, &x.OverdueActions, &x.ActionsDueSevenDays)
	if err != nil {
		return Pipeline{}, fmt.Errorf("pipeline analytics: %w", err)
	}
	return x, nil
}
