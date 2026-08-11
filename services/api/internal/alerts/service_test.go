package alerts

import "testing"

func TestValidateAlertRule(t *testing.T) {
	score := 7500
	input := Input{Name: "  Priority changes  ", EventTypes: []string{"opportunity.new", "opportunity.material_change"}, Filters: Filters{MinimumScoreBP: &score}}
	if err := validate(&input); err != nil {
		t.Fatal(err)
	}
	if input.Name != "Priority changes" || input.Filters.Relevance != "all" {
		t.Fatalf("unexpected normalized input: %#v", input)
	}
}

func TestValidateAlertRuleRejectsDuplicateEvents(t *testing.T) {
	input := Input{Name: "Duplicate", EventTypes: []string{"opportunity.new", "opportunity.new"}}
	if validate(&input) == nil {
		t.Fatal("expected duplicate event rejection")
	}
}
