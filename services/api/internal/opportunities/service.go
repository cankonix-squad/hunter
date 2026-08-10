package opportunities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Item struct {
	ID                   string          `json:"id"`
	PackageIdentity      string          `json:"package_identity"`
	Title                string          `json:"title"`
	Agency               string          `json:"agency"`
	ProcurementType      string          `json:"procurement_type"`
	LifecycleStatus      string          `json:"lifecycle_status"`
	BudgetMinor          *int64          `json:"budget_minor,omitempty"`
	Currency             string          `json:"currency"`
	RegistrationDeadline *time.Time      `json:"registration_deadline,omitempty"`
	CanonicalURL         string          `json:"canonical_url"`
	SourceTimestamp      *time.Time      `json:"source_timestamp,omitempty"`
	FirstSeenAt          time.Time       `json:"first_seen_at"`
	LastSeenAt           time.Time       `json:"last_seen_at"`
	SourceID             string          `json:"source_id"`
	SourceName           string          `json:"source_name"`
	VersionCount         int             `json:"version_count"`
	Classification       *Classification `json:"classification,omitempty"`
	Score                *Score          `json:"score,omitempty"`
	History              []Version       `json:"history,omitempty"`
}

type Version struct {
	ID              int64      `json:"id"`
	SnapshotID      string     `json:"snapshot_id"`
	Fingerprint     string     `json:"fingerprint"`
	LifecycleStatus string     `json:"lifecycle_status"`
	Data            Normalized `json:"data"`
	ChangedFields   []string   `json:"changed_fields"`
	ObservedAt      time.Time  `json:"observed_at"`
}

type Classification struct {
	Disposition  string                   `json:"disposition"`
	CategoryCode string                   `json:"category_code,omitempty"`
	ConfidenceBP int                      `json:"confidence_bp"`
	ReasonCode   string                   `json:"reason_code"`
	Explanation  string                   `json:"explanation"`
	AutoRejected bool                     `json:"auto_rejected"`
	Evidence     []ClassificationEvidence `json:"evidence"`
}

type ClassificationEvidence struct {
	RuleCode     string `json:"rule_code"`
	Kind         string `json:"kind"`
	CategoryCode string `json:"category_code,omitempty"`
	Keyword      string `json:"keyword"`
	SourceField  string `json:"source_field"`
	Excerpt      string `json:"excerpt"`
}

type Score struct {
	Version      string           `json:"version"`
	TotalScoreBP int              `json:"total_score_bp"`
	ConfidenceBP int              `json:"confidence_bp"`
	NeedsReview  bool             `json:"needs_review"`
	ReasonCode   string           `json:"reason_code"`
	Explanation  string           `json:"explanation"`
	Dimensions   []ScoreDimension `json:"dimensions"`
}

type ScoreDimension struct {
	Code         string `json:"code"`
	WeightBP     int    `json:"weight_bp"`
	RawScoreBP   int    `json:"raw_score_bp"`
	Contribution int    `json:"contribution"`
	ReasonCode   string `json:"reason_code"`
	Explanation  string `json:"explanation"`
	Evidence     string `json:"evidence"`
}

