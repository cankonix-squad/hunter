package opportunities

import (
	"reflect"
	"testing"
	"time"
)

func TestChangedFields(t *testing.T) {
	budgetBefore, budgetAfter := int64(100), int64(150)
	deadlineBefore := time.Date(2026, 8, 10, 1, 0, 0, 0, time.UTC)
	deadlineAfter := deadlineBefore.Add(time.Hour)
	before := Normalized{Title: "Title", Agency: "Agency", ProcurementType: "Goods", LifecycleStatus: "announced", BudgetMinor: &budgetBefore, Currency: "IDR", RegistrationDeadline: &deadlineBefore}
	after := before
	after.LifecycleStatus = "registration_open"
	after.BudgetMinor = &budgetAfter
	after.RegistrationDeadline = &deadlineAfter
	want := []string{"lifecycle_status", "budget_minor", "registration_deadline"}
	if got := ChangedFields(&before, after); !reflect.DeepEqual(got, want) {
		t.Fatalf("changedFields() = %v, want %v", got, want)
	}
	if got := ChangedFields(nil, after); len(got) != 0 {
		t.Fatalf("initial version changes = %v", got)
	}
}
