# Staging and production monitoring

The API and collector expose Prometheus text metrics at `/metrics` on internal container ports. The
staging edge does not route these paths, so they are not internet-facing. Request labels use the Go
route pattern rather than raw URL paths, preventing opportunity IDs or attacker-controlled paths
from creating high-cardinality series or leaking into monitoring.

The staging overlay runs Prometheus 3.13.1 LTS distroless, stores 15 days of data, and binds its UI
only to host loopback on port 9090 by default. Access it through the approved administrative tunnel;
do not publish it publicly. Persistent storage must be capacity-monitored and is operational data,
not a backup of the PostgreSQL system of record.

Metrics include process uptime, request count by method/route/status, cumulative duration
histograms, per-source health/lag/pending/failure/circuit state, persisted source-run status, and
notification delivery status. Database-backed collection has a two-second timeout and publishes a
scrape-success gauge so a partial exporter failure is visible. Alerts detect scrape/API outage,
sustained 5xx rate above 5%, sustained p95 latency above one second, operational scrape failure,
source lag/backlog/open circuit/dead letters, and notification dead letters. Connect Prometheus
alerts to the organization-owned Alertmanager/on-call route during hosted deployment and test one
controlled alert before UAT approval. A rule being loaded is not proof that a human will receive it.

Collector process metrics include uptime, scheduler-enabled state, claimed/idle/error iterations,
scheduler ticks, queued runs, and scheduler errors. Docker checks `/healthz` on internal port 8081;
Prometheus scrapes the same port and alerts within two minutes when the process disappears or when
scheduler errors increase.

Validate configuration with `scripts/check-monitoring.sh`. During release stabilization, record
dashboard/query links and UTC windows for availability, 5xx, p95, resource capacity, database
connections, collector source health/lag/queue, and notification dead letters. Dedicated host/
database exporters and distributed tracing remain follow-up observability work if hosted operations
require them.
