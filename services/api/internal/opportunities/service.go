package opportunities

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Item struct {
	ID                   string     `json:"id"`
	PackageIdentity      string     `json:"package_identity"`
	Title                string     `json:"title"`
	Agency               string     `json:"agency"`
	ProcurementType      string     `json:"procurement_type"`
	LifecycleStatus      string     `json:"lifecycle_status"`
	BudgetMinor          *int64     `json:"budget_minor,omitempty"`
	Currency             string     `json:"currency"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty"`
	CanonicalURL         string     `json:"canonical_url"`
	SourceTimestamp      *time.Time `json:"source_timestamp,omitempty"`
	FirstSeenAt          time.Time  `json:"first_seen_at"`
	LastSeenAt           time.Time  `json:"last_seen_at"`
	SourceID             string     `json:"source_id"`
	SourceName           string     `json:"source_name"`
	VersionCount         int        `json:"version_count"`
}

type Query struct {
	Search   string
	Status   string
	SourceID string
	Page     int
	PageSize int
}

type Page struct {
	Items      []Item `json:"items"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Total      int    `json:"total"`
	TotalPages int    `json:"total_pages"`
}

type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) List(ctx context.Context, query Query) (Page, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	query.Search = strings.TrimSpace(query.Search)
	query.Status = strings.TrimSpace(query.Status)
	query.SourceID = strings.TrimSpace(query.SourceID)
	if query.Status == "" {
		query.Status = "active"
	}
	if query.Status != "all" && query.Status != "active" && !validStatuses[query.Status] {
		return Page{}, errors.New("invalid lifecycle status")
	}

	statusSQL := "($2 = 'all' OR ($2 = 'active' AND o.lifecycle_status IN ('announced','registration_open')) OR o.lifecycle_status = $2)"
	filterSQL := "($1 = '' OR (o.title || ' ' || o.agency || ' ' || o.package_identity) ILIKE '%' || $1 || '%') AND " + statusSQL + " AND ($3 = '' OR o.source_id::text = $3)"
	var total int
	if err := s.db.QueryRow(ctx, "SELECT count(*) FROM opportunities o WHERE "+filterSQL, query.Search, query.Status, query.SourceID).Scan(&total); err != nil {
		return Page{}, fmt.Errorf("count opportunities: %w", err)
	}
	rows, err := s.db.Query(ctx, `SELECT o.id::text, o.package_identity, o.title, o.agency, o.procurement_type,
		o.lifecycle_status, o.budget_minor, o.currency, o.registration_deadline, o.canonical_url,
		o.source_timestamp, o.first_seen_at, o.last_seen_at, s.id::text, s.name,
		(SELECT count(*) FROM opportunity_versions v WHERE v.opportunity_id = o.id)
		FROM opportunities o JOIN sources s ON s.id = o.source_id WHERE `+filterSQL+`
		ORDER BY o.registration_deadline ASC NULLS LAST, o.last_seen_at DESC
		LIMIT $4 OFFSET $5`, query.Search, query.Status, query.SourceID, query.PageSize, (query.Page-1)*query.PageSize)
	if err != nil {
		return Page{}, fmt.Errorf("list opportunities: %w", err)
	}
	defer rows.Close()
	items := make([]Item, 0)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return Page{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return Page{}, fmt.Errorf("iterate opportunities: %w", err)
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + query.PageSize - 1) / query.PageSize
	}
	return Page{Items: items, Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
}

func (s *Service) Get(ctx context.Context, id string) (Item, error) {
	row := s.db.QueryRow(ctx, `SELECT o.id::text, o.package_identity, o.title, o.agency, o.procurement_type,
		o.lifecycle_status, o.budget_minor, o.currency, o.registration_deadline, o.canonical_url,
		o.source_timestamp, o.first_seen_at, o.last_seen_at, s.id::text, s.name,
		(SELECT count(*) FROM opportunity_versions v WHERE v.opportunity_id = o.id)
		FROM opportunities o JOIN sources s ON s.id = o.source_id WHERE o.id = $1`, id)
	item, err := scanItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, pgx.ErrNoRows
	}
	return item, err
}

type scanner interface{ Scan(...any) error }

func scanItem(row scanner) (Item, error) {
	var item Item
	err := row.Scan(&item.ID, &item.PackageIdentity, &item.Title, &item.Agency, &item.ProcurementType,
		&item.LifecycleStatus, &item.BudgetMinor, &item.Currency, &item.RegistrationDeadline,
		&item.CanonicalURL, &item.SourceTimestamp, &item.FirstSeenAt, &item.LastSeenAt,
		&item.SourceID, &item.SourceName, &item.VersionCount)
	if err != nil {
		return Item{}, err
	}
	return item, nil
}
