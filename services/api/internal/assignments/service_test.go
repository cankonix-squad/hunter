package assignments

import "testing"

func TestValidUUID(t *testing.T) {
	if !validUUID("8f75b02b-d04f-43bc-99bd-f67d76d9e1dd") || validUUID("invalid") {
		t.Fatal("unexpected UUID validation")
	}
}
