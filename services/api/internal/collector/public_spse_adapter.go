package collector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const publicSPSEMaxResponse = 1 << 20

var authenticityTokenPattern = regexp.MustCompile("d\\.authenticityToken\\s*=\\s*'([^']+)'")

type PublicSPSEAdapter struct {
	client      *http.Client
	mu          sync.Mutex
	lastRequest time.Time
}

func NewPublicSPSEAdapter(client *http.Client) *PublicSPSEAdapter {
	if client == nil {
		jar, _ := cookiejar.New(nil)
		client = &http.Client{Timeout: 20 * time.Second, Jar: jar}
	}
	return &PublicSPSEAdapter{client: client}
}

func (*PublicSPSEAdapter) Name() string    { return "spse-public-list" }
func (*PublicSPSEAdapter) Version() string { return "v1" }

func (a *PublicSPSEAdapter) Discover(ctx context.Context, source Source) ([]Candidate, error) {
	var raw map[string]any
	if json.Unmarshal(source.Policy, &raw) != nil {
		return nil, errors.New("invalid public SPSE policy")
	}
	approved, _ := raw["public_collection_approved"].(bool)
	tenant, _ := raw["tenant"].(string)
	if !approved {
		return nil, errors.New("public collection is not approved for this source")
	}
	if tenant == "" || strings.ContainsAny(tenant, "/\\") {
		return nil, errors.New("public SPSE tenant is invalid")
	}
	year := time.Now().Year()
	if number, ok := raw["year"].(float64); ok {
		year = int(number)
	}
	pageSize := 20
	if number, ok := raw["page_size"].(float64); ok {
		pageSize = int(number)
	}
	if pageSize < 1 || pageSize > 50 {
		return nil, errors.New("public SPSE page_size must be between 1 and 50")
	}
	search, _ := raw["search"].(string)
	base, err := url.Parse(source.BaseURL)
	if err != nil || base.Scheme != "https" {
		return nil, errors.New("public SPSE base URL must use HTTPS")
	}

	listURL := strings.TrimRight(source.BaseURL, "/") + "/" + tenant + "/lelang"
	body, err := a.request(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return nil, fmt.Errorf("load public listing: %w", err)
	}
	match := authenticityTokenPattern.FindSubmatch(body)
	if len(match) != 2 {
		return nil, errors.New("public listing authenticity token is unavailable")
	}
	form := url.Values{"authenticityToken": {string(match[1])}, "draw": {"1"}, "start": {"0"},
		"length": {strconv.Itoa(pageSize)}, "search[value]": {search}}
	endpoint := fmt.Sprintf("%s/%s/dt/lelang?tahun=%d", strings.TrimRight(source.BaseURL, "/"), tenant, year)
	body, err = a.request(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("load public package metadata: %w", err)
	}
	return parsePublicSPSE(body, source.BaseURL, tenant)
}

func (a *PublicSPSEAdapter) request(ctx context.Context, method, target string, body io.Reader) ([]byte, error) {
	a.mu.Lock()
	wait := 3*time.Second - time.Since(a.lastRequest)
	if wait > 0 {
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			a.mu.Unlock()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	a.lastRequest = time.Now()
	a.mu.Unlock()
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "CANKONIX-OI/0.1 public-procurement-metadata")
	req.Header.Set("Accept", "application/json,text/html;q=0.9")
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	response, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode == 403 || response.StatusCode == 429 {
		return nil, fmt.Errorf("source access stopped with HTTP %d", response.StatusCode)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("source returned HTTP %d", response.StatusCode)
	}
	payload, err := io.ReadAll(io.LimitReader(response.Body, publicSPSEMaxResponse+1))
	if err != nil {
		return nil, err
	}
	if len(payload) > publicSPSEMaxResponse {
		return nil, errors.New("source response exceeds 1 MiB")
	}
	return payload, nil
}

func parsePublicSPSE(payload []byte, baseURL, tenant string) ([]Candidate, error) {
	var response struct{ Data [][]any }
	if err := json.Unmarshal(payload, &response); err != nil {
		return nil, fmt.Errorf("parse SPSE listing JSON: %w", err)
	}
	items := make([]Candidate, 0, len(response.Data))
	for index, row := range response.Data {
		if len(row) < 11 {
			return nil, fmt.Errorf("SPSE row %d has %d columns, want at least 11", index, len(row))
		}
		code, title, agency, stage := cell(row, 0), cell(row, 1), cell(row, 2), cell(row, 3)
		procurementType := strings.TrimSpace(strings.Split(cell(row, 8), " - TA ")[0])
		if code == "" || title == "" || agency == "" || procurementType == "" {
			return nil, fmt.Errorf("SPSE row %d is missing critical metadata", index)
		}
		budget, err := parseIDRMinor(cell(row, 10))
		if err != nil {
			return nil, fmt.Errorf("SPSE row %d budget: %w", index, err)
		}
		metadata := map[string]any{"title": title, "agency": agency, "procurement_type": procurementType,
			"status": mapSPSEStage(stage), "currency": "IDR", "pagu_minor": budget}
		items = append(items, Candidate{PackageIdentity: code,
			CanonicalURL: fmt.Sprintf("%s/%s/lelang/%s/pengumumanlelang", strings.TrimRight(baseURL, "/"), tenant, code), Metadata: metadata})
	}
	return items, nil
}

func cell(row []any, index int) string {
	value, _ := row[index].(string)
	return strings.TrimSpace(value)
}

func mapSPSEStage(stage string) string {
	value := strings.ToLower(stage)
	switch {
	case strings.Contains(value, "batal"):
		return "cancelled"
	case strings.Contains(value, "selesai"):
		return "completed"
	case strings.Contains(value, "pengumuman"):
		return "announced"
	case strings.Contains(value, "pendaftaran"), strings.Contains(value, "pemasukan penawaran"):
		return "registration_open"
	case strings.Contains(value, "evaluasi"), strings.Contains(value, "pembuktian"):
		return "evaluation"
	case strings.Contains(value, "penetapan"), strings.Contains(value, "pemenang"):
		return "awarded"
	default:
		return "registration_closed"
	}
}

func parseIDRMinor(raw string) (int64, error) {
	value := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "Rp."))
	parts := strings.Split(value, ",")
	whole := strings.ReplaceAll(strings.TrimSpace(parts[0]), ".", "")
	fraction := "00"
	if len(parts) > 1 {
		fraction = strings.TrimSpace(parts[1])
		if len(fraction) == 1 {
			fraction += "0"
		}
	}
	if len(fraction) != 2 {
		return 0, errors.New("invalid IDR decimal")
	}
	return strconv.ParseInt(whole+fraction, 10, 64)
}
