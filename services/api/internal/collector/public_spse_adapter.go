package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
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

var (
	authenticityTokenPattern = regexp.MustCompile("d\\.authenticityToken\\s*=\\s*'([^']+)'")
	htmlTagPattern           = regexp.MustCompile(`<[^>]+>`)
)

type PublicSPSEAdapter struct {
	client      *http.Client
	mu          sync.Mutex
	lastRequest time.Time
	minInterval time.Duration
	metrics     map[string]int
}

func NewPublicSPSEAdapter(client *http.Client) *PublicSPSEAdapter {
	if client == nil {
		jar, _ := cookiejar.New(nil)
		client = &http.Client{Timeout: 20 * time.Second, Jar: jar}
	}
	return &PublicSPSEAdapter{client: client, minInterval: 3 * time.Second}
}

func (*PublicSPSEAdapter) Name() string    { return "spse-public-list" }
func (*PublicSPSEAdapter) Version() string { return "v2" }

func (a *PublicSPSEAdapter) DiscoveryMetrics() map[string]int {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make(map[string]int, len(a.metrics))
	for key, value := range a.metrics {
		result[key] = value
	}
	return result
}

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
	maxPages := 3
	if number, ok := raw["max_pages_per_keyword"].(float64); ok {
		maxPages = int(number)
	}
	if maxPages < 1 || maxPages > 10 {
		return nil, errors.New("public SPSE max_pages_per_keyword must be between 1 and 10")
	}
	intervalMS := 3000
	if number, ok := raw["min_interval_ms"].(float64); ok {
		intervalMS = int(number)
	}
	if intervalMS < 2000 {
		return nil, errors.New("public SPSE min_interval_ms must be at least 2000")
	}
	a.mu.Lock()
	a.minInterval = time.Duration(intervalMS) * time.Millisecond
	a.mu.Unlock()
	keywords := policyKeywords(raw)
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
	endpoint := fmt.Sprintf("%s/%s/dt/lelang?tahun=%d", strings.TrimRight(source.BaseURL, "/"), tenant, year)
	unique := make(map[string]Candidate)
	metrics := map[string]int{"keywords": len(keywords)}
	for keywordIndex, keyword := range keywords {
		for page := 0; page < maxPages; page++ {
			form := dataTablesForm(string(match[1]), keyword, keywordIndex*maxPages+page+1, page*pageSize, pageSize)
			body, err = a.request(ctx, http.MethodPost, endpoint, []byte(form.Encode()))
			if err != nil {
				return nil, fmt.Errorf("load public package metadata for keyword %q page %d: %w", keyword, page+1, err)
			}
			items, err := parsePublicSPSE(body, source.BaseURL, tenant)
			if err != nil {
				return nil, err
			}
			metrics["pages"]++
			metrics["raw_rows"] += len(items)
			for _, item := range items {
				if !candidateTitleMatchesKeyword(item, keyword) {
					metrics["keyword_misses"]++
					continue
				}
				if _, exists := unique[item.PackageIdentity]; exists {
					metrics["duplicates"]++
					continue
				}
				unique[item.PackageIdentity] = item
				status, _ := item.Metadata["status"].(string)
				metrics["lifecycle_"+status]++
			}
			if len(items) < pageSize {
				break
			}
		}
	}
	items := make([]Candidate, 0, len(unique))
	for _, item := range unique {
		items = append(items, item)
	}
	metrics["unique_candidates"] = len(items)
	a.mu.Lock()
	a.metrics = metrics
	a.mu.Unlock()
	return items, nil
}

func policyKeywords(raw map[string]any) []string {
	result := make([]string, 0)
	if values, ok := raw["keywords"].([]any); ok {
		for _, value := range values {
			if keyword, ok := value.(string); ok && strings.TrimSpace(keyword) != "" {
				result = append(result, strings.TrimSpace(keyword))
			}
		}
	}
	if len(result) == 0 {
		if keyword, ok := raw["search"].(string); ok && strings.TrimSpace(keyword) != "" {
			result = append(result, strings.TrimSpace(keyword))
		}
	}
	if len(result) == 0 {
		result = []string{"sistem informasi", "aplikasi", "integrasi data", "jaringan"}
	}
	if len(result) > 8 {
		result = result[:8]
	}
	return result
}

