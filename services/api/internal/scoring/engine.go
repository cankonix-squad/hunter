package scoring

import (
	"fmt"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/classification"
)

const Version = "explainable-v1"

type Weight struct {
	Dimension string
	WeightBP  int
}

var Weights = []Weight{
	{Dimension: "capability_fit", WeightBP: 4000},
	{Dimension: "value", WeightBP: 2000},
	{Dimension: "location", WeightBP: 1000},
	{Dimension: "deadline", WeightBP: 1500},
	{Dimension: "risk", WeightBP: 1500},
}

type Input struct {
	Classification       classification.Result
	BudgetMinor          *int64
	Currency             string
	Location             string
	LifecycleStatus      string
	RegistrationDeadline *time.Time
	EvaluatedAt          time.Time
}

type Dimension struct {
	Code         string `json:"code"`
	WeightBP     int    `json:"weight_bp"`
	RawScoreBP   int    `json:"raw_score_bp"`
	Contribution int    `json:"contribution"`
	ReasonCode   string `json:"reason_code"`
	Explanation  string `json:"explanation"`
	Evidence     string `json:"evidence"`
}

type Result struct {
	Version      string      `json:"version"`
	TotalScoreBP int         `json:"total_score_bp"`
	ConfidenceBP int         `json:"confidence_bp"`
	NeedsReview  bool        `json:"needs_review"`
	ReasonCode   string      `json:"reason_code"`
	Explanation  string      `json:"explanation"`
	EvaluatedAt  time.Time   `json:"evaluated_at"`
	Dimensions   []Dimension `json:"dimensions"`
}

func Evaluate(input Input) Result {
	evaluatedAt := input.EvaluatedAt.UTC()
	if evaluatedAt.IsZero() {
		evaluatedAt = time.Now().UTC()
	}
	dimensions := []Dimension{
		capability(input.Classification),
		value(input.BudgetMinor, input.Currency),
		location(input.Location),
		deadline(input.RegistrationDeadline, evaluatedAt),
		risk(input),
	}
	total := 0
	known := 0
	for i := range dimensions {
		dimensions[i].WeightBP = Weights[i].WeightBP
		dimensions[i].Contribution = dimensions[i].RawScoreBP * dimensions[i].WeightBP / 10000
		total += dimensions[i].Contribution
		if dimensions[i].ReasonCode != "unknown" {
			known += dimensions[i].WeightBP
		}
	}
	confidence := (input.Classification.ConfidenceBP*6000/10000 + known*4000/10000)
	needsReview := input.Classification.Disposition == "needs_review" || confidence < 7000
	reason, explanation := "score_ready", "Score deterministik lengkap dan dapat digunakan sebagai dukungan review."
	if needsReview {
		reason, explanation = "low_confidence", "Data atau classification belum cukup kuat; opportunity perlu review manusia."
	}
	return Result{Version: Version, TotalScoreBP: total, ConfidenceBP: confidence, NeedsReview: needsReview,
		ReasonCode: reason, Explanation: explanation, EvaluatedAt: evaluatedAt, Dimensions: dimensions}
}

func capability(c classification.Result) Dimension {
	switch c.Disposition {
	case "relevant":
		return Dimension{Code: "capability_fit", RawScoreBP: 9000, ReasonCode: "deterministic_fit", Explanation: "Rule deterministik menemukan kapabilitas CANKONIX yang sesuai.", Evidence: c.ReasonCode}
	case "excluded":
		return Dimension{Code: "capability_fit", RawScoreBP: 0, ReasonCode: "deterministic_exclusion", Explanation: "Rule deterministik mengecualikan paket dari kapabilitas inti.", Evidence: c.ReasonCode}
	default:
		return Dimension{Code: "capability_fit", RawScoreBP: 4000, ReasonCode: "classification_ambiguous", Explanation: "Capability fit belum dapat ditentukan secara deterministik.", Evidence: c.ReasonCode}
	}
}

