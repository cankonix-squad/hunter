# Deployment and promotion runbook

This runbook promotes one immutable, reviewed commit from staging to production. It complements
`STAGING_DEPLOYMENT.md`, `SECRETS.md`, `BACKUP_RESTORE.md`, and
`FAILURE_RECOVERY_RUNBOOK.md`. Commands must use all three Compose files and run from the checked
out release revision. Store operational evidence outside the repository without secret values.

## Release ownership

Assign a release commander, database/backup operator, application verifier, security approver,
product UAT approver, rollback owner, and communication channel. Use UTC for the change record.
Only the release commander may declare GO/NO-GO; destructive recovery requires incident-owner
approval. Separate the deployer from at least one verifier where staffing permits.

## Pre-deployment gate

1. Freeze the release commit and record its Git SHA plus immutable image digests. Require hosted CI
   green for that exact SHA, no high/critical releasable vulnerability, and a clean worktree.
2. Confirm DNS, CA chain/hostname/expiry, renewal alert, firewall, host patch level, disk capacity,
   secret versions, least-privilege file permissions, and monitoring contacts. Do not record values.
3. Run `scripts/check-production-secrets.sh`, `scripts/check-staging-config.sh`, migration checks,
   and `scripts/staging-smoke.sh https://HOST`.
4. Quiesce writers for the backup window and create the verified pre-deployment backup using
   `BACKUP_WRITERS_QUIESCED=true scripts/backup-postgres.sh`. Record artifact identifier, checksum,
   encrypted external location, and restore-drill result.
5. Review forward migrations and previous-image schema compatibility. Confirm rollback revision,
   owner, decision threshold, and maintenance/communication window.
6. Require every UAT scenario and all three approvals in `UAT_CHECKLIST.md`. Any blocking failure
   or missing evidence makes the decision NO-GO.

## Deploy

Render the merged configuration first and verify only edge ports are published:

```sh
docker compose -f compose.yaml -f compose.production.yaml -f compose.staging.yaml config --quiet
docker compose -f compose.yaml -f compose.production.yaml -f compose.staging.yaml pull
docker compose -f compose.yaml -f compose.production.yaml -f compose.staging.yaml up -d
docker compose -f compose.yaml -f compose.production.yaml -f compose.staging.yaml ps
```

Deploy within the approved window. Do not use floating source code, edit a running container,
print the rendered environment, or expose PostgreSQL/API ports. Migrations are forward-only and
must finish before writers resume. If an image is built on-host, record its digest and prove it
matches the reviewed source; registry-pinned digests are preferred.

## Post-deployment verification

Run the HTTPS smoke script, inspect container health and capped fresh logs, confirm the latest
migration version, then execute one collector canary and one dry-run notification through their
normal durable queues. Verify active-feed lifecycle integrity, no duplicate record/event/delivery,
queue age stable or falling, source health recovering, and audit actor/action metadata. Observe
error rate, latency, readiness, CPU, memory, disk, database connections, collector lag, and
dead-letter counts through the agreed stabilization window.

Record only statuses, timestamps, request/run/delivery/audit IDs, metrics, commit, image digest,
and secret version identifiers. If verification fails, stop expansion and apply the rollback
criteria below; never bypass RBAC or repair durable state with ad-hoc SQL.

## Rollback

Rollback is required for readiness failure beyond the approved threshold, repeated 5xx, broken
authorization/audit, lifecycle leakage, data corruption/loss, secret exposure, or uncontrolled
queue duplication. Pause only affected writers, preserve logs and IDs, and redeploy the previous
immutable images/config if compatible with the current forward schema. Validate HTTPS health,
readiness, integrity, and a canary after rollback.

Do not reverse migration files or restore the database merely to roll back application code. A
database restore targets a clean database and follows `BACKUP_RESTORE.md`; use it only for proven
data loss/corruption with incident-owner approval. Rotate exposed credentials immediately and
follow the failure-recovery severity/escalation process.

## Production promotion

Promotion requires completed staging UAT, hosted CI, seven-day source stability, a valid recent
backup/restore drill, security/TLS evidence, operational monitoring, assigned on-call ownership,
and zero unresolved release blockers. Re-run preflight against production DNS/certificate/firewall
and deploy the exact approved artifacts. After the stabilization window, the release commander
records GO completion or rollback, evidence locations, remaining risks, and follow-up owners.

Production is not live merely because containers start. It is live only after HTTPS smoke,
readiness, integrity/canary checks, observability, approvals, and rollback readiness are evidenced.
