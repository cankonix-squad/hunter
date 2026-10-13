package classificationv2

import "strings"

type ScoreDimension struct {
	Code     string `json:"code"`
	Weight   int    `json:"weight_percent"`
	Value    *int   `json:"value,omitempty"`
	Status   string `json:"status"`
	Evidence string `json:"evidence,omitempty"`
}

type EligibilityFlag struct {
	Requirement string `json:"requirement"`
	Status      string `json:"status"` // met, unmet, or unknown
	Evidence    string `json:"evidence,omitempty"`
}

type ScoreResult struct {
	Version           string            `json:"version"`
	Score             *int              `json:"score,omitempty"`
	PriorityBand      string            `json:"priority_band,omitempty"`
	Provisional       bool              `json:"provisional"`
	DataCompleteness  float64           `json:"data_completeness"`
	EligibilityStatus string            `json:"eligibility_status"`
	EligibilityFlags  []EligibilityFlag `json:"eligibility_flags"`
	Dimensions        []ScoreDimension  `json:"dimensions"`
}

var scoreWeights = []struct {
	code   string
	weight int
}{
	{"capability_fit", 35}, {"scope_fit", 25}, {"supplier_qualification", 20},
	{"value_margin", 10}, {"delivery_readiness", 10},
}

// Score keeps missing dimensions unknown. It never converts absent data to zero.
// The numeric score is normalized over known dimensions, with completeness separate.
func Score(values map[string]*int, evidence map[string]string, flags []EligibilityFlag) ScoreResult {
	result := ScoreResult{Version: "opportunity-v2-candidate", EligibilityStatus: "unknown", EligibilityFlags: flags, Dimensions: make([]ScoreDimension, 0, 5)}
	weightedSum, knownWeight := 0, 0
	for _, def := range scoreWeights {
		d := ScoreDimension{Code: def.code, Weight: def.weight, Status: "unknown"}
		if v := values[def.code]; v != nil && *v >= 0 && *v <= 100 && strings.TrimSpace(evidence[def.code]) != "" {
			d.Value, d.Status, d.Evidence = v, "known", evidence[def.code]
			weightedSum += *v * def.weight
			knownWeight += def.weight
		}
		result.Dimensions = append(result.Dimensions, d)
	}
	result.DataCompleteness = float64(knownWeight) / 100
	result.Provisional = knownWeight < 100
	if knownWeight > 0 {
		score := (weightedSum + knownWeight/2) / knownWeight
		result.Score = &score
		switch {
		case score >= 80:
			result.PriorityBand = "high"
		case score >= 60:
			result.PriorityBand = "medium"
		default:
			result.PriorityBand = "low"
		}
	}
	met, unknown := 0, 0
	for _, flag := range flags {
		if strings.TrimSpace(flag.Evidence) == "" {
			unknown++
			continue
		}
		switch flag.Status {
		case "unmet":
			result.EligibilityStatus = "ineligible"
			return result
		case "met":
			met++
		default:
			unknown++
		}
	}
	if len(flags) > 0 && unknown == 0 && met == len(flags) {
		result.EligibilityStatus = "eligible"
	}
	return result
}
