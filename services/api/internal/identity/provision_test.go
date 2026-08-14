package identity

import "testing"

func TestValidateProvisionInput(t *testing.T) {
	email, name, roles, err := validateProvisionInput(" UAT@example.test ", " UAT User ", "01234567890123456789", []string{"operator", "operator"})
	if err != nil || email != "uat@example.test" || name != "UAT User" || len(roles) != 1 || roles[0] != "operator" {
		t.Fatalf("unexpected validation result: %q %q %#v %v", email, name, roles, err)
	}
	if _, _, _, err := validateProvisionInput("uat@example.test", "UAT", "01234567890123456789", []string{"owner"}); err == nil {
		t.Fatal("unknown role must be rejected")
	}
	if _, _, _, err := validateProvisionInput("uat@example.test", "UAT", "short", nil); err == nil {
		t.Fatal("short password must be rejected")
	}
}
