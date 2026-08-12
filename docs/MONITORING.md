# Staging and production monitoring

The API exposes Prometheus text metrics at `/metrics` on its internal container port. The staging
edge does not route this path to the API, so it is not internet-facing. Request labels use the Go
route pattern rather than raw URL paths, preventing opportunity IDs or attacker-controlled paths
from creating high-cardinality series or leaking into monitoring.

The staging overlay runs Prometheus 3.13.1 LTS distroless, stores 15 days of data, and binds its UI
only to host loopback on port 9090 by default. Access it through the approved administrative tunnel;
do not publish it publicly. Persistent storage must be capacity-monitored and is operational data,
not a backup of the PostgreSQL system of record.

Metrics include process uptime, request count by method/route/status, and cumulative duration
histograms. Initial alerts detect scrape/API outage, sustained 5xx rate above 5%, and sustained p95
latency above one second. Connect Prometheus alerts to the organization-owned Alertmanager/on-call
route during hosted deployment and test one controlled alert before UAT approval. A rule being
loaded is not proof that a human will receive it.

Validate configuration with `scripts/check-monitoring.sh`. During release stabilization, record
dashboard/query links and UTC windows for availability, 5xx, p95, resource capacity, database
connections, collector source health/lag/queue, and notification dead letters. Application request
metrics complement the existing source/operations APIs; collector/database exporters and
distributed tracing remain follow-up observability work if hosted operations require them.
