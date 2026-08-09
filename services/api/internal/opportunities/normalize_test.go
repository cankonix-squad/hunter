package opportunities

import "testing"

func TestNormalizeCandidate(t *testing.T) {
	item, err := Normalize(map[string]any{
		"title": " Integrasi Data ", "agency": "Instansi", "procurement_type": "Jasa",
		"status": "registration_open", "currency": "idr", "pagu_minor": float64(75000000),
		"registration_deadline": "2026-08-20T09:00:00+07:00",
	})
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != "Integrasi Data" || item.Currency != "IDR" || item.BudgetMinor == nil || *item.BudgetMinor != 75000000 {
		t.Fatalf("unexpected normalization: %#v", item)
	}
	if item.RegistrationDeadline == nil || item.RegistrationDeadline.Hour() != 2 {
		t.Fatalf("deadline was not normalized to UTC: %#v", item.RegistrationDeadline)
	}
}

func TestNormalizeRejectsUnknownLifecycle(t *testing.T) {
	_, err := Normalize(map[string]any{"title": "A", "agency": "B", "procurement_type": "C", "status": "mystery", "currency": "IDR"})
	if err == nil {
		t.Fatal("unknown lifecycle must fail")
	}
}
