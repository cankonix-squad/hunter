// Package classificationv2 is a review-only classifier. The collector still uses v1.
package classificationv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"
)

type ExclusionRule struct {
	Phrase string `json:"phrase"`
	Reason string `json:"reason"`
}

type Category struct {
	ID                string              `json:"category_id"`
	Name              string              `json:"name"`
	Description       string              `json:"business_description"`
	Priority          string              `json:"priority"`
	DiscoveryKeywords []string            `json:"discovery_keywords"`
	PositiveKeywords  []string            `json:"positive_keywords"`
	StrongPhrases     []string            `json:"strong_phrases"`
	Synonyms          map[string][]string `json:"synonyms"`
	ContextKeywords   []string            `json:"context_keywords"`
	NegativeKeywords  []string            `json:"negative_keywords"`
	ExclusionRules    []ExclusionRule     `json:"exclusion_rules"`
}

type Config struct {
	Version                string     `json:"version"`
	Categories             []Category `json:"categories"`
	GlobalExclusionPhrases []string   `json:"global_exclusion_phrases"`
	AmbiguousPhrases       []string   `json:"ambiguous_phrases"`
}

type Input struct {
	Title           string `json:"title"`
	Description     string `json:"description,omitempty"`
	TOR             string `json:"tor,omitempty"`
	Requirements    string `json:"requirements,omitempty"`
	ProcurementType string `json:"procurement_type,omitempty"`
}

type Evidence struct {
	CategoryID string `json:"category_id,omitempty"`
	Kind       string `json:"kind"`
	Phrase     string `json:"phrase"`
	Field      string `json:"field"`
}

type Result struct {
	TaxonomyVersion     string     `json:"taxonomy_version"`
	Disposition         string     `json:"disposition"`
	PrimaryCategory     string     `json:"primary_category,omitempty"`
	SecondaryCategories []string   `json:"secondary_categories,omitempty"`
	ReasonCode          string     `json:"reason_code"`
	Explanation         string     `json:"explanation"`
	Evidence            []Evidence `json:"evidence"`
	AutoRejected        bool       `json:"auto_rejected"`
}

type Engine struct{ config Config }

