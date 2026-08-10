package readiness

import "testing"

func TestChecklistDefinitionIsStable(t *testing.T) {
	want := []string{"scope_validated", "eligibility_confirmed", "delivery_capacity", "commercial_review", "document_plan"}
	if Version != "readiness-v1" || len(definitions) != len(want) {
		t.Fatal("unexpected checklist definition")
	}
	for i, code := range want {
		if definitions[i].Code != code || !validCode(code) {
			t.Fatalf("unexpected item %d", i)
		}
	}
	if validCode("unknown") {
		t.Fatal("accepted unknown code")
	}
}
