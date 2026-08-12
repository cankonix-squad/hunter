#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_dir/services/api"

exec go run ./cmd/evaluate \
  -dataset testdata/classification_golden_v1.json \
  -minimum-precision-bp 8000 \
  -require-approved