type Query struct {
	Search         string
	Status         string
	SourceID       string
	Page           int
	PageSize       int
	Relevance      string
	MinimumScoreBP int
	NeedsReview    string
	Shortlisted    string
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
	query.Relevance = strings.TrimSpace(query.Relevance)
	query.NeedsReview = strings.TrimSpace(query.NeedsReview)
	query.Shortlisted = strings.TrimSpace(query.Shortlisted)
	if query.Status == "" {
		query.Status = "active"
	}
	if query.Status != "all" && query.Status != "active" && !validStatuses[query.Status] {
		return Page{}, errors.New("invalid lifecycle status")
	}
	if query.Relevance == "" {
		query.Relevance = "all"
	}
	if query.Relevance != "all" && query.Relevance != "relevant" && query.Relevance != "excluded" && query.Relevance != "needs_review" {
		return Page{}, errors.New("invalid relevance")
	}
	if query.MinimumScoreBP < -1 || query.MinimumScoreBP > 10000 {
		return Page{}, errors.New("invalid minimum score")
	}
	if query.NeedsReview == "" {
		query.NeedsReview = "all"
	}
	if query.NeedsReview != "all" && query.NeedsReview != "true" && query.NeedsReview != "false" {
		return Page{}, errors.New("invalid needs review")
	}
	if query.Shortlisted == "" {
		query.Shortlisted = "all"
	}
	if query.Shortlisted != "all" && query.Shortlisted != "true" && query.Shortlisted != "false" {
		return Page{}, errors.New("invalid shortlisted")
	}

	statusSQL := "($2 = 'all' OR ($2 = 'active' AND o.lifecycle_status IN ('announced','registration_open')) OR o.lifecycle_status = $2)"
	filterSQL := "($1 = '' OR (o.title || ' ' || o.agency || ' ' || o.package_identity) ILIKE '%' || $1 || '%') AND " + statusSQL +
		" AND ($3 = '' OR o.source_id::text = $3) AND ($4 = 'all' OR c.disposition = $4)" +
		" AND ($5 < 0 OR sc.total_score_bp >= $5) AND ($6 = 'all' OR sc.needs_review = ($6 = 'true'))" +
		" AND ($7 = 'all' OR (sl.opportunity_id IS NOT NULL) = ($7 = 'true'))"
	joins := ` JOIN sources src ON src.id = o.source_id
		LEFT JOIN LATERAL (SELECT * FROM opportunity_classifications x WHERE x.opportunity_id=o.id AND x.opportunity_fingerprint=o.current_fingerprint ORDER BY x.evaluated_at DESC LIMIT 1) c ON true
		LEFT JOIN LATERAL (SELECT * FROM opportunity_scores x WHERE x.opportunity_id=o.id AND x.opportunity_fingerprint=o.current_fingerprint ORDER BY x.evaluated_at DESC LIMIT 1) sc ON true
		LEFT JOIN opportunity_shortlist sl ON sl.opportunity_id=o.id `
	var total int
	args := []any{query.Search, query.Status, query.SourceID, query.Relevance, query.MinimumScoreBP, query.NeedsReview, query.Shortlisted}
	if err := s.db.QueryRow(ctx, "SELECT count(*) FROM opportunities o "+joins+" WHERE "+filterSQL, args...).Scan(&total); err != nil {
		return Page{}, fmt.Errorf("count opportunities: %w", err)
	}
	rows, err := s.db.Query(ctx, `SELECT o.id::text, o.package_identity, o.title, o.agency, o.procurement_type,
		o.lifecycle_status, o.budget_minor, o.currency, o.registration_deadline, o.canonical_url,
		o.source_timestamp, o.first_seen_at, o.last_seen_at, src.id::text, src.name,
		(SELECT count(*) FROM opportunity_versions v WHERE v.opportunity_id = o.id),
		c.id IS NOT NULL, coalesce(c.disposition,''), coalesce(c.category_code,''), coalesce(c.confidence_bp,0),
		coalesce(c.reason_code,''), coalesce(c.explanation,''), coalesce(c.auto_rejected,false),
		sc.id IS NOT NULL, coalesce(sc.scoring_version,''), coalesce(sc.total_score_bp,0), coalesce(sc.confidence_bp,0),
		coalesce(sc.needs_review,false), coalesce(sc.reason_code,''), coalesce(sc.explanation,'')
		FROM opportunities o `+joins+` WHERE `+filterSQL+`
		ORDER BY o.registration_deadline ASC NULLS LAST, o.last_seen_at DESC
		LIMIT $8 OFFSET $9`, append(args, query.PageSize, (query.Page-1)*query.PageSize)...)
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
		o.source_timestamp, o.first_seen_at, o.last_seen_at, src.id::text, src.name,
		(SELECT count(*) FROM opportunity_versions v WHERE v.opportunity_id = o.id),
		c.id IS NOT NULL, coalesce(c.disposition,''), coalesce(c.category_code,''), coalesce(c.confidence_bp,0),
		coalesce(c.reason_code,''), coalesce(c.explanation,''), coalesce(c.auto_rejected,false),
		sc.id IS NOT NULL, coalesce(sc.scoring_version,''), coalesce(sc.total_score_bp,0), coalesce(sc.confidence_bp,0),
		coalesce(sc.needs_review,false), coalesce(sc.reason_code,''), coalesce(sc.explanation,'')
		FROM opportunities o JOIN sources src ON src.id = o.source_id
		LEFT JOIN LATERAL (SELECT * FROM opportunity_classifications x WHERE x.opportunity_id=o.id AND x.opportunity_fingerprint=o.current_fingerprint ORDER BY x.evaluated_at DESC LIMIT 1) c ON true
		LEFT JOIN LATERAL (SELECT * FROM opportunity_scores x WHERE x.opportunity_id=o.id AND x.opportunity_fingerprint=o.current_fingerprint ORDER BY x.evaluated_at DESC LIMIT 1) sc ON true
		WHERE o.id = $1`, id)
	item, err := scanItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, pgx.ErrNoRows
	}
	if err != nil {
		return Item{}, err
	}
	if item.Classification != nil {
		rows, err := s.db.Query(ctx, `SELECT e.rule_code,e.kind,coalesce(e.category_code,''),e.keyword,e.source_field,e.excerpt
			FROM classification_evidence e WHERE e.classification_id=(SELECT c.id FROM opportunity_classifications c
			WHERE c.opportunity_id=$1 AND c.opportunity_fingerprint=(SELECT current_fingerprint FROM opportunities WHERE id=$1)
			ORDER BY c.evaluated_at DESC LIMIT 1)
			ORDER BY e.id`, id)
		if err != nil {
			return Item{}, fmt.Errorf("load classification evidence: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var e ClassificationEvidence
			if err := rows.Scan(&e.RuleCode, &e.Kind, &e.CategoryCode, &e.Keyword, &e.SourceField, &e.Excerpt); err != nil {
				return Item{}, err
			}
			item.Classification.Evidence = append(item.Classification.Evidence, e)
		}
	}
	if item.Score != nil {
		rows, err := s.db.Query(ctx, `SELECT d.dimension_code,d.weight_bp,d.raw_score_bp,d.contribution,d.reason_code,d.explanation,d.evidence
			FROM score_dimensions d WHERE d.score_id=(SELECT sc.id FROM opportunity_scores sc
			WHERE sc.opportunity_id=$1 AND sc.opportunity_fingerprint=(SELECT current_fingerprint FROM opportunities WHERE id=$1)
			ORDER BY sc.evaluated_at DESC LIMIT 1)
			ORDER BY d.id`, id)
		if err != nil {
			return Item{}, fmt.Errorf("load score dimensions: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var d ScoreDimension
			if err := rows.Scan(&d.Code, &d.WeightBP, &d.RawScoreBP, &d.Contribution, &d.ReasonCode, &d.Explanation, &d.Evidence); err != nil {
				return Item{}, err
			}
			item.Score.Dimensions = append(item.Score.Dimensions, d)
		}
	}
	historyRows, err := s.db.Query(ctx, `SELECT id,snapshot_id::text,fingerprint,lifecycle_status,normalized_data,observed_at
		FROM opportunity_versions WHERE opportunity_id=$1 ORDER BY observed_at,id`, id)
	if err != nil {
		return Item{}, fmt.Errorf("load opportunity history: %w", err)
	}
	defer historyRows.Close()
	chronological := make([]Version, 0, item.VersionCount)
	var previous *Normalized
	for historyRows.Next() {
		var version Version
		var raw []byte
		if err := historyRows.Scan(&version.ID, &version.SnapshotID, &version.Fingerprint, &version.LifecycleStatus, &raw, &version.ObservedAt); err != nil {
			return Item{}, fmt.Errorf("scan opportunity history: %w", err)
		}
		if err := json.Unmarshal(raw, &version.Data); err != nil {
			return Item{}, fmt.Errorf("decode opportunity history: %w", err)
		}
		version.ChangedFields = changedFields(previous, version.Data)
		current := version.Data
		previous = &current
		chronological = append(chronological, version)
	}
	if err := historyRows.Err(); err != nil {
		return Item{}, fmt.Errorf("iterate opportunity history: %w", err)
	}
	item.History = make([]Version, len(chronological))
	for index := range chronological {
		item.History[len(chronological)-1-index] = chronological[index]
	}
	return item, nil
}

func changedFields(previous *Normalized, current Normalized) []string {
	if previous == nil {
		return []string{}
	}
	fields := make([]string, 0)
	if previous.Title != current.Title {
		fields = append(fields, "title")
	}
	if previous.Agency != current.Agency {
		fields = append(fields, "agency")
	}
	if previous.ProcurementType != current.ProcurementType {
		fields = append(fields, "procurement_type")
	}
	if previous.LifecycleStatus != current.LifecycleStatus {
		fields = append(fields, "lifecycle_status")
	}
	if !equalInt64(previous.BudgetMinor, current.BudgetMinor) {
		fields = append(fields, "budget_minor")
	}
	if previous.Currency != current.Currency {
		fields = append(fields, "currency")
	}
	if !equalTime(previous.RegistrationDeadline, current.RegistrationDeadline) {
		fields = append(fields, "registration_deadline")
	}
	return fields
}

func equalInt64(left, right *int64) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func equalTime(left, right *time.Time) bool {
	return left == nil && right == nil || left != nil && right != nil && left.Equal(*right)
}

type scanner interface{ Scan(...any) error }

func scanItem(row scanner) (Item, error) {
	var item Item
	var hasClassification, hasScore bool
	var classification Classification
	var score Score
	err := row.Scan(&item.ID, &item.PackageIdentity, &item.Title, &item.Agency, &item.ProcurementType,
		&item.LifecycleStatus, &item.BudgetMinor, &item.Currency, &item.RegistrationDeadline,
		&item.CanonicalURL, &item.SourceTimestamp, &item.FirstSeenAt, &item.LastSeenAt,
		&item.SourceID, &item.SourceName, &item.VersionCount,
		&hasClassification, &classification.Disposition, &classification.CategoryCode, &classification.ConfidenceBP,
		&classification.ReasonCode, &classification.Explanation, &classification.AutoRejected,
		&hasScore, &score.Version, &score.TotalScoreBP, &score.ConfidenceBP, &score.NeedsReview,
		&score.ReasonCode, &score.Explanation)
	if err != nil {
		return Item{}, err
	}
	if hasClassification {
		classification.Evidence = []ClassificationEvidence{}
		item.Classification = &classification
	}
	if hasScore {
		score.Dimensions = []ScoreDimension{}
		item.Score = &score
	}
	return item, nil
}
