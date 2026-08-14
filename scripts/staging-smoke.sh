#!/bin/sh
set -eu

base_url=${1:-}
case "$base_url" in
  https://*/* | *'@'* | *'?'* | *'#'*)
    echo "usage: $0 https://staging-host (origin only, without credentials/path/query)" >&2
    exit 2
    ;;
  https://*) ;;
  *)
    echo "staging URL must use HTTPS" >&2
    exit 2
    ;;
esac

test_dir=$(mktemp -d)
trap 'rm -r "$test_dir"' EXIT INT TERM

request() {
  name=$1
  path=$2
  expected=$3
  code=$(curl --fail-with-body --silent --show-error --location --max-time 15 \
    --dump-header "$test_dir/$name.headers" --output "$test_dir/$name.body" \
    --write-out '%{http_code}' "$base_url$path" || true)
  if [ "$code" != "$expected" ]; then
    echo "$name: expected HTTP $expected, got $code" >&2
    exit 1
  fi
  echo "$name: HTTP $code"
}

request health /healthz 200
request readiness /readyz 200

code=$(curl --silent --show-error --max-time 15 --output "$test_dir/auth.body" \
  --write-out '%{http_code}' "$base_url/api/v1/opportunities" || true)
if [ "$code" != "401" ]; then
  echo "authorization boundary: expected HTTP 401, got $code" >&2
  exit 1
fi
echo "authorization boundary: HTTP 401"

headers="$test_dir/health.headers"
for header in strict-transport-security content-security-policy x-content-type-options x-frame-options referrer-policy permissions-policy; do
  if ! grep -Eiq "^$header:" "$headers"; then
    echo "missing required response header: $header" >&2
    exit 1
  fi
done

echo "staging HTTPS smoke: ok"
