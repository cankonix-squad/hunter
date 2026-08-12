#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
script="$repo_dir/scripts/stability-report.sh"
doc="$repo_dir/docs/STABILITY_OBSERVATION.md"

sh -n "$script"
test -s "$doc"
for field in priority_sources missing_source_days success_rate_percent dead_letters source_configuration_interventions status; do
  grep -q "$field" "$script"
done
grep -Fq 'Exit `0`' "$doc"
grep -Fq 'Exit `2`' "$doc"
grep -Fq 'Exit `1`' "$doc"
echo "seven-day stability gate contract: ok"
