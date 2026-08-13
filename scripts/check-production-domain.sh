#!/bin/sh
set -eu

public_url=${CANKONIX_PRODUCTION_URL:-https://hunter.cankonix.com}
case "$public_url" in
  https://*) ;;
  *) echo "FAIL: CANKONIX_PRODUCTION_URL must use https://" >&2; exit 1 ;;
esac
public_url=${public_url%/}

check() {
  name=$1
  path=$2
  expected=$3
  response_file=$(mktemp)
  trap 'rm -f "$response_file"' EXIT INT TERM
  if ! code=$(curl -fsS --max-time 15 -o "$response_file" -w '%{http_code}' "$public_url$path"); then
    echo "FAIL: $name TLS/HTTP request failed for $public_url$path" >&2
    exit 1
  fi
  if [ "$code" != "$expected" ]; then
    echo "FAIL: $name HTTP $code, want $expected" >&2
    exit 1
  fi
  rm -f "$response_file"
  trap - EXIT INT TERM
  echo "PASS: $name HTTP $code"
}

check root / 200
check health /healthz 200
check readiness /readyz 200
echo "CANKONIX production domain gate: PASS ($public_url)"
