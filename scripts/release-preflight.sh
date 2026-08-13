#!/bin/sh
set -u

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
blocked=0

block() {
  echo "BLOCKED: $1" >&2
  blocked=1
}

require_value() {
  key=$1
  eval "value=\${$key:-}"
  if [ -z "$value" ]; then
    block "$key is required"
  fi
}

for key in RELEASE_COMMIT HOSTED_CI_URL HOSTED_CI_COMMIT STAGING_URL \
  UAT_PRODUCT_APPROVAL_REF UAT_OPERATIONS_APPROVAL_REF UAT_SECURITY_APPROVAL_REF \
  BACKUP_RESTORE_EVIDENCE_REF ONCALL_OWNER ROLLBACK_COMMIT \
  API_IMAGE_DIGEST COLLECTOR_IMAGE_DIGEST NOTIFIER_IMAGE_DIGEST WEB_IMAGE_DIGEST \
  API_IMAGE_REF COLLECTOR_IMAGE_REF NOTIFIER_IMAGE_REF WEB_IMAGE_REF; do
  require_value "$key"
done
for pair in "API_IMAGE_REF:API_IMAGE_DIGEST" "COLLECTOR_IMAGE_REF:COLLECTOR_IMAGE_DIGEST" \
  "NOTIFIER_IMAGE_REF:NOTIFIER_IMAGE_DIGEST" "WEB_IMAGE_REF:WEB_IMAGE_DIGEST"; do
  ref_key=${pair%%:*}
  digest_key=${pair#*:}
  eval "ref=\${$ref_key:-}"
  eval "digest=\${$digest_key:-}"
  if [ -n "$ref" ] && [ -n "$digest" ]; then
    case "$ref" in
      *@"$digest") ;;
      *) block "$ref_key must end with @$digest_key" ;;
    esac
  fi
done

cd "$repo_dir"
head_commit=$(git rev-parse HEAD)
if [ -n "$(git status --porcelain)" ]; then
  block "worktree is not clean"
fi
if ! git remote get-url origin >/dev/null 2>&1; then
  block "Git remote origin is not configured"
fi
if [ -n "${RELEASE_COMMIT:-}" ] && [ "$RELEASE_COMMIT" != "$head_commit" ]; then
  block "RELEASE_COMMIT does not match HEAD"
fi
if [ -n "${HOSTED_CI_COMMIT:-}" ] && [ "$HOSTED_CI_COMMIT" != "$head_commit" ]; then
  block "HOSTED_CI_COMMIT does not match HEAD"
fi
if [ -n "${HOSTED_CI_URL:-}" ]; then
  case "$HOSTED_CI_URL" in https://*) ;; *) block "HOSTED_CI_URL must use HTTPS" ;; esac
fi
if [ -n "${STAGING_URL:-}" ]; then
  case "$STAGING_URL" in https://*) ;; *) block "STAGING_URL must use HTTPS" ;; esac
fi

for key in API_IMAGE_DIGEST COLLECTOR_IMAGE_DIGEST NOTIFIER_IMAGE_DIGEST WEB_IMAGE_DIGEST; do
  eval "digest=\${$key:-}"
  if [ -n "$digest" ] && ! printf '%s\n' "$digest" | grep -Eq '^sha256:[0-9a-f]{64}$'; then
    block "$key must be an immutable sha256 digest"
  fi
done
if [ -n "${ROLLBACK_COMMIT:-}" ] && ! git cat-file -e "$ROLLBACK_COMMIT^{commit}" 2>/dev/null; then
  block "ROLLBACK_COMMIT is not available in this repository"
fi

if [ "$blocked" -ne 0 ]; then
  echo "release_preflight=NO-GO"
  exit 2
fi

set -e
evidence_tmp=$(mktemp -d)
trap 'rm -r "$evidence_tmp"' EXIT INT TERM

./scripts/check-production-secrets.sh
./scripts/check-release-images.sh
./scripts/check-staging-config.sh
./scripts/check-uat-runbook.sh
./scripts/check-stability-gate.sh
./scripts/classification-release-gate.sh >"$evidence_tmp/classification-release.json"
./scripts/stability-report.sh "${STABILITY_END_DATE:-$(date -u +%F)}" >"$evidence_tmp/stability-release.txt"
./scripts/staging-smoke.sh "$STAGING_URL"

echo "release_commit=$head_commit"
echo "release_preflight=GO"
