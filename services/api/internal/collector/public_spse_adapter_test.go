package collector

import (
	"context"
	_ "embed"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
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

func TestDiscoverTreatsNotFoundAfterFirstPageAsPaginationEnd(t *testing.T) {
	var mu sync.Mutex
	starts := make([]string, 0, 2)
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		response := `<html><script>d.authenticityToken = 'token'</script></html>`
		status := http.StatusOK
		if request.Method == http.MethodPost {
			body, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatal(err)
			}
			form := string(body)
			mu.Lock()
			starts = append(starts, form)
			mu.Unlock()
			if strings.Contains(form, "start=1") {
				status = http.StatusNotFound
				response = ""
			} else {
				response = `{"data":[["123","Pengembangan Sistem","Jakarta","Pengumuman Pascakualifikasi","","","Tender","","Jasa Lainnya - TA 2026","","Rp. 1.000,00"]]}`
			}
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(response))}, nil
	})}
	adapter := NewPublicSPSEAdapter(client)
	source := Source{BaseURL: "https://spse.example.test", Policy: []byte(`{"public_collection_approved":true,"tenant":"jakarta","keywords":["sistem"],"page_size":1,"max_pages_per_keyword":3,"min_interval_ms":2000}`)}

	items, err := adapter.Discover(context.Background(), source)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(items) != 1 || items[0].PackageIdentity != "123" {
		t.Fatalf("unexpected items: %#v", items)
	}
	if len(starts) != 2 || adapter.DiscoveryMetrics()["pagination_end_404"] != 1 {
		t.Fatalf("starts=%#v metrics=%#v", starts, adapter.DiscoveryMetrics())
	}
}

func TestDiscoverKeepsNotFoundOnFirstPageAsFailure(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method == http.MethodGet {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`<script>d.authenticityToken = 'token'</script>`))}, nil
		}
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	adapter := NewPublicSPSEAdapter(client)
	source := Source{BaseURL: "https://spse.example.test", Policy: []byte(`{"public_collection_approved":true,"tenant":"jakarta","keywords":["sistem"],"page_size":1,"min_interval_ms":2000}`)}

	_, err := adapter.Discover(context.Background(), source)
	if err == nil || !isHTTPStatus(err, http.StatusNotFound) {
		t.Fatalf("first-page 404 must fail with its status preserved, got %v", err)
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
