package classification

import "testing"

func TestDeterministicClassification(t *testing.T) {
	tests := []struct {
		name, title, want, category, reason string
		auto                                bool
	}{
		{"positive", "Pengembangan Aplikasi Web Pelayanan Publik", "relevant", "software-engineering", "include_software", false},
		{"fixture positive", "Pengadaan Layanan Integrasi Data", "relevant", "data-ai", "include_data", false},
		{"negative", "Pengadaan Alat Tulis Kantor Tahun 2026", "excluded", "", "exclude_non_it_supply", true},
		{"ambiguous", "Jasa Konsultansi Transformasi Organisasi", "needs_review", "", "no_deterministic_match", false},
		{"conflict", "Pengembangan aplikasi dan pengadaan alat tulis kantor", "needs_review", "", "conflicting_rules", false},
	}
	engine := DefaultEngine()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.Evaluate(Input{Title: tt.title, ProcurementType: "Jasa"})
			if got.Disposition != tt.want || got.Category != tt.category || got.ReasonCode != tt.reason || got.AutoRejected != tt.auto {
				t.Fatalf("unexpected result: %#v", got)
			}
		})
	}
}

func TestEvidenceIdentifiesMatchedFieldAndKeyword(t *testing.T) {
	got := DefaultEngine().Evaluate(Input{Title: "Belanja umum", ProcurementType: "Jasa Integrasi Sistem"})
	if len(got.Evidence) != 1 || got.Evidence[0].SourceField != "procurement_type" || got.Evidence[0].Keyword != "integrasi sistem" {
		t.Fatalf("unexpected evidence: %#v", got.Evidence)
	}
}