func dataTablesForm(token, keyword string, draw, start, length int) url.Values {
	form := url.Values{"authenticityToken": {token}, "draw": {strconv.Itoa(draw)}, "start": {strconv.Itoa(start)},
		"length": {strconv.Itoa(length)}, "search[value]": {keyword}, "search[regex]": {"false"},
		"order[0][column]": {"5"}, "order[0][dir]": {"desc"}}
	for index := 0; index < 6; index++ {
		prefix := fmt.Sprintf("columns[%d]", index)
		form.Set(prefix+"[data]", strconv.Itoa(index))
		form.Set(prefix+"[name]", "")
		form.Set(prefix+"[searchable]", strconv.FormatBool(index != 3))
		form.Set(prefix+"[orderable]", strconv.FormatBool(index != 3))
		form.Set(prefix+"[search][value]", "")
		form.Set(prefix+"[search][regex]", "false")
	}
	return form
}

func (a *PublicSPSEAdapter) request(ctx context.Context, method, target string, body []byte) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		payload, retry, err := a.requestOnce(ctx, method, target, body)
		if err == nil {
			return payload, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(1<<attempt) * time.Second):
		}
	}
	return nil, lastErr
}

func (a *PublicSPSEAdapter) requestOnce(ctx context.Context, method, target string, body []byte) ([]byte, bool, error) {
	a.mu.Lock()
	wait := a.minInterval - time.Since(a.lastRequest)
	if wait > 0 {
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			a.mu.Unlock()
			return nil, false, ctx.Err()
		case <-timer.C:
		}
	}
	a.lastRequest = time.Now()
	a.mu.Unlock()
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", "CANKONIX-OI/0.1 public-procurement-metadata")
	req.Header.Set("Accept", "application/json,text/html;q=0.9")
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	response, err := a.client.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer response.Body.Close()
	if response.StatusCode == 403 || response.StatusCode == 429 {
		return nil, false, permanentSourceError{fmt.Errorf("source access stopped with HTTP %d", response.StatusCode)}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		err := fmt.Errorf("source returned HTTP %d", response.StatusCode)
		if response.StatusCode < 500 {
			return nil, false, permanentSourceError{err}
		}
		return nil, true, err
	}
	payload, err := io.ReadAll(io.LimitReader(response.Body, publicSPSEMaxResponse+1))
	if err != nil {
		return nil, true, err
	}
	if len(payload) > publicSPSEMaxResponse {
		return nil, false, permanentSourceError{errors.New("source response exceeds 1 MiB")}
	}
	return payload, false, nil
}

type permanentSourceError struct{ err error }

func (e permanentSourceError) Error() string   { return e.err.Error() }
func (e permanentSourceError) Unwrap() error   { return e.err }
func (e permanentSourceError) Permanent() bool { return true }

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
		code, title, agency, stage := cell(row, 0), cleanListingText(cell(row, 1)), cell(row, 2), cell(row, 3)
		procurementType := strings.TrimSpace(strings.Split(cell(row, 8), " - TA ")[0])
		if code == "" || title == "" || agency == "" || procurementType == "" {
			return nil, fmt.Errorf("SPSE row %d is missing critical metadata", index)
		}
		budget, err := parseOptionalIDRMinor(cell(row, 10))
		if err != nil {
			return nil, fmt.Errorf("SPSE row %d budget: %w", index, err)
		}
		metadata := map[string]any{"title": title, "agency": agency, "procurement_type": procurementType,
			"status": mapSPSEStage(stage), "currency": "IDR"}
		if budget != nil {
			metadata["pagu_minor"] = *budget
		}
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

func parseOptionalIDRMinor(raw string) (*int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" || strings.Contains(strings.ToLower(value), "belum dibuat") {
		return nil, nil
	}
	minor, err := parseIDRMinor(value)
	if err != nil {
		return nil, err
	}
	return &minor, nil
}

func cleanListingText(raw string) string {
	withoutTags := htmlTagPattern.ReplaceAllString(raw, " ")
	return strings.Join(strings.Fields(html.UnescapeString(withoutTags)), " ")
}

func candidateTitleMatchesKeyword(candidate Candidate, keyword string) bool {
	title, _ := candidate.Metadata["title"].(string)
	return strings.Contains(" "+searchWords(title)+" ", " "+searchWords(keyword)+" ")
}

func searchWords(value string) string {
	return strings.Join(strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	}), " ")
}
