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

func TestValidatePreferences(t *testing.T) {
	if err := validatePreferences(PreferencesInput{QuietHoursEnabled: true, QuietHoursStart: "22:00", QuietHoursEnd: "07:00", Timezone: "Asia/Jakarta"}); err != nil {
		t.Fatal(err)
	}
	if validatePreferences(PreferencesInput{QuietHoursStart: "07:00", QuietHoursEnd: "07:00", Timezone: "Asia/Jakarta"}) == nil {
		t.Fatal("expected equal range rejection")
	}
	if validatePreferences(PreferencesInput{QuietHoursStart: "22:00", QuietHoursEnd: "07:00", Timezone: "Mars/Olympus"}) == nil {
		t.Fatal("expected timezone rejection")
	}
}
