package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/cankonix/cankonix-oi/services/api/internal/classificationv2"
)

type testCase struct {
	ID                          string   `json:"id"`
	Title                       string   `json:"title"`
	Description                 string   `json:"description"`
	TOR                         string   `json:"tor"`
	Requirements                string   `json:"requirements"`
	ProcurementType             string   `json:"procurement_type"`
	ExpectedDisposition         string   `json:"expected_disposition"`
	ExpectedPrimaryCategory     string   `json:"expected_primary_category"`
	ExpectedSecondaryCategories []string `json:"expected_secondary_categories"`
}
type dataset struct {
	Version         string     `json:"version"`
	TaxonomyVersion string     `json:"taxonomy_version"`
	ReviewStatus    string     `json:"review_status"`
	Provenance      string     `json:"provenance"`
	Cases           []testCase `json:"cases"`
}
type failure struct{ ID, Expected, Actual, CategoryExpected, CategoryActual string }

func main() {
	configPath := flag.String("config", "../../docs/classification-v2/classification-keywords-v2.json", "taxonomy configuration")
	datasetPath := flag.String("dataset", "../../docs/classification-v2/classification-testcases-v2.json", "test cases")
	flag.Parse()
	configBytes, err := os.ReadFile(*configPath)
	must(err)
	engine, err := classificationv2.Load(bytes.NewReader(configBytes))
	must(err)
	var config classificationv2.Config
	must(json.Unmarshal(configBytes, &config))
	discoveryTerms := 0
	for _, category := range config.Categories {
		discoveryTerms += len(category.DiscoveryKeywords)
	}
	dataBytes, err := os.ReadFile(*datasetPath)
	must(err)
	var data dataset
	must(json.Unmarshal(dataBytes, &data))
	if data.Version == "" || data.TaxonomyVersion != "cankonix-business-v2-candidate" || data.ReviewStatus != "candidate" || len(data.Cases) != 120 {
		panic("invalid v2 dataset metadata or count")
	}
	labels := []string{"relevant", "excluded", "needs_review"}
	confusion := map[string]map[string]int{}
	for _, label := range labels {
		confusion[label] = map[string]int{}
	}
	failures := []failure{}
	correct, categoryCorrect := 0, 0
	for _, tc := range data.Cases {
		got := engine.Evaluate(classificationv2.Input{Title: tc.Title, Description: tc.Description, TOR: tc.TOR, Requirements: tc.Requirements, ProcurementType: tc.ProcurementType})
		confusion[tc.ExpectedDisposition][got.Disposition]++
		if got.Disposition == tc.ExpectedDisposition {
			correct++
		}
		secondary := append([]string{}, got.SecondaryCategories...)
		expectedSecondary := append([]string{}, tc.ExpectedSecondaryCategories...)
		sort.Strings(secondary)
		sort.Strings(expectedSecondary)
		categoryMatch := got.PrimaryCategory == tc.ExpectedPrimaryCategory && strings.Join(secondary, ",") == strings.Join(expectedSecondary, ",")
		if got.Disposition == tc.ExpectedDisposition && categoryMatch {
			categoryCorrect++
		} else {
			failures = append(failures, failure{tc.ID, tc.ExpectedDisposition, got.Disposition, tc.ExpectedPrimaryCategory + "/" + strings.Join(expectedSecondary, ","), got.PrimaryCategory + "/" + strings.Join(secondary, ",")})
		}
	}
	fmt.Println("# Classification evaluation v2")
	fmt.Printf("\nDataset: `%s` (%s). Provenance: %s.\n", data.Version, data.ReviewStatus, data.Provenance)
	fmt.Printf("\nOwner master terms: %d discovery phrases across %d categories. Discovery-only matches do not decide relevance.\n", discoveryTerms, len(config.Categories))
	fmt.Printf("\nCases: %d. Label accuracy: %.2f%%. Label plus category exact accuracy: %.2f%%.\n", len(data.Cases), 100*float64(correct)/float64(len(data.Cases)), 100*float64(categoryCorrect)/float64(len(data.Cases)))
	fmt.Print("\n## Confusion matrix\n\nRows are expected; columns are predicted.\n\n")
	fmt.Println("| Expected \\ predicted | relevant | excluded | needs_review |\n|---|---:|---:|---:|")
	for _, expected := range labels {
		fmt.Printf("| %s | %d | %d | %d |\n", expected, confusion[expected]["relevant"], confusion[expected]["excluded"], confusion[expected]["needs_review"])
	}
	fmt.Println("\n## Precision and recall\n\n| Label | Precision | Recall | TP | FP | FN |\n|---|---:|---:|---:|---:|---:|")
	for _, label := range labels {
		tp, fp, fn := confusion[label][label], 0, 0
		for _, other := range labels {
			if other != label {
				fp += confusion[other][label]
				fn += confusion[label][other]
			}
		}
		fmt.Printf("| %s | %.2f%% | %.2f%% | %d | %d | %d |\n", label, percentage(tp, tp+fp), percentage(tp, tp+fn), tp, fp, fn)
	}
	fmt.Print("\n## Label or category mismatches\n\n")
	if len(failures) == 0 {
		fmt.Println("No mismatches on this synthetic candidate set.")
	} else {
		fmt.Println("| Case | Expected label | Actual label | Expected primary/secondary | Actual primary/secondary |\n|---|---|---|---|---|")
		for _, f := range failures {
			fmt.Printf("| %s | %s | %s | %s | %s |\n", f.ID, f.Expected, f.Actual, f.CategoryExpected, f.CategoryActual)
		}
	}
	fmt.Print("\n## False positives and false negatives\n\n")
	fmt.Println("A false positive predicts `relevant` for an expected `excluded` or `needs_review` case. A false negative expects `relevant` but predicts another label. Category mismatches are listed above separately.")
	fmt.Println("\n| Kind | Case IDs |\n|---|---|")
	for _, kind := range []string{"false positive", "false negative"} {
		var ids []string
		for _, f := range failures {
			if kind == "false positive" && f.Actual == "relevant" && f.Expected != "relevant" || kind == "false negative" && f.Expected == "relevant" && f.Actual != "relevant" {
				ids = append(ids, f.ID)
			}
		}
		if len(ids) == 0 {
			ids = []string{"none"}
		}
		fmt.Printf("| %s | %s |\n", kind, strings.Join(ids, ", "))
	}
	fmt.Print("\n## Production interpretation\n\n")
	fmt.Println("These cases are synthetic and authored alongside the rules. Their metrics are regression feedback, not production readiness evidence. An independent reviewer must label a real procurement sample, including hard negatives and mixed scopes, before release. The collector and production classifier remain on v1.")
	fmt.Print("\n## Coverage and next validation\n\n")
	fmt.Println("The dataset retains 15 owner-approved v1 labels as a temporary baseline and covers every v2 category with positive, negative, and ambiguous cases. It includes keyword collisions, multiple categories, Barang IT work, KAK/TOR context, specialist training, and common abbreviation/typo variants. No real LPSE package has been independently labeled for this v2 report.")
	fmt.Print("\nBefore production integration, collect a time-separated sample of real packages from the 20 priority sources; have a reviewer label title, description, TOR, eligibility, and primary/secondary categories without seeing model output. Report precision, recall, confusion matrix, false positives, false negatives, source coverage, and confidence intervals on that held-out sample. Resolve owner corrections, then rerun this evaluator and the Go tests on an immutable taxonomy version.\n")
	fmt.Print("\n## Scoring contract\n\n")
	fmt.Println("The v2 candidate score uses 35/25/20/10/10 weights for capability, scope, supplier qualification, value/margin, and delivery readiness. A numeric dimension requires evidence; otherwise it remains unknown. Unknown dimensions are omitted from the weighted numerator and denominator; data completeness is reported separately and is not a calibrated confidence probability. An incomplete score is `provisional=true`. A verified unmet mandatory requirement sets `eligibility_status=ineligible` without rewriting the score. Scoring is not wired into the production collector.")
	fit, scope, qualification, value, readiness := 90, 80, 70, 60, 50
	unknown := classificationv2.Score(nil, nil, nil)
	partial := classificationv2.Score(map[string]*int{"capability_fit": &fit}, map[string]string{"capability_fit": "reviewed fit"}, nil)
	complete := classificationv2.Score(map[string]*int{"capability_fit": &fit, "scope_fit": &scope, "supplier_qualification": &qualification, "value_margin": &value, "delivery_readiness": &readiness}, map[string]string{"capability_fit": "reviewed fit", "scope_fit": "reviewed TOR", "supplier_qualification": "verified registry", "value_margin": "reviewed budget and margin", "delivery_readiness": "reviewed schedule"}, nil)
	blocked := classificationv2.Score(map[string]*int{"capability_fit": &fit}, map[string]string{"capability_fit": "reviewed fit"}, []classificationv2.EligibilityFlag{{Requirement: "mandatory certificate", Status: "unmet", Evidence: "verified registry"}})
	fmt.Print("\n| Check | Score | Band | Data completeness | Provisional | Eligibility |\n|---|---:|---|---:|---|---|\n")
	fmt.Printf("| No verified dimensions | %s | — | %.0f%% | %t | %s |\n", showScore(unknown.Score), unknown.DataCompleteness*100, unknown.Provisional, unknown.EligibilityStatus)
	fmt.Printf("| Capability only (90) | %s | %s | %.0f%% | %t | %s |\n", showScore(partial.Score), partial.PriorityBand, partial.DataCompleteness*100, partial.Provisional, partial.EligibilityStatus)
	fmt.Printf("| Five verified dimensions | %s | %s | %.0f%% | %t | %s |\n", showScore(complete.Score), complete.PriorityBand, complete.DataCompleteness*100, complete.Provisional, complete.EligibilityStatus)
	fmt.Printf("| Verified mandatory failure | %s | %s | %.0f%% | %t | %s |\n", showScore(blocked.Score), blocked.PriorityBand, blocked.DataCompleteness*100, blocked.Provisional, blocked.EligibilityStatus)
}

func percentage(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return 100 * float64(n) / float64(d)
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}

func showScore(value *int) string {
	if value == nil {
		return "unknown"
	}
	return fmt.Sprint(*value)
}
