package collector

import (
	"context"
	"testing"
)

func TestFixtureAdapterIsStable(t *testing.T) {
	adapter := FixtureAdapter{}
	items, err := adapter.Discover(context.Background(), Source{})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d candidates, want 2", len(items))
	}
	a, err := items[0].Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	b, err := items[0].Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if a != b || len(a) != 71 {
		t.Fatalf("fingerprint is not stable: %q / %q", a, b)
	}
}

func TestRegistryUsesNameAndVersion(t *testing.T) {
	r := NewRegistry(FixtureAdapter{})
	if _, err := r.Resolve("spse-fixture", "v1"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Resolve("spse-fixture", "v2"); err == nil {
		t.Fatal("unsupported version must fail")
	}
}
