package savedviews

import "testing"

func TestValidateDefaultsAndRejectsInvalidFilters(t *testing.T) {
	input := CreateInput{Name: "  Priority  ", Filters: Filters{}}
	if err := validate(&input); err != nil {
		t.Fatal(err)
	}
	if input.Name != "Priority" || input.Filters.Status != "active" || input.Filters.Relevance != "all" {
		t.Fatalf("unexpected defaults: %#v", input)
	}
	badScore := 10001
	bad := []CreateInput{
		{Name: "", Filters: Filters{}},
		{Name: "bad status", Filters: Filters{Status: "unknown"}},
		{Name: "bad score", Filters: Filters{MinimumScoreBP: &badScore}},
		{Name: "bad source", Filters: Filters{SourceID: "not-a-uuid"}},
	}
	for _, item := range bad {
		if err := validate(&item); err == nil {
			t.Fatalf("validate(%#v) unexpectedly succeeded", item)
		}
	}
}
