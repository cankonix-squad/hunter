package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
	db      *pgxpool.Pool
}

func newHTTPMetrics(db *pgxpool.Pool) *httpMetrics {
	return &httpMetrics{started: time.Now(), values: make(map[metricKey]*requestMetric), db: db}
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

func (m *httpMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	m.mu.Lock()
	keys := make([]metricKey, 0, len(m.values))
	values := make(map[metricKey]requestMetric, len(m.values))
	for key := range m.values {
		keys = append(keys, key)
		metric := m.values[key]
		values[key] = requestMetric{count: metric.count, sum: metric.sum, buckets: append([]uint64(nil), metric.buckets...)}
	}
	m.mu.Unlock()
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
		metric := values[key]
		labels := fmt.Sprintf(`method=%q,route=%q,status=%q`, escapeLabel(key.method), escapeLabel(key.route), strconv.Itoa(key.status))
		fmt.Fprintf(w, "cankonix_http_requests_total{%s} %d\n", labels, metric.count)
		for index, upperBound := range requestDurationBuckets {
			fmt.Fprintf(w, "cankonix_http_request_duration_seconds_bucket{%s,le=%q} %d\n", labels, strconv.FormatFloat(upperBound, 'g', -1, 64), metric.buckets[index])
		}
		fmt.Fprintf(w, "cankonix_http_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n", labels, metric.count)
		fmt.Fprintf(w, "cankonix_http_request_duration_seconds_sum{%s} %.6f\n", labels, metric.sum)
		fmt.Fprintf(w, "cankonix_http_request_duration_seconds_count{%s} %d\n", labels, metric.count)
	}
	if m.db != nil {
		m.writeOperationalMetrics(r.Context(), w)
	}
}

func (m *httpMetrics) writeOperationalMetrics(parent context.Context, w http.ResponseWriter) {
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	success := 1

	fmt.Fprintln(w, "# HELP cankonix_source_health_score Persisted source health score.")
	fmt.Fprintln(w, "# TYPE cankonix_source_health_score gauge")
	fmt.Fprintln(w, "# HELP cankonix_source_lag_seconds Seconds since the last successful run; -1 means never successful.")
	fmt.Fprintln(w, "# TYPE cankonix_source_lag_seconds gauge")
	fmt.Fprintln(w, "# HELP cankonix_source_pending_runs Queued or running source runs.")
	fmt.Fprintln(w, "# TYPE cankonix_source_pending_runs gauge")
	fmt.Fprintln(w, "# HELP cankonix_source_consecutive_failures Current consecutive failures.")
	fmt.Fprintln(w, "# TYPE cankonix_source_consecutive_failures gauge")
	fmt.Fprintln(w, "# HELP cankonix_source_circuit_open Whether the source circuit breaker is open.")
	fmt.Fprintln(w, "# TYPE cankonix_source_circuit_open gauge")
	rows, err := m.db.Query(ctx, `SELECT s.id::text,s.name,s.health_score,s.consecutive_failures,
		CASE WHEN s.circuit_open_until > now() THEN 1 ELSE 0 END,
		coalesce(extract(epoch FROM now()-s.last_success_at)::bigint,-1),
		(SELECT count(*) FROM source_runs r WHERE r.source_id=s.id AND r.status IN ('queued','running'))
		FROM sources s WHERE s.status='active' AND coalesce((s.policy->>'schedule_enabled')::boolean,false)
		ORDER BY s.id`)
	if err != nil {
		success = 0
	} else {
		for rows.Next() {
			var id, name string
			var health, failures, circuit, lag, pending int64
			if err := rows.Scan(&id, &name, &health, &failures, &circuit, &lag, &pending); err != nil {
				success = 0
				break
			}
			labels := fmt.Sprintf(`source_id=%q,source=%q`, escapeLabel(id), escapeLabel(name))
			fmt.Fprintf(w, "cankonix_source_health_score{%s} %d\n", labels, health)
			fmt.Fprintf(w, "cankonix_source_lag_seconds{%s} %d\n", labels, lag)
			fmt.Fprintf(w, "cankonix_source_pending_runs{%s} %d\n", labels, pending)
			fmt.Fprintf(w, "cankonix_source_consecutive_failures{%s} %d\n", labels, failures)
			fmt.Fprintf(w, "cankonix_source_circuit_open{%s} %d\n", labels, circuit)
		}
		if rows.Err() != nil {
			success = 0
		}
		rows.Close()
	}

	fmt.Fprintln(w, "# HELP cankonix_source_runs Current persisted source runs by status.")
	fmt.Fprintln(w, "# TYPE cankonix_source_runs gauge")
	if err := m.writeCountByLabels(ctx, w, `SELECT status::text,count(*) FROM source_runs GROUP BY status ORDER BY status`, "cankonix_source_runs", "status"); err != nil {
		success = 0
	}
	var recentDeadLetters int64
	if err := m.db.QueryRow(ctx, `SELECT count(*) FROM source_runs WHERE status='dead_letter' AND finished_at >= now()-interval '1 hour'`).Scan(&recentDeadLetters); err != nil {
		success = 0
	} else {
		fmt.Fprintln(w, "# HELP cankonix_source_dead_letters_recent Source runs dead-lettered in the last hour.")
		fmt.Fprintln(w, "# TYPE cankonix_source_dead_letters_recent gauge")
		fmt.Fprintf(w, "cankonix_source_dead_letters_recent %d\n", recentDeadLetters)
	}
	fmt.Fprintln(w, "# HELP cankonix_notification_deliveries Current notification deliveries by channel and status.")
	fmt.Fprintln(w, "# TYPE cankonix_notification_deliveries gauge")
	rows, err = m.db.Query(ctx, `SELECT channel,status,count(*) FROM notification_deliveries GROUP BY channel,status ORDER BY channel,status`)
	if err != nil {
		success = 0
	} else {
		for rows.Next() {
			var channel, status string
			var count int64
			if err := rows.Scan(&channel, &status, &count); err != nil {
				success = 0
				break
			}
			fmt.Fprintf(w, "cankonix_notification_deliveries{channel=%q,status=%q} %d\n", escapeLabel(channel), escapeLabel(status), count)
		}
		if rows.Err() != nil {
			success = 0
		}
		rows.Close()
	}
	fmt.Fprintln(w, "# HELP cankonix_operational_metrics_scrape_success Whether database-backed metrics were collected successfully.")
	fmt.Fprintln(w, "# TYPE cankonix_operational_metrics_scrape_success gauge")
	fmt.Fprintf(w, "cankonix_operational_metrics_scrape_success %d\n", success)
}

func (m *httpMetrics) writeCountByLabels(ctx context.Context, w http.ResponseWriter, query, metricName, labelName string) error {
	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var label string
		var count int64
		if err := rows.Scan(&label, &count); err != nil {
			return err
		}
		fmt.Fprintf(w, "%s{%s=%q} %d\n", metricName, labelName, escapeLabel(label), count)
	}
	return rows.Err()
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
