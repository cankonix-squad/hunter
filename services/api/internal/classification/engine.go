package classification

import (
	"sort"
	"strings"
)

const TaxonomyVersion = "cankonix-it-v1"

type Rule struct {
	Code               string
	Category           string
	Kind               string
	Keywords           []string
	Reason             string
	ApprovedAutoReject bool
}

type Evidence struct {
	RuleCode    string `json:"rule_code"`
	Kind        string `json:"kind"`
	Category    string `json:"category"`
	Keyword     string `json:"keyword"`
	SourceField string `json:"source_field"`
	Excerpt     string `json:"excerpt"`
}

type Result struct {
	TaxonomyVersion string     `json:"taxonomy_version"`
	RuleVersion     string     `json:"rule_version"`
	Disposition     string     `json:"disposition"`
	Category        string     `json:"category,omitempty"`
	ConfidenceBP    int        `json:"confidence_bp"`
	ReasonCode      string     `json:"reason_code"`
	Explanation     string     `json:"explanation"`
	AutoRejected    bool       `json:"auto_rejected"`
	Evidence        []Evidence `json:"evidence"`
}

type Input struct {
	Title           string
	ProcurementType string
}

type Engine struct {
	version string
	rules   []Rule
}

func DefaultEngine() Engine {
	return Engine{version: "deterministic-v1", rules: []Rule{
		{Code: "include_software", Category: "software-engineering", Kind: "include", Keywords: []string{"pengembangan aplikasi", "aplikasi web", "sistem informasi", "integrasi sistem", "software"}, Reason: "Paket menyebut pekerjaan pengembangan atau integrasi perangkat lunak."},
		{Code: "include_data", Category: "data-ai", Kind: "include", Keywords: []string{"integrasi data", "data warehouse", "analitik data", "artificial intelligence", "machine learning", "kecerdasan buatan"}, Reason: "Paket menyebut kapabilitas data atau AI."},
		{Code: "include_cloud", Category: "cloud-infrastructure", Kind: "include", Keywords: []string{"cloud", "komputasi awan", "devops", "container platform"}, Reason: "Paket menyebut kapabilitas cloud atau platform."},
		{Code: "include_security", Category: "cybersecurity", Kind: "include", Keywords: []string{"keamanan siber", "cyber security", "penetration test", "security assessment"}, Reason: "Paket menyebut layanan keamanan siber."},
		{Code: "exclude_non_it_supply", Kind: "exclude", Keywords: []string{"alat tulis kantor", "seragam", "katering", "kendaraan dinas", "bahan makanan"}, Reason: "Paket secara deterministik merupakan pengadaan non-IT.", ApprovedAutoReject: true},
		{Code: "exclude_it_consumable", Kind: "exclude", Keywords: []string{"toner printer", "tinta printer", "kertas printer"}, Reason: "Paket hanya menyebut barang habis pakai IT dan bukan layanan/produk inti CANKONIX.", ApprovedAutoReject: true},
	}}
}

func (e Engine) Evaluate(input Input) Result {
	fields := []struct{ name, value string }{{"title", input.Title}, {"procurement_type", input.ProcurementType}}
	var evidence []Evidence
	for _, rule := range e.rules {
		for _, field := range fields {
			value := strings.ToLower(strings.TrimSpace(field.value))
			for _, keyword := range rule.Keywords {
				if strings.Contains(value, strings.ToLower(keyword)) {
					evidence = append(evidence, Evidence{RuleCode: rule.Code, Kind: rule.Kind, Category: rule.Category, Keyword: keyword, SourceField: field.name, Excerpt: field.value})
					break
				}
			}
		}
	}
	sort.SliceStable(evidence, func(i, j int) bool { return evidence[i].RuleCode < evidence[j].RuleCode })

	includes, excludes := filter(evidence, "include"), filter(evidence, "exclude")
	result := Result{TaxonomyVersion: TaxonomyVersion, RuleVersion: e.version, Disposition: "needs_review", ConfidenceBP: 5000, ReasonCode: "no_deterministic_match", Explanation: "Tidak ada rule deterministik yang cukup untuk menentukan relevansi; perlu review manusia.", Evidence: evidence}
	if len(includes) > 0 && len(excludes) == 0 {
		result.Disposition, result.Category, result.ConfidenceBP = "relevant", includes[0].Category, 9000
		result.ReasonCode, result.Explanation = includes[0].RuleCode, reasonFor(e.rules, includes[0].RuleCode)
	} else if len(excludes) > 0 && len(includes) == 0 {
		result.Disposition, result.ConfidenceBP = "excluded", 9500
		result.ReasonCode, result.Explanation = excludes[0].RuleCode, reasonFor(e.rules, excludes[0].RuleCode)
		result.AutoRejected = approved(e.rules, excludes[0].RuleCode)
	} else if len(includes) > 0 && len(excludes) > 0 {
		result.ReasonCode, result.Explanation = "conflicting_rules", "Rule include dan exclude sama-sama cocok; perlu review manusia."
	}
	return result
}

func filter(items []Evidence, kind string) []Evidence {
	result := make([]Evidence, 0)
	for _, item := range items {
		if item.Kind == kind {
			result = append(result, item)
		}
	}
	return result
}

func reasonFor(rules []Rule, code string) string {
	for _, rule := range rules {
		if rule.Code == code {
			return rule.Reason
		}
	}
	return ""
}

func approved(rules []Rule, code string) bool {
	for _, rule := range rules {
		if rule.Code == code {
			return rule.ApprovedAutoReject
		}
	}
	return false
}
