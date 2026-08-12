package evaluation

import (
	"strings"
	"testing"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/classification"
)

func TestEvaluateComputesConfusionAndPrecision(t *testing.T) {
	dataset := Dataset{Version: "test-v1", TaxonomyVersion: classification.TaxonomyVersion, ReviewStatus: "candidate", Cases: []Case{
		{ID: "relevant", Title: "Pengembangan aplikasi web", ExpectedDisposition: "relevant", ExpectedCategory: "software-engineering"},
		{ID: "excluded", Title: "Pengadaan alat tulis kantor", ExpectedDisposition: "excluded"},
		{ID: "review", Title: "Transformasi organisasi", ExpectedDisposition: "needs_review"},
	}}
	report, err := Evaluate(dataset, classification.DefaultEngine())
	if err != nil {
		t.Fatal(err)
	}
	if report.AccuracyBP != 10000 || report.MacroPrecisionBP != 10000 || report.Correct != 3 || len(report.Failures) != 0 {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestApprovedDatasetRequiresAuditableReviewMetadata(t *testing.T) {
	dataset := Dataset{Version: "test-v1", TaxonomyVersion: classification.TaxonomyVersion, ReviewStatus: "approved", Cases: []Case{
		{ID: "relevant", Title: "Pengembangan aplikasi web", ExpectedDisposition: "relevant", ExpectedCategory: "software-engineering"},
	}}
	if _, err := Evaluate(dataset, classification.DefaultEngine()); err == nil {
		t.Fatal("approved dataset without review metadata was accepted")
	}
	dataset.ReviewedBy = "Named Domain Owner"
	dataset.ReviewedAt = time.Now().UTC().Format(time.RFC3339)
	dataset.ApprovalRef = "UAT-CLASSIFICATION-001"
	if _, err := Evaluate(dataset, classification.DefaultEngine()); err != nil {
		t.Fatalf("auditable approved dataset rejected: %v", err)
	}
}

func TestCandidateDatasetRejectsApprovalMetadata(t *testing.T) {
	dataset := Dataset{Version: "test-v1", TaxonomyVersion: classification.TaxonomyVersion, ReviewStatus: "candidate",
		ReviewedBy: "Premature Reviewer", Cases: []Case{{ID: "one", Title: "One", ExpectedDisposition: "needs_review"}}}
	if _, err := Evaluate(dataset, classification.DefaultEngine()); err == nil {
		t.Fatal("candidate dataset with approval metadata was accepted")
	}
}

func TestLoadRejectsUnknownFieldsAndDuplicateIDs(t *testing.T) {
	unknown := `{"version":"v1","taxonomy_version":"cankonix-it-v1","review_status":"candidate","cases":[],"extra":true}`
	if _, err := Load(strings.NewReader(unknown)); err == nil {
		t.Fatal("Load() accepted an unknown field")
	}
	duplicate := `{"version":"v1","taxonomy_version":"cankonix-it-v1","review_status":"candidate","cases":[{"id":"same","title":"one","expected_disposition":"excluded"},{"id":"same","title":"two","expected_disposition":"excluded"}]}`
	if _, err := Load(strings.NewReader(duplicate)); err == nil {
		t.Fatal("Load() accepted duplicate IDs")
	}
}
