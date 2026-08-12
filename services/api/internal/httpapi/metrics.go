package httpapi

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var requestDurationBuckets = []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5}

type metricKey struct {
	method string
	route  string
	status int
}

type requestMetric struct {
	count   uint64
	sum     float64
	buckets []uint64
}

type httpMetrics struct {
	mu      sync.Mutex
	started time.Time
	values  map[metricKey]*requestMetric
}

func newHTTPMetrics() *httpMetrics {
	return &httpMetrics{started: time.Now(), values: make(map[metricKey]*requestMetric)}
}

func (m *httpMetrics) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		tracked := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		defer func() { m.observe(r.Method, r.Pattern, tracked.status, time.Since(started)) }()
		next.ServeHTTP(tracked, r)
	})
}

func (m *httpMetrics) observe(method, route string, status int, duration time.Duration) {
	if route == "" {
		route = "unmatched"
	}
	key := metricKey{method: method, route: route, status: status}
	m.mu.Lock()
	defer m.mu.Unlock()
	metric := m.values[key]
	if metric == nil {
		metric = &requestMetric{buckets: make([]uint64, len(requestDurationBuckets))}
		m.values[key] = metric
	}
	metric.count++
	seconds := duration.Seconds()
	metric.sum += seconds
	for index, upperBound := range requestDurationBuckets {
		if seconds <= upperBound {
			metric.buckets[index]++
		}
	}
}

func (m *httpMetrics) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	m.mu.Lock()
	defer m.mu.Unlock()
	keys := make([]metricKey, 0, len(m.values))
	for key := range m.values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].method+keys[i].route+strconv.Itoa(keys[i].status) < keys[j].method+keys[j].route+strconv.Itoa(keys[j].status)
	})
	fmt.Fprintln(w, "# HELP cankonix_process_uptime_seconds API process uptime.")
	fmt.Fprintln(w, "# TYPE cankonix_process_uptime_seconds gauge")
	fmt.Fprintf(w, "cankonix_process_uptime_seconds %.3f\n", time.Since(m.started).Seconds())
	fmt.Fprintln(w, "# HELP cankonix_http_requests_total Completed HTTP requests.")
	fmt.Fprintln(w, "# TYPE cankonix_http_requests_total counter")
	fmt.Fprintln(w, "# HELP cankonix_http_request_duration_seconds HTTP request duration.")
	fmt.Fprintln(w, "# TYPE cankonix_http_request_duration_seconds histogram")
	for _, key := range keys {
		metric := m.values[key]
		labels := fmt.Sprintf(`method=%q,route=%q,status=%q`, escapeLabel(key.method), escapeLabel(key.route), strconv.Itoa(key.status))
		fmt.Fprintf(w, "cankonix_http_requests_total{%s} %d\n", labels, metric.count)
		for index, upperBound := range requestDurationBuckets {
			fmt.Fprintf(w, "cankonix_http_request_duration_seconds_bucket{%s,le=%q} %d\n", labels, strconv.FormatFloat(upperBound, 'g', -1, 64), metric.buckets[index])
		}
		fmt.Fprintf(w, "cankonix_http_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n", labels, metric.count)
		fmt.Fprintf(w, "cankonix_http_request_duration_seconds_sum{%s} %.6f\n", labels, metric.sum)
		fmt.Fprintf(w, "cankonix_http_request_duration_seconds_count{%s} %d\n", labels, metric.count)
	}
}

func escapeLabel(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\n", "\\n")
	return strings.ReplaceAll(value, "\"", "\\\"")
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
