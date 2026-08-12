package collector

import (
	"context"
	_ "embed"
	"errors"
	"io"
	"net/http"
	"strings"
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

func TestParsePublicSPSEToleratesUnavailableContractValueAndCleansTitle(t *testing.T) {
	payload := []byte(`{"data":[["123","Pengembangan Sistem <span class='badge'>Tender Ulang</span>","Komdigi","Pengumuman Pascakualifikasi","","","Tender","","Jasa Lainnya - TA 2026","","Nilai Kontrak belum dibuat"]]}`)
	items, err := parsePublicSPSE(payload, "https://spse.inaproc.id", "komdigi")
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
	if got := items[0].Metadata["title"]; got != "Pengembangan Sistem Tender Ulang" {
		t.Fatalf("title=%q", got)
	}
	if _, exists := items[0].Metadata["pagu_minor"]; exists {
		t.Fatal("unavailable contract value must remain unknown")
	}
}

func TestDataTablesFormContainsServerSideSearchContract(t *testing.T) {
	form := dataTablesForm("token", "sistem informasi", 2, 20, 20)
	checks := map[string]string{"search[value]": "sistem informasi", "search[regex]": "false",
		"columns[0][data]": "0", "columns[5][data]": "5", "order[0][column]": "5",
		"start": "20", "length": "20"}
	for key, want := range checks {
		if got := form.Get(key); got != want {
			t.Fatalf("%s=%q, want %q", key, got, want)
		}
	}
}

func TestPolicyKeywordsAreBoundedAndBackwardCompatible(t *testing.T) {
	raw := map[string]any{"keywords": []any{"aplikasi", "integrasi data", "", "jaringan"}}
	got := policyKeywords(raw)
	if len(got) != 3 || got[0] != "aplikasi" {
		t.Fatalf("unexpected keywords: %#v", got)
	}
	if legacy := policyKeywords(map[string]any{"search": "software"}); len(legacy) != 1 || legacy[0] != "software" {
		t.Fatalf("unexpected legacy keywords: %#v", legacy)
	}
}

func TestAccessBoundaryResponseIsPermanent(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusTooManyRequests, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	adapter := NewPublicSPSEAdapter(client)

	_, _, err := adapter.requestOnce(context.Background(), http.MethodGet, "https://spse.example.test", nil)
	var permanent interface{ Permanent() bool }
	if !errors.As(err, &permanent) || !permanent.Permanent() {
		t.Fatalf("429 error must be permanent, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestCandidateTitleKeywordMatchUsesWordBoundaries(t *testing.T) {
	item := Candidate{Metadata: map[string]any{"title": "Pembangunan Sarana Peribadatan"}}
	if candidateTitleMatchesKeyword(item, "data") {
		t.Fatal("data must not match the middle of peribadatan")
	}
	item.Metadata["title"] = "Operasional Sistem Big Data Nasional"
	if !candidateTitleMatchesKeyword(item, "data") || !candidateTitleMatchesKeyword(item, "sistem") {
		t.Fatal("whole title words must match")
	}
}