func value(budget *int64, currency string) Dimension {
	if budget == nil || currency != "IDR" {
		return Dimension{Code: "value", RawScoreBP: 5000, ReasonCode: "unknown", Explanation: "Nilai IDR tidak tersedia; digunakan nilai netral.", Evidence: "budget unavailable"}
	}
	// Values are integer minor units. IDR uses two configured decimal places here.
	switch {
	case *budget >= 100_000_000_000:
		return Dimension{Code: "value", RawScoreBP: 9000, ReasonCode: "high_value", Explanation: "Nilai paket berada pada band prioritas tinggi.", Evidence: fmt.Sprintf("budget_minor=%d %s", *budget, currency)}
	case *budget >= 25_000_000_000:
		return Dimension{Code: "value", RawScoreBP: 7500, ReasonCode: "target_value", Explanation: "Nilai paket berada pada band target.", Evidence: fmt.Sprintf("budget_minor=%d %s", *budget, currency)}
	default:
		return Dimension{Code: "value", RawScoreBP: 4000, ReasonCode: "low_value", Explanation: "Nilai paket berada di bawah band target.", Evidence: fmt.Sprintf("budget_minor=%d %s", *budget, currency)}
	}
}

func location(place string) Dimension {
	if place == "" {
		return Dimension{Code: "location", RawScoreBP: 5000, ReasonCode: "unknown", Explanation: "Lokasi belum tersedia; digunakan nilai netral.", Evidence: "location unavailable"}
	}
	return Dimension{Code: "location", RawScoreBP: 7000, ReasonCode: "location_available", Explanation: "Lokasi tersedia untuk review operasional.", Evidence: place}
}

func deadline(value *time.Time, now time.Time) Dimension {
	if value == nil {
		return Dimension{Code: "deadline", RawScoreBP: 3000, ReasonCode: "unknown", Explanation: "Deadline registrasi tidak tersedia.", Evidence: "deadline unavailable"}
	}
	days := int(value.Sub(now).Hours() / 24)
	switch {
	case days < 0:
		return Dimension{Code: "deadline", RawScoreBP: 0, ReasonCode: "deadline_passed", Explanation: "Deadline registrasi telah lewat.", Evidence: value.UTC().Format(time.RFC3339)}
	case days < 3:
		return Dimension{Code: "deadline", RawScoreBP: 2500, ReasonCode: "deadline_critical", Explanation: "Waktu respons kurang dari tiga hari.", Evidence: value.UTC().Format(time.RFC3339)}
	case days <= 14:
		return Dimension{Code: "deadline", RawScoreBP: 8000, ReasonCode: "deadline_actionable", Explanation: "Waktu respons berada pada rentang yang dapat ditindaklanjuti.", Evidence: value.UTC().Format(time.RFC3339)}
	default:
		return Dimension{Code: "deadline", RawScoreBP: 6500, ReasonCode: "deadline_long", Explanation: "Deadline masih panjang.", Evidence: value.UTC().Format(time.RFC3339)}
	}
}

func risk(input Input) Dimension {
	if input.Classification.Disposition == "excluded" {
		return Dimension{Code: "risk", RawScoreBP: 0, ReasonCode: "excluded_risk", Explanation: "Deterministic exclusion membuat risiko pursuit sangat tinggi.", Evidence: input.Classification.ReasonCode}
	}
	if input.Classification.Disposition == "needs_review" {
		return Dimension{Code: "risk", RawScoreBP: 3500, ReasonCode: "classification_risk", Explanation: "Classification ambigu meningkatkan risiko keputusan.", Evidence: input.Classification.ReasonCode}
	}
	return Dimension{Code: "risk", RawScoreBP: 8000, ReasonCode: "baseline_risk", Explanation: "Tidak ada deterministic exclusion atau konflik classification.", Evidence: input.LifecycleStatus}
}
