package savedviews

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
	Search         string `json:"q"`
	Status         string `json:"status"`
	SourceID       string `json:"source_id,omitempty"`
	Relevance      string `json:"relevance"`
	MinimumScoreBP *int   `json:"min_score_bp,omitempty"`
	ReviewOnly     bool   `json:"review_only"`
	Shortlisted    string `json:"shortlisted"`
}

type View struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Filters   Filters   `json:"filters"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInput struct {
	Name    string  `json:"name"`
	Filters Filters `json:"filters"`
}

type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) List(ctx context.Context, userID string) ([]View, error) {
	rows, err := s.db.Query(ctx, `SELECT id::text,name,filters,created_at,updated_at
		FROM saved_opportunity_views WHERE user_id=$1 ORDER BY lower(name),created_at`, userID)
	if err != nil {
		return nil, fmt.Errorf("list saved views: %w", err)
	}
	defer rows.Close()
	items := make([]View, 0)
	for rows.Next() {
		var item View
		var raw []byte
		if err := rows.Scan(&item.ID, &item.Name, &raw, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan saved view: %w", err)
		}
		if err := json.Unmarshal(raw, &item.Filters); err != nil {
			return nil, fmt.Errorf("decode saved view filters: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (View, error) {
	if err := validate(&input); err != nil {
		return View{}, err
	}
	raw, err := json.Marshal(input.Filters)
	if err != nil {
		return View{}, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return View{}, err
	}
	defer tx.Rollback(ctx)
	var item View
	item.Filters = input.Filters
	err = tx.QueryRow(ctx, `INSERT INTO saved_opportunity_views(user_id,name,filters) VALUES($1,$2,$3)
		RETURNING id::text,name,created_at,updated_at`, userID, input.Name, raw).
		Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return View{}, fmt.Errorf("create saved view: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata)
		VALUES($1,'saved_view.created','saved_opportunity_view',$2,jsonb_build_object('name',$3::text))`, userID, item.ID, item.Name); err != nil {
		return View{}, fmt.Errorf("audit saved view: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return View{}, err
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
	if err := tx.QueryRow(ctx, `DELETE FROM saved_opportunity_views WHERE id=$1 AND user_id=$2 RETURNING name`, id, userID).Scan(&name); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata)
		VALUES($1,'saved_view.deleted','saved_opportunity_view',$2,jsonb_build_object('name',$3::text))`, userID, id, name); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func validate(input *CreateInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.Filters.Search = strings.TrimSpace(input.Filters.Search)
	if input.Name == "" || len([]rune(input.Name)) > 80 {
		return errors.New("name must contain between 1 and 80 characters")
	}
	if input.Filters.Status == "" {
		input.Filters.Status = "active"
	}
	validStatus := map[string]bool{"active": true, "all": true, "announced": true, "registration_open": true, "registration_closed": true, "evaluation": true, "awarded": true, "contracted": true, "completed": true, "cancelled": true}
	if !validStatus[input.Filters.Status] {
		return errors.New("filters.status is invalid")
	}
	if input.Filters.Relevance == "" {
		input.Filters.Relevance = "all"
	}
	if input.Filters.Shortlisted == "" {
		input.Filters.Shortlisted = "all"
	}
	if input.Filters.Shortlisted != "all" && input.Filters.Shortlisted != "true" && input.Filters.Shortlisted != "false" {
		return errors.New("filters.shortlisted is invalid")
	}
	if input.Filters.Relevance != "all" && input.Filters.Relevance != "relevant" && input.Filters.Relevance != "excluded" && input.Filters.Relevance != "needs_review" {
		return errors.New("filters.relevance is invalid")
	}
	if input.Filters.MinimumScoreBP != nil && (*input.Filters.MinimumScoreBP < 0 || *input.Filters.MinimumScoreBP > 10000) {
		return errors.New("filters.min_score_bp must be between 0 and 10000")
	}
	if input.Filters.SourceID != "" {
		var parsed pgtype.UUID
		if err := parsed.Scan(input.Filters.SourceID); err != nil || !parsed.Valid {
			return errors.New("filters.source_id must be a UUID")
		}
	}
	return nil
}
