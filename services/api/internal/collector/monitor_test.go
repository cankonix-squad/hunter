package collector

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCollectorMonitorHealthAndBoundedMetrics(t *testing.T) {
	monitor := NewMonitor(true)
	monitor.ObserveIteration(true, nil)
	monitor.ObserveIteration(false, nil)
	monitor.ObserveIteration(false, errors.New("database unavailable"))
	monitor.ObserveSchedule(3, nil)
	monitor.ObserveSchedule(0, errors.New("schedule failed"))

	health := httptest.NewRecorder()
	monitor.Handler().ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK || health.Body.String() != `{"status":"ok"}` {
		t.Fatalf("health response = %d %q", health.Code, health.Body.String())
	}
	metrics := httptest.NewRecorder()
	monitor.Handler().ServeHTTP(metrics, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	for _, required := range []string{
		"cankonix_collector_scheduler_enabled 1",
		`cankonix_collector_iterations_total{result="claimed"} 1`,
		`cankonix_collector_iterations_total{result="idle"} 1`,
		`cankonix_collector_iterations_total{result="error"} 1`,
		"cankonix_collector_schedule_ticks_total 2",
		"cankonix_collector_scheduled_runs_total 3",
		"cankonix_collector_schedule_errors_total 1",
	} {
		if !strings.Contains(metrics.Body.String(), required) {
			t.Fatalf("metrics missing %q:\n%s", required, metrics.Body.String())
		}
	}
}
