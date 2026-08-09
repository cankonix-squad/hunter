package scoring

import (
	"testing"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/classification"
)

func TestScoreRegressionRelevantFixture(t *testing.T) {
	budget := int64(75_000_000_000)
	deadline := time.Date(2026, 8, 20, 9, 0, 0, 0, time.UTC)
	result := Evaluate(Input{Classification: classification.DefaultEngine().Evaluate(classification.Input{Title: "Pengadaan Layanan Integrasi Data"}),
		BudgetMinor: &budget, Currency: "IDR", LifecycleStatus: "registration_open", RegistrationDeadline: &deadline,
		EvaluatedAt: time.Date(2026, 8, 9, 2, 0, 0, 0, time.UTC)})
	if result.TotalScoreBP != 8000 || result.ConfidenceBP != 9000 || result.NeedsReview {
		t.Fatalf("scoring regression: %#v", result)
	}
	if len(result.Dimensions) != 5 {
		t.Fatalf("got %d dimensions", len(result.Dimensions))
	}
}

func TestScoreMarksAmbiguousClassificationForReview(t *testing.T) {
	result := Evaluate(Input{Classification: classification.DefaultEngine().Evaluate(classification.Input{Title: "Transformasi Organisasi"}), EvaluatedAt: time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC)})
	if !result.NeedsReview || result.ReasonCode != "low_confidence" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestWeightsAreVersionedAndSumToOneHundredPercent(t *testing.T) {
	total := 0
	for _, weight := range Weights {
		total += weight.WeightBP
	}
	if Version != "explainable-v1" || total != 10000 {
		t.Fatalf("version=%q total=%d", Version, total)
	}
}
