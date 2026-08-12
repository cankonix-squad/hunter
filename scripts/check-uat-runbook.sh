#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
runbook="$repo_dir/docs/DEPLOYMENT_RUNBOOK.md"
uat="$repo_dir/docs/UAT_CHECKLIST.md"

test -s "$runbook"
test -s "$uat"
sh -n "$repo_dir/scripts/staging-smoke.sh"

for section in 'Release ownership' 'Pre-deployment gate' 'Deploy' 'Post-deployment verification' 'Rollback' 'Production promotion'; do
  grep -Fq "## $section" "$runbook"
done

for check_id in UAT-ACCESS UAT-DISCOVER UAT-LIFECYCLE UAT-REVIEW UAT-WORKFLOW UAT-ALERT UAT-OPERATIONS UAT-AUDIT UAT-PERFORMANCE; do
  count=$(grep -c "| $check_id |" "$uat")
  test "$count" -eq 1
done

grep -Fq 'Product owner approval:' "$uat"
grep -Fq 'Operations approval:' "$uat"
grep -Fq 'Security approval:' "$uat"
echo "UAT and deployment runbook contract: ok"
