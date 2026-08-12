package collector

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type Monitor struct {
	started          time.Time
	schedulerEnabled bool
	claimed          atomic.Uint64
	idle             atomic.Uint64
	iterationErrors  atomic.Uint64
	scheduleTicks    atomic.Uint64
	scheduledRuns    atomic.Uint64
	scheduleErrors   atomic.Uint64
}

func NewMonitor(schedulerEnabled bool) *Monitor {
	return &Monitor{started: time.Now(), schedulerEnabled: schedulerEnabled}
}

func (m *Monitor) ObserveIteration(processed bool, err error) {
	if err != nil {
		m.iterationErrors.Add(1)
		return
	}
	if processed {
		m.claimed.Add(1)
	} else {
		m.idle.Add(1)
	}
}

func (m *Monitor) ObserveSchedule(queued int, err error) {
	m.scheduleTicks.Add(1)
	if err != nil {
		m.scheduleErrors.Add(1)
		return
	}
	if queued > 0 {
		m.scheduledRuns.Add(uint64(queued))
	}
}

func (m *Monitor) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		fmt.Fprintln(w, "# HELP cankonix_collector_process_uptime_seconds Collector process uptime.")
		fmt.Fprintln(w, "# TYPE cankonix_collector_process_uptime_seconds gauge")
		fmt.Fprintf(w, "cankonix_collector_process_uptime_seconds %.3f\n", time.Since(m.started).Seconds())
		fmt.Fprintln(w, "# HELP cankonix_collector_scheduler_enabled Whether automatic scheduling is enabled.")
		fmt.Fprintln(w, "# TYPE cankonix_collector_scheduler_enabled gauge")
		fmt.Fprintf(w, "cankonix_collector_scheduler_enabled %d\n", boolInt(m.schedulerEnabled))
		writeCollectorCounter(w, "cankonix_collector_iterations_total", "claimed", m.claimed.Load())
		writeCollectorCounter(w, "cankonix_collector_iterations_total", "idle", m.idle.Load())
		writeCollectorCounter(w, "cankonix_collector_iterations_total", "error", m.iterationErrors.Load())
		fmt.Fprintln(w, "# HELP cankonix_collector_schedule_ticks_total Collector scheduler ticks.")
		fmt.Fprintln(w, "# TYPE cankonix_collector_schedule_ticks_total counter")
		fmt.Fprintf(w, "cankonix_collector_schedule_ticks_total %d\n", m.scheduleTicks.Load())
		fmt.Fprintln(w, "# HELP cankonix_collector_scheduled_runs_total Runs queued by the scheduler.")
		fmt.Fprintln(w, "# TYPE cankonix_collector_scheduled_runs_total counter")
		fmt.Fprintf(w, "cankonix_collector_scheduled_runs_total %d\n", m.scheduledRuns.Load())
		fmt.Fprintln(w, "# HELP cankonix_collector_schedule_errors_total Scheduler errors.")
		fmt.Fprintln(w, "# TYPE cankonix_collector_schedule_errors_total counter")
		fmt.Fprintf(w, "cankonix_collector_schedule_errors_total %d\n", m.scheduleErrors.Load())
	})
	return mux
}

func writeCollectorCounter(w http.ResponseWriter, name, result string, value uint64) {
	if result == "claimed" {
		fmt.Fprintln(w, "# HELP cankonix_collector_iterations_total Collector loop iterations by result.")
		fmt.Fprintln(w, "# TYPE cankonix_collector_iterations_total counter")
	}
	fmt.Fprintf(w, "%s{result=%q} %d\n", name, result, value)
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
