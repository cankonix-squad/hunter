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
