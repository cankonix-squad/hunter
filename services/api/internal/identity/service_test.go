package identity

import (
	"testing"
	"time"
)

func TestAccessTokenRoundTrip(t *testing.T) {
	s := New(nil, "01234567890123456789012345678901", time.Minute, time.Hour)
	want := Principal{UserID: "9d62cf04-c1d3-46dc-ac4c-25fc944061a8", Email: "admin@example.test", Name: "Admin", Roles: []string{"administrator"}}
	tokens, err := s.accessTokens(want, "refresh")
	if err != nil {
		t.Fatalf("accessTokens() error = %v", err)
	}
	got, err := s.ParseAccess(tokens.AccessToken)
	if err != nil {
		t.Fatalf("ParseAccess() error = %v", err)
	}
	if got.UserID != want.UserID || got.Email != want.Email || len(got.Roles) != 1 || got.Roles[0] != "administrator" {
		t.Fatalf("ParseAccess() = %#v, want %#v", got, want)
	}
}

func TestOpaqueTokensAreUnique(t *testing.T) {
	a, ah, err := newOpaqueToken()
	if err != nil {
		t.Fatal(err)
	}
	b, bh, err := newOpaqueToken()
	if err != nil {
		t.Fatal(err)
	}
	if a == b || string(ah) == string(bh) {
		t.Fatal("tokens must be unique")
	}
}
