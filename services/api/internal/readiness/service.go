package readiness

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

const Version = "readiness-v1"

type definition struct{ Code, Label string }

var definitions = []definition{{"scope_validated", "Scope and deliverables validated"}, {"eligibility_confirmed", "Eligibility requirements confirmed"}, {"delivery_capacity", "Delivery capacity confirmed"}, {"commercial_review", "Commercial assumptions reviewed"}, {"document_plan", "Submission document plan prepared"}}

type Item struct {
	Code        string     `json:"code"`
	Label       string     `json:"label"`
	Completed   bool       `json:"completed"`
	Note        string     `json:"note"`
	UpdatedBy   string     `json:"updated_by,omitempty"`
	UpdaterName string     `json:"updater_name,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}
type Checklist struct {
	OpportunityID string `json:"opportunity_id"`
	Version       string `json:"version"`
	Completed     int    `json:"completed"`
	Total         int    `json:"total"`
	Items         []Item `json:"items"`
}
type UpdateInput struct {
	Code      string `json:"code"`
	Completed bool   `json:"completed"`
	Note      string `json:"note"`
}
type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) Get(ctx context.Context, opportunityID string) (Checklist, error) {
	if !validUUID(opportunityID) {
		return Checklist{}, pgx.ErrNoRows
	}
	var exists bool
	if err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM opportunities WHERE id=$1)`, opportunityID).Scan(&exists); err != nil {
		return Checklist{}, err
	}
	if !exists {
		return Checklist{}, pgx.ErrNoRows
	}
	stored := make(map[string]Item)
	rows, err := s.db.Query(ctx, `SELECT r.item_code,r.completed,r.note,r.updated_by::text,u.display_name,r.updated_at FROM opportunity_readiness_items r JOIN users u ON u.id=r.updated_by WHERE r.opportunity_id=$1 AND r.checklist_version=$2`, opportunityID, Version)
	if err != nil {
		return Checklist{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.Code, &item.Completed, &item.Note, &item.UpdatedBy, &item.UpdaterName, &item.UpdatedAt); err != nil {
			return Checklist{}, err
		}
		stored[item.Code] = item
	}
	if err := rows.Err(); err != nil {
		return Checklist{}, err
	}
	result := Checklist{OpportunityID: opportunityID, Version: Version, Total: len(definitions), Items: make([]Item, 0, len(definitions))}
	for _, d := range definitions {
		item := stored[d.Code]
		item.Code = d.Code
		item.Label = d.Label
		if item.Completed {
			result.Completed++
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func (s *Service) Update(ctx context.Context, opportunityID, actorID string, input UpdateInput) (Checklist, error) {
	input.Code = strings.TrimSpace(input.Code)
	input.Note = strings.TrimSpace(input.Note)
	if !validCode(input.Code) {
		return Checklist{}, errors.New("unknown readiness item")
	}
	if len([]rune(input.Note)) > 1000 {
		return Checklist{}, errors.New("note must not exceed 1000 characters")
	}
	if !validUUID(opportunityID) {
		return Checklist{}, pgx.ErrNoRows
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Checklist{}, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM opportunities WHERE id=$1)`, opportunityID).Scan(&exists); err != nil {
		return Checklist{}, err
	}
	if !exists {
		return Checklist{}, pgx.ErrNoRows
	}
	_, err = tx.Exec(ctx, `INSERT INTO opportunity_readiness_items(opportunity_id,checklist_version,item_code,completed,note,updated_by) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(opportunity_id,checklist_version,item_code) DO UPDATE SET completed=excluded.completed,note=excluded.note,updated_by=excluded.updated_by,updated_at=now()`, opportunityID, Version, input.Code, input.Completed, input.Note, actorID)
	if err != nil {
		return Checklist{}, fmt.Errorf("update readiness: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata) VALUES($1,'opportunity.readiness_updated','opportunity',$2,jsonb_build_object('version',$3::text,'item_code',$4::text,'completed',$5::boolean,'note',$6::text))`, actorID, opportunityID, Version, input.Code, input.Completed, input.Note)
	if err != nil {
		return Checklist{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Checklist{}, err
	}
	return s.Get(ctx, opportunityID)
}
func validCode(code string) bool {
	for _, item := range definitions {
		if item.Code == code {
			return true
		}
	}
	return false
}
func validUUID(value string) bool { var id pgtype.UUID; return id.Scan(value) == nil && id.Valid }
