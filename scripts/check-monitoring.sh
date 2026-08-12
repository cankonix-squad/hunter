#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

docker run --rm \
  -v "$repo_dir/monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro" \
  -v "$repo_dir/monitoring/alerts.yml:/etc/prometheus/alerts.yml:ro" \
  --entrypoint /bin/promtool \
  prom/prometheus:v3.13.1-distroless \
  check config /etc/prometheus/prometheus.yml

grep -q '127.0.0.1:' "$repo_dir/compose.staging.yaml"
echo "monitoring configuration: ok"
