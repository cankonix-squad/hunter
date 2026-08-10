package comments

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Comment struct {
	ID            string    `json:"id"`
	OpportunityID string    `json:"opportunity_id"`
	AuthorUserID  string    `json:"author_user_id"`
	AuthorName    string    `json:"author_name"`
	Body          string    `json:"body"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateInput struct {
	Body string `json:"body"`
}
type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) List(ctx context.Context, opportunityID string) ([]Comment, error) {
	if !validUUID(opportunityID) {
		return nil, pgx.ErrNoRows
	}
	var exists bool
	if err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM opportunities WHERE id=$1)`, opportunityID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, pgx.ErrNoRows
	}
	rows, err := s.db.Query(ctx, `SELECT c.id::text,c.opportunity_id::text,c.author_user_id::text,
		u.display_name,c.body,c.created_at FROM opportunity_comments c JOIN users u ON u.id=c.author_user_id
		WHERE c.opportunity_id=$1 ORDER BY c.created_at,c.id`, opportunityID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()
	items := make([]Comment, 0)
	for rows.Next() {
		var item Comment
		if err := rows.Scan(&item.ID, &item.OpportunityID, &item.AuthorUserID, &item.AuthorName, &item.Body, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Create(ctx context.Context, opportunityID, authorID string, input CreateInput) (Comment, error) {
	input.Body = strings.TrimSpace(input.Body)
	if input.Body == "" || len([]rune(input.Body)) > 4000 {
		return Comment{}, errors.New("body must contain between 1 and 4000 characters")
	}
	if !validUUID(opportunityID) {
		return Comment{}, pgx.ErrNoRows
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Comment{}, err
	}
	defer tx.Rollback(ctx)
	var item Comment
	err = tx.QueryRow(ctx, `INSERT INTO opportunity_comments(opportunity_id,author_user_id,body)
		SELECT $1,$2,$3 WHERE EXISTS(SELECT 1 FROM opportunities WHERE id=$1)
		RETURNING id::text,opportunity_id::text,author_user_id::text,body,created_at`, opportunityID, authorID, input.Body).
		Scan(&item.ID, &item.OpportunityID, &item.AuthorUserID, &item.Body, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Comment{}, pgx.ErrNoRows
	}
	if err != nil {
		return Comment{}, fmt.Errorf("create comment: %w", err)
	}
	if err := tx.QueryRow(ctx, `SELECT display_name FROM users WHERE id=$1`, authorID).Scan(&item.AuthorName); err != nil {
		return Comment{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata)
		VALUES($1,'opportunity.comment_created','opportunity',$2,jsonb_build_object('comment_id',$3::text))`, authorID, opportunityID, item.ID)
	if err != nil {
		return Comment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Comment{}, err
	}
	return item, nil
}

func validUUID(value string) bool { var id pgtype.UUID; return id.Scan(value) == nil && id.Valid }
