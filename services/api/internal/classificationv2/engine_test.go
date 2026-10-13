package classificationv2

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func docsPath(name string) string {
	return filepath.Join("..", "..", "..", "..", "docs", "classification-v2", name)
}

func fixtureEngine(t *testing.T) Engine {
	t.Helper()
	f, err := os.Open(docsPath("classification-keywords-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	engine, err := Load(f)
	if err != nil {
		t.Fatal(err)
	}
	return engine
}

func TestCandidateDataset(t *testing.T) {
	data, err := os.ReadFile(docsPath("classification-testcases-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		ReviewStatus string `json:"review_status"`
		Cases        []struct {
			ID string `json:"id"`
			Input
			ExpectedDisposition string   `json:"expected_disposition"`
			ExpectedPrimary     string   `json:"expected_primary_category"`
			ExpectedSecondary   []string `json:"expected_secondary_categories"`
			CategoryUnderTest   string   `json:"category_under_test"`
			Tags                []string `json:"tags"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.ReviewStatus != "candidate" || len(fixture.Cases) != 120 {
		t.Fatalf("metadata: status=%s count=%d", fixture.ReviewStatus, len(fixture.Cases))
	}
	engine := fixtureEngine(t)
	counts := map[string]int{}
	baseline := 0
	seen := map[string]bool{}
	coverage := map[string]map[string]bool{}
	for _, tc := range fixture.Cases {
		if seen[tc.ID] {
			t.Fatalf("duplicate %s", tc.ID)
		}
		seen[tc.ID] = true
		counts[tc.ExpectedDisposition]++
		category := tc.ExpectedPrimary
		if category == "" {
			category = tc.CategoryUnderTest
		}
		if category != "" {
			if coverage[category] == nil {
				coverage[category] = map[string]bool{}
			}
			coverage[category][tc.ExpectedDisposition] = true
		}
		for _, tag := range tc.Tags {
			if tag == "v1_baseline" {
				baseline++
			}
		}
		got := engine.Evaluate(tc.Input)
		secondary := append([]string{}, got.SecondaryCategories...)
		expected := append([]string{}, tc.ExpectedSecondary...)
		sort.Strings(secondary)
		sort.Strings(expected)
		if got.Disposition != tc.ExpectedDisposition || got.PrimaryCategory != tc.ExpectedPrimary || strings.Join(secondary, ",") != strings.Join(expected, ",") {
			t.Errorf("%s: got %s %s %v, want %s %s %v", tc.ID, got.Disposition, got.PrimaryCategory, secondary, tc.ExpectedDisposition, tc.ExpectedPrimary, expected)
		}
		if got.AutoRejected {
			t.Errorf("v2 candidate must not auto-reject %s", tc.ID)
		}
	}
	if baseline != 15 || counts["relevant"] != 50 || counts["excluded"] != 40 || counts["needs_review"] != 30 {
		t.Fatalf("distribution=%v baseline=%d", counts, baseline)
	}
	if len(coverage) != 12 {
		t.Fatalf("category coverage=%d", len(coverage))
	}
	for category, labels := range coverage {
		if !labels["relevant"] || !labels["excluded"] || !labels["needs_review"] {
			t.Errorf("missing label coverage for %s: %v", category, labels)
		}
	}
}

func TestContextAndDocumentEvidence(t *testing.T) {
	engine := fixtureEngine(t)
	bare := engine.Evaluate(Input{Title: "Pengadaan perangkat jaringan", ProcurementType: "Barang"})
	if bare.Disposition != "needs_review" {
		t.Fatalf("bare hardware: %#v", bare)
	}
	withTOR := engine.Evaluate(Input{Title: "Pengadaan perangkat jaringan", TOR: "Implementasi jaringan dan konfigurasi jaringan cabang", ProcurementType: "Barang"})
	if withTOR.Disposition != "relevant" || withTOR.PrimaryCategory != "network-infrastructure" {
		t.Fatalf("TOR context: %#v", withTOR)
	}
	foundTOR := false
	for _, item := range withTOR.Evidence {
		if item.Field == "tor" {
			foundTOR = true
		}
	}
	if !foundTOR {
		t.Fatal("missing TOR evidence")
	}
	mixed := engine.Evaluate(Input{Title: "Pengembangan aplikasi dan pengadaan alat tulis kantor"})
	if mixed.Disposition != "needs_review" || mixed.ReasonCode != "mixed_scope" {
		t.Fatalf("mixed scope: %#v", mixed)
	}
	typo := engine.Evaluate(Input{Title: "Pengembangan apliksi khusus pelayanan warga"})
	if typo.Disposition != "relevant" || typo.PrimaryCategory != "software-engineering" {
		t.Fatalf("typo: %#v", typo)
	}
}

func TestMasterTermsDoNotMakeGenericWorkRelevant(t *testing.T) {
	engine := fixtureEngine(t)
	tests := []struct{ title, want, category string }{
		{"dashboard", "needs_review", ""},
		{"dashboard monitoring rapat", "needs_review", ""},
		{"aplikasi", "needs_review", ""},
		{"konsultansi", "needs_review", ""},
		{"ISO 27001", "needs_review", ""},
		{"Pengadaan lisensi aplikasi umum", "needs_review", ""},
		{"Pemeliharaan gedung", "excluded", ""},
		{"Pemeliharaan sistem informasi", "relevant", "managed-it-services"},
		{"Monitoring server", "relevant", "managed-it-services"},
		{"Implementasi SIEM", "relevant", "cybersecurity"},
		{"Audit ISO 27001", "relevant", "it-grc"},
		{"Implementasi SSO", "relevant", "system-integration"},
		{"IT master plan", "relevant", "it-consulting"},
		{"Dashboard eksekutif", "needs_review", ""},
		{"Pengembangan dashboard eksekutif analitik", "relevant", "data-ai"},
	}
	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			got := engine.Evaluate(Input{Title: tc.title})
			if got.Disposition != tc.want || got.PrimaryCategory != tc.category {
				t.Errorf("got %s/%s, want %s/%s", got.Disposition, got.PrimaryCategory, tc.want, tc.category)
			}
		})
	}
}

func TestScoreUnknownAndEligibility(t *testing.T) {
	empty := Score(map[string]*int{}, nil, nil)
	if empty.Score != nil || empty.DataCompleteness != 0 || empty.PriorityBand != "" {
		t.Fatalf("empty score: %#v", empty)
	}
	fit := 90
	partial := Score(map[string]*int{"capability_fit": &fit}, map[string]string{"capability_fit": "reviewed capabilities"}, nil)
	if partial.Score == nil || *partial.Score != 90 || partial.DataCompleteness != .35 || !partial.Provisional || partial.Dimensions[1].Status != "unknown" {
		t.Fatalf("partial score: %#v", partial)
	}
	failed := Score(map[string]*int{"capability_fit": &fit}, map[string]string{"capability_fit": "reviewed capabilities"}, []EligibilityFlag{{Requirement: "mandatory certificate", Status: "unmet", Evidence: "verified registry"}})
	if failed.EligibilityStatus != "ineligible" || failed.Score == nil || *failed.Score != 90 {
		t.Fatalf("eligibility must remain separate: %#v", failed)
	}
	unproven := Score(map[string]*int{"capability_fit": &fit}, nil, []EligibilityFlag{{Requirement: "mandatory certificate", Status: "unmet"}})
	if unproven.EligibilityStatus != "unknown" || unproven.Score != nil {
		t.Fatalf("unproven data must remain unknown: %#v", unproven)
	}
	for _, tc := range []struct {
		value int
		band  string
	}{{59, "low"}, {60, "medium"}, {79, "medium"}, {80, "high"}} {
		v := tc.value
		got := Score(map[string]*int{"capability_fit": &v}, map[string]string{"capability_fit": "verified fit"}, nil)
		if got.Score == nil || *got.Score != v || got.PriorityBand != tc.band {
			t.Errorf("band %d: %#v", v, got)
		}
	}
}
