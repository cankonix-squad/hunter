package collector

import (
	_ "embed"
	"testing"
)

//go:embed fixtures/spse_public_maluku_v1.json
var publicSPSEFixture []byte

func TestParsePublicSPSEFixture(t *testing.T) {
	items, err := parsePublicSPSE(publicSPSEFixture, "https://spse.inaproc.id", "malukuprov")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items", len(items))
	}
	if items[0].Metadata["status"] != "completed" || items[1].Metadata["status"] != "registration_open" {
		t.Fatalf("unexpected lifecycle mapping: %#v", items)
	}
	if items[1].Metadata["pagu_minor"] != int64(100_000_000_000) {
		t.Fatalf("unexpected budget: %#v", items[1].Metadata["pagu_minor"])
	}
}

func TestUnknownStageIsNotAddedToActiveFeed(t *testing.T) {
	if got := mapSPSEStage("Tahap Baru yang Belum Dikenal"); got != "registration_closed" {
		t.Fatalf("got %q", got)
	}
}

func TestParseIDRMinorUsesIntegers(t *testing.T) {
	got, err := parseIDRMinor("Rp. 497.260.626,00")
	if err != nil || got != 49_726_062_600 {
		t.Fatalf("got %d, err=%v", got, err)
	}
}
