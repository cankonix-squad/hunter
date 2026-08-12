package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cankonix/cankonix-oi/services/api/internal/classification"
	"github.com/cankonix/cankonix-oi/services/api/internal/evaluation"
)

func main() {
	datasetPath := flag.String("dataset", "testdata/classification_golden_v1.json", "path to the versioned golden dataset")
	minimumPrecision := flag.Int("minimum-precision-bp", 8000, "minimum macro precision in basis points")
	requireApproved := flag.Bool("require-approved", false, "fail unless the dataset review status is approved")
	outputFormat := flag.String("format", "json", "output format: json or review")
	flag.Parse()
	if *minimumPrecision < 0 || *minimumPrecision > 10000 {
		fail("minimum precision must be between 0 and 10000")
	}
	file, err := os.Open(*datasetPath)
	if err != nil {
		fail(fmt.Sprintf("open golden dataset: %v", err))
	}
	defer file.Close()
	dataset, err := evaluation.Load(file)
	if err != nil {
		fail(err.Error())
	}
	if *requireApproved && dataset.ReviewStatus != "approved" {
		fail(fmt.Sprintf("golden dataset is %q, not approved", dataset.ReviewStatus))
	}
	report, err := evaluation.Evaluate(dataset, classification.DefaultEngine())
	if err != nil {
		fail(err.Error())
	}
	switch *outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fail(fmt.Sprintf("encode report: %v", err))
		}
	case "review":
		writeReview(dataset, report)
	default:
		fail("format must be json or review")
	}
	if report.MacroPrecisionBP < *minimumPrecision {
		os.Exit(1)
	}
}

func writeReview(dataset evaluation.Dataset, report evaluation.Report) {
	fmt.Printf("# Classification golden dataset review — %s\n\n", dataset.Version)
	fmt.Printf("Taxonomy: `%s` · Rules: `%s` · Candidate macro precision: %.2f%%\n\n", dataset.TaxonomyVersion, report.RuleVersion, float64(report.MacroPrecisionBP)/100)
	fmt.Println("| ID | Title | Procurement type | Expected disposition | Expected category | Domain owner decision |")
	fmt.Println("|---|---|---|---|---|---|")
	for _, item := range dataset.Cases {
		fmt.Printf("| %s | %s | %s | %s | %s | APPROVE / CHANGE |\n", reviewCell(item.ID), reviewCell(item.Title), reviewCell(item.ProcurementType), reviewCell(item.ExpectedDisposition), reviewCell(item.ExpectedCategory))
	}
	fmt.Println("\nApproval requires every row reviewed, coverage comments, reviewer identity, UTC timestamp, and approval reference. Do not place credentials or personal data in this packet.")
}

func reviewCell(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	return strings.Join(strings.Fields(value), " ")
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
