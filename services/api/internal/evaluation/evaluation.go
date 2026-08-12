package evaluation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/classification"
)

var dispositions = []string{"relevant", "excluded", "needs_review"}

type Dataset struct {
	Version         string `json:"version"`
	TaxonomyVersion string `json:"taxonomy_version"`
	ReviewStatus    string `json:"review_status"`
	ReviewedBy      string `json:"reviewed_by,omitempty"`
	ReviewedAt      string `json:"reviewed_at,omitempty"`
	ApprovalRef     string `json:"approval_reference,omitempty"`
	Cases           []Case `json:"cases"`
}

type Case struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	ProcurementType     string `json:"procurement_type"`
	ExpectedDisposition string `json:"expected_disposition"`
	ExpectedCategory    string `json:"expected_category,omitempty"`
}

type LabelMetrics struct {
	TruePositive  int `json:"true_positive"`
	FalsePositive int `json:"false_positive"`
	FalseNegative int `json:"false_negative"`
	PrecisionBP   int `json:"precision_bp"`
	RecallBP      int `json:"recall_bp"`
}

type Failure struct {
	ID                  string `json:"id"`
	ExpectedDisposition string `json:"expected_disposition"`
	ActualDisposition   string `json:"actual_disposition"`
	ExpectedCategory    string `json:"expected_category,omitempty"`
	ActualCategory      string `json:"actual_category,omitempty"`
}

type Report struct {
	DatasetVersion   string                    `json:"dataset_version"`
	TaxonomyVersion  string                    `json:"taxonomy_version"`
	ReviewStatus     string                    `json:"review_status"`
	ReviewedBy       string                    `json:"reviewed_by,omitempty"`
	ReviewedAt       string                    `json:"reviewed_at,omitempty"`
	ApprovalRef      string                    `json:"approval_reference,omitempty"`
	RuleVersion      string                    `json:"rule_version"`
	Total            int                       `json:"total"`
	Correct          int                       `json:"correct"`
	AccuracyBP       int                       `json:"accuracy_bp"`
	MacroPrecisionBP int                       `json:"macro_precision_bp"`
	Labels           map[string]LabelMetrics   `json:"labels"`
	Confusion        map[string]map[string]int `json:"confusion"`
	Failures         []Failure                 `json:"failures"`
}

func Load(reader io.Reader) (Dataset, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var dataset Dataset
	if err := decoder.Decode(&dataset); err != nil {
		return Dataset{}, fmt.Errorf("decode golden dataset: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return Dataset{}, errors.New("golden dataset must contain one JSON object")
	}
	if err := validate(dataset); err != nil {
		return Dataset{}, err
	}
	return dataset, nil
}

func Evaluate(dataset Dataset, engine classification.Engine) (Report, error) {
	if err := validate(dataset); err != nil {
		return Report{}, err
	}
	report := Report{
		DatasetVersion: dataset.Version, TaxonomyVersion: dataset.TaxonomyVersion, ReviewStatus: dataset.ReviewStatus,
		ReviewedBy: dataset.ReviewedBy, ReviewedAt: dataset.ReviewedAt, ApprovalRef: dataset.ApprovalRef,
		Total: len(dataset.Cases), Labels: make(map[string]LabelMetrics),
		Confusion: make(map[string]map[string]int), Failures: make([]Failure, 0),
	}
	for _, expected := range dispositions {
		report.Confusion[expected] = make(map[string]int)
	}
	for _, item := range dataset.Cases {
		actual := engine.Evaluate(classification.Input{Title: item.Title, ProcurementType: item.ProcurementType})
		report.RuleVersion = actual.RuleVersion
		report.Confusion[item.ExpectedDisposition][actual.Disposition]++
		correct := actual.Disposition == item.ExpectedDisposition &&
			(item.ExpectedCategory == "" || actual.Category == item.ExpectedCategory)
		if correct {
			report.Correct++
		} else {
			report.Failures = append(report.Failures, Failure{ID: item.ID,
				ExpectedDisposition: item.ExpectedDisposition, ActualDisposition: actual.Disposition,
				ExpectedCategory: item.ExpectedCategory, ActualCategory: actual.Category})
		}
	}
	report.AccuracyBP = basisPoints(report.Correct, report.Total)
	precisionTotal := 0
	for _, label := range dispositions {
		tp := report.Confusion[label][label]
		fp, fn := 0, 0
		for _, other := range dispositions {
			if other != label {
				fp += report.Confusion[other][label]
				fn += report.Confusion[label][other]
			}
		}
		metrics := LabelMetrics{TruePositive: tp, FalsePositive: fp, FalseNegative: fn,
			PrecisionBP: basisPoints(tp, tp+fp), RecallBP: basisPoints(tp, tp+fn)}
		report.Labels[label] = metrics
		precisionTotal += metrics.PrecisionBP
	}
	report.MacroPrecisionBP = precisionTotal / len(dispositions)
	return report, nil
}

func validate(dataset Dataset) error {
	if strings.TrimSpace(dataset.Version) == "" {
		return errors.New("golden dataset version is required")
	}
	if dataset.TaxonomyVersion != classification.TaxonomyVersion {
		return fmt.Errorf("golden dataset taxonomy_version %q does not match engine %q", dataset.TaxonomyVersion, classification.TaxonomyVersion)
	}
	if dataset.ReviewStatus != "candidate" && dataset.ReviewStatus != "approved" {
		return errors.New("golden dataset review_status must be candidate or approved")
	}
	if dataset.ReviewStatus == "approved" {
		if strings.TrimSpace(dataset.ReviewedBy) == "" || strings.TrimSpace(dataset.ApprovalRef) == "" {
			return errors.New("approved golden dataset requires reviewed_by and approval_reference")
		}
		if reviewedAt, err := time.Parse(time.RFC3339, dataset.ReviewedAt); err != nil || reviewedAt.After(time.Now().Add(5*time.Minute)) {
			return errors.New("approved golden dataset requires a valid, non-future reviewed_at RFC3339 timestamp")
		}
	} else if dataset.ReviewedBy != "" || dataset.ReviewedAt != "" || dataset.ApprovalRef != "" {
		return errors.New("candidate golden dataset must not contain approval metadata")
	}
	if len(dataset.Cases) == 0 {
		return errors.New("golden dataset must contain at least one case")
	}
	seen := make(map[string]bool)
	for index, item := range dataset.Cases {
		if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Title) == "" {
			return fmt.Errorf("golden dataset case %d requires id and title", index)
		}
		if seen[item.ID] {
			return fmt.Errorf("golden dataset case id %q is duplicated", item.ID)
		}
		seen[item.ID] = true
		if !contains(dispositions, item.ExpectedDisposition) {
			return fmt.Errorf("golden dataset case %q has invalid disposition %q", item.ID, item.ExpectedDisposition)
		}
		if item.ExpectedDisposition != "relevant" && item.ExpectedCategory != "" {
			return fmt.Errorf("golden dataset case %q may only set category when relevant", item.ID)
		}
	}
	return nil
}

func basisPoints(numerator, denominator int) int {
	if denominator == 0 {
		return 0
	}
	return numerator * 10000 / denominator
}

func contains(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
