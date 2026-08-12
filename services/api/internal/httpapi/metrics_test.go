package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPMetricsUseBoundedRouteAndHistogramLabels(t *testing.T) {
	metrics := newHTTPMetrics()
	handler := metrics.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Pattern = "GET /api/v1/opportunities/{id}"
		w.WriteHeader(http.StatusCreated)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/opportunities/private-id", nil))

	recorder := httptest.NewRecorder()
	metrics.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := recorder.Body.String()
	if strings.Contains(body, "private-id") {
		t.Fatal("metrics leaked raw path parameters")
	}
	for _, required := range []string{`route="GET /api/v1/opportunities/{id}"`, `status="201"`, "cankonix_http_requests_total", "cankonix_http_request_duration_seconds_bucket", `le="+Inf"`} {
		if !strings.Contains(body, required) {
			t.Fatalf("metrics missing %q:\n%s", required, body)
		}
	}
}

func TestRetryDurationBucketsAreCumulative(t *testing.T) {
	metrics := newHTTPMetrics()
	metrics.observe(http.MethodGet, "GET /healthz", http.StatusOK, 75*time.Millisecond)
	metric := metrics.values[metricKey{method: http.MethodGet, route: "GET /healthz", status: http.StatusOK}]
	if metric.buckets[0] != 0 || metric.buckets[1] != 1 || metric.buckets[len(metric.buckets)-1] != 1 {
		t.Fatalf("unexpected cumulative buckets: %#v", metric.buckets)
	}
}
