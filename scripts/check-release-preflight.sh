#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
preflight="$repo_dir/scripts/release-preflight.sh"

sh -n "$preflight"
for field in RELEASE_COMMIT HOSTED_CI_URL HOSTED_CI_COMMIT STAGING_URL \
  UAT_PRODUCT_APPROVAL_REF UAT_OPERATIONS_APPROVAL_REF UAT_SECURITY_APPROVAL_REF \
  BACKUP_RESTORE_EVIDENCE_REF ONCALL_OWNER ROLLBACK_COMMIT API_IMAGE_DIGEST \
  COLLECTOR_IMAGE_DIGEST NOTIFIER_IMAGE_DIGEST WEB_IMAGE_DIGEST; do
  grep -q "$field" "$preflight"
done
grep -q 'release_preflight=NO-GO' "$preflight"
grep -q 'release_preflight=GO' "$preflight"
echo "release preflight contract: ok"
