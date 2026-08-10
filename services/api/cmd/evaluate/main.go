package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/cankonix/cankonix-oi/services/api/internal/classification"
	"github.com/cankonix/cankonix-oi/services/api/internal/evaluation"
)

func main() {
	datasetPath := flag.String("dataset", "testdata/classification_golden_v1.json", "path to the versioned golden dataset")
	minimumPrecision := flag.Int("minimum-precision-bp", 8000, "minimum macro precision in basis points")
	requireApproved := flag.Bool("require-approved", false, "fail unless the dataset review status is approved")
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
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fail(fmt.Sprintf("encode report: %v", err))
	}
	if report.MacroPrecisionBP < *minimumPrecision {
		os.Exit(1)
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
