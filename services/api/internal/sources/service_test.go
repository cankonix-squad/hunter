package sources

import (
	"encoding/json"
	"testing"
)

func TestValidateAppliesDefaults(t *testing.T) {
	in := CreateInput{Name: " LPSE Test ", BaseURL: "https://lpse.example.go.id/", AdapterName: "spse", AdapterVersion: "v1"}
	if err := validate(&in); err != nil {
		t.Fatalf("validate() error = %v", err)
	}
	if in.BaseURL != "https://lpse.example.go.id" || in.Timezone != "Asia/Jakarta" || string(in.Policy) != "{}" {
		t.Fatalf("unexpected normalized input: %#v", in)
	}
}

func TestValidateRejectsUnsafeURLAndPolicy(t *testing.T) {
	cases := []CreateInput{
		{Name: "test", BaseURL: "http://example.test", AdapterName: "spse", AdapterVersion: "v1"},
		{Name: "test", BaseURL: "https://example.test", AdapterName: "spse", AdapterVersion: "v1", Policy: json.RawMessage(`[]`)},
		{Name: "test", BaseURL: "https://example.test", AdapterName: "spse", AdapterVersion: "v1", Policy: json.RawMessage(`null`)},
	}
	for _, in := range cases {
		if err := validate(&in); err == nil {
			t.Fatalf("validate(%#v) unexpectedly succeeded", in)
		}
	}
}

func TestValidateRejectsInvalidSchedulePolicy(t *testing.T) {
	cases := []json.RawMessage{
		json.RawMessage(`{"schedule_enabled":"yes"}`),
		json.RawMessage(`{"schedule_interval_minutes":4}`),
		json.RawMessage(`{"schedule_interval_minutes":60.5}`),
		json.RawMessage(`{"schedule_interval_minutes":"60"}`),
	}
	for _, policy := range cases {
		in := CreateInput{Name: "test", BaseURL: "https://example.test", AdapterName: "spse", AdapterVersion: "v1", Policy: policy}
		if err := validate(&in); err == nil {
			t.Fatalf("validate() unexpectedly accepted policy %s", policy)
		}
	}
}