func Load(reader io.Reader) (Engine, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var config Config
	if err := decoder.Decode(&config); err != nil {
		return Engine{}, fmt.Errorf("decode taxonomy: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return Engine{}, errors.New("taxonomy must contain one object")
	}
	if config.Version == "" || len(config.Categories) != 12 {
		return Engine{}, errors.New("taxonomy needs a version and twelve categories")
	}
	seen := map[string]bool{}
	for _, category := range config.Categories {
		if category.ID == "" || seen[category.ID] || category.Name == "" || category.Description == "" ||
			(category.Priority != "P1" && category.Priority != "P2") || len(category.DiscoveryKeywords) == 0 || len(category.PositiveKeywords) == 0 ||
			len(category.ContextKeywords) == 0 || len(category.NegativeKeywords) == 0 || len(category.ExclusionRules) == 0 {
			return Engine{}, fmt.Errorf("incomplete or duplicate category %q", category.ID)
		}
		seen[category.ID] = true
	}
	return Engine{config: config}, nil
}

func Normalize(s string) string {
	var out strings.Builder
	space := true
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			out.WriteRune(r)
			space = false
		} else if !space {
			out.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(out.String())
}

func match(value, phrase string) bool {
	value, phrase = Normalize(value), Normalize(phrase)
	return phrase != "" && strings.Contains(" "+value+" ", " "+phrase+" ")
}

type field struct{ name, value string }

func find(fields []field, phrases []string, kind, category string) []Evidence {
	var found []Evidence
	for _, phrase := range phrases {
		for _, f := range fields {
			if match(f.value, phrase) {
				found = append(found, Evidence{CategoryID: category, Kind: kind, Phrase: phrase, Field: f.name})
				break
			}
		}
	}
	return found
}

type candidate struct {
	category      Category
	strength      int
	titlePosition int
}

func contextualPair(positive, context []Evidence) bool {
	for _, p := range positive {
		for _, c := range context {
			if p.Field == c.Field && !strings.Contains(Normalize(p.Phrase), Normalize(c.Phrase)) && !strings.Contains(Normalize(c.Phrase), Normalize(p.Phrase)) {
				return true
			}
		}
	}
	return false
}

func (e Engine) Evaluate(input Input) Result {
	fields := []field{{"title", input.Title}, {"description", input.Description}, {"tor", input.TOR}, {"requirements", input.Requirements}}
	result := Result{TaxonomyVersion: e.config.Version, Disposition: "needs_review", ReasonCode: "insufficient_context", Explanation: "Bukti pekerjaan belum cukup untuk memutuskan relevansi.", Evidence: []Evidence{}}
	var candidates []candidate
	for _, category := range e.config.Categories {
		result.Evidence = append(result.Evidence, find(fields, category.DiscoveryKeywords, "discovery_only", category.ID)...)
		terms := append([]string{}, category.PositiveKeywords...)
		for _, aliases := range category.Synonyms {
			terms = append(terms, aliases...)
		}
		positive := find(fields, terms, "positive", category.ID)
		strong := find(fields, category.StrongPhrases, "exact_phrase", category.ID)
		context := find(fields, category.ContextKeywords, "context", category.ID)
		negative := find(fields, category.NegativeKeywords, "negative", category.ID)
		for _, rule := range category.ExclusionRules {
			negative = append(negative, find(fields, []string{rule.Phrase}, "exclusion_rule", category.ID)...)
		}
		if len(positive)+len(strong) == 0 {
			continue
		}
		result.Evidence = append(result.Evidence, positive...)
		result.Evidence = append(result.Evidence, strong...)
		result.Evidence = append(result.Evidence, context...)
		result.Evidence = append(result.Evidence, negative...)
		if len(negative) > 0 {
			continue
		}
		if len(strong) > 0 || contextualPair(positive, context) {
			strength := len(strong)*4 + len(positive)*2 + len(context)
			if category.Priority == "P1" {
				strength++
			}
			position := len(Normalize(input.Title)) + 1
			for _, item := range strong {
				if item.Field == "title" {
					if at := strings.Index(Normalize(input.Title), Normalize(item.Phrase)); at >= 0 && at < position {
						position = at
					}
				}
			}
			candidates = append(candidates, candidate{category, strength, position})
		}
	}
	global := find(fields, e.config.GlobalExclusionPhrases, "global_exclusion", "")
	ambiguous := find(fields, e.config.AmbiguousPhrases, "ambiguity", "")
	result.Evidence = append(result.Evidence, global...)
	result.Evidence = append(result.Evidence, ambiguous...)
	if len(global) > 0 && len(candidates) > 0 {
		result.ReasonCode, result.Explanation = "mixed_scope", "Lingkup TI dan non-TI bercampur; komposisi pekerjaan perlu ditinjau."
		return result
	}
	if len(global) > 0 {
		result.Disposition, result.ReasonCode, result.Explanation = "excluded", "clear_non_it_scope", "Judul/uraian secara jelas hanya memuat lingkup di luar bisnis CANKONIX."
		return result
	}
	if len(candidates) == 0 {
		return result
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].titlePosition != candidates[j].titlePosition {
			return candidates[i].titlePosition < candidates[j].titlePosition
		}
		if candidates[i].strength != candidates[j].strength {
			return candidates[i].strength > candidates[j].strength
		}
		return candidates[i].category.ID < candidates[j].category.ID
	})
	result.Disposition, result.PrimaryCategory = "relevant", candidates[0].category.ID
	result.ReasonCode, result.Explanation = "contextual_fit", "Frasa pekerjaan dan konteks mendukung kapabilitas CANKONIX."
	for _, secondary := range candidates[1:] {
		result.SecondaryCategories = append(result.SecondaryCategories, secondary.category.ID)
	}
	return result
}
