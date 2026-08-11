# Failure recovery runbook

This runbook covers the local/staging Compose deployment. Use application APIs for
recovery actions and PostgreSQL only for read-only diagnosis. Record times in UTC and
include the source timezone when it matters.

## Severity and ownership

| Severity | Examples | Response |
|---|---|---|
| SEV-1 | Data loss/corruption, exposed credentials, all collection unavailable, or active-feed lifecycle integrity failure | Stop the affected writer, notify the incident owner immediately, preserve evidence, and require explicit approval before resuming |
| SEV-2 | Several sources degraded, readiness failing, growing delivery backlog, or repeated dead letters | Assign an administrator/operator, stabilize in the current response window, and escalate if impact grows |
| SEV-3 | One source or delivery fails transiently without integrity impact | Triage during normal operations and monitor the bounded retry |

The incident owner coordinates recovery. Administrators may change source status and
policy. Administrators and operators may queue source runs and replay notification dead
letters. Never record tokens, passwords, SMTP credentials, webhook secrets, full external
response bodies, or irrelevant participant data.

## Detect and stabilize

Capture state before restarting anything:

```bash
date -u
docker compose ps
curl -fsS http://localhost:3000/healthz
curl -fsS http://localhost:3000/readyz
docker compose logs --since=30m --tail=300 api collector notifier postgres edge web
```

`/healthz` proves the API process is alive. `/readyz` also checks its dependency state.
Preserve request IDs, UTC timestamps, source/run or delivery IDs, capped safe errors,
response codes, and deployed commit/config version.

For a SEV-1 writer or integrity incident, stop only the affected worker while collecting
evidence:

```bash
docker compose stop collector
docker compose stop notifier
```

Stopping both is not the default. Keep unaffected paths available. Do not delete a
volume, truncate a table, edit queue rows, or roll a migration backward.

## Read-only diagnosis

The source console and `GET /api/v1/sources` expose status, health score, consecutive
failures, circuit-open time, last success/failure, latest run status/metrics/safe error,
lag, and pending-run count. With an access token:

```bash
curl -fsS http://localhost:3000/api/v1/sources \
  -H "Authorization: Bearer $ACCESS_TOKEN"
curl -fsS http://localhost:3000/api/v1/operations/dead-letters \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

The dead-letter endpoint requires administrator or operator role. HTTP 401/403 is an
authorization failure, not a reason to bypass RBAC.

When API data is insufficient, connect using credentials already inside PostgreSQL:

```bash
docker compose exec -T postgres sh -lc \
  'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -P pager=off'
```

Run only read-only queries during diagnosis:

```sql
SELECT version, applied_at
FROM schema_migrations
ORDER BY version DESC
LIMIT 5;

SELECT sr.id, s.name, sr.status, sr.attempts, sr.next_attempt_at,
       sr.error_code, sr.safe_error, sr.created_at, sr.finished_at
FROM source_runs sr
JOIN sources s ON s.id = sr.source_id
ORDER BY sr.created_at DESC
LIMIT 50;

SELECT channel, status, count(*) AS deliveries,
       min(created_at) AS oldest_created_at
FROM notification_deliveries
GROUP BY channel, status
ORDER BY channel, status;
```

Exit with `\q`. Treat `safe_error` as context, not proof of root cause. Correlate it
with structured logs and request/run IDs.

## Recovery procedures

### API, edge, or database readiness

1. Check `docker compose ps`, then `postgres`, `api`, and `edge` logs in that order.
   A healthy API behind a failing edge points to routing; live health with failing
   readiness usually points to a dependency.
2. Correct configuration or dependency availability without changing stored data.
3. Recreate only the affected service with `docker compose up --build -d <service>`.
4. Require both health endpoints to return HTTP 200 and confirm no repeating error in
   fresh logs.

Migrations are forward-only. Restore a previous application image/config only when it
is compatible with the applied schema; otherwise ship a forward fix. Database restore
belongs to the backup/restore procedure and requires incident-owner approval.

### Collector or source failure

1. Inspect source health and recent runs. `queued`/`running` are pending, `failed` is
   eligible for bounded retry, `dead_letter` is exhausted, and `succeeded` is complete.
   Expired worker leases recover automatically.
2. For upstream 403/429, CAPTCHA, authentication, or a changed access boundary, do not
   evade it. An administrator must set the source to `paused` or `degraded` through
   `PATCH /api/v1/admin/sources/{id}` and review its public-access policy/adapter.
3. A circuit opens after consecutive failures and blocks new work for its cooldown.
   Fix the cause and wait for `circuit_open_until`; never clear it directly in SQL.
4. Once valid and active, queue one canary through the durable API with a unique,
   incident-specific idempotency key:

```bash
curl -fsS -X POST http://localhost:3000/api/v1/admin/source-runs \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"source_id":"SOURCE_UUID","idempotency_key":"incident-YYYYMMDD-SOURCE-canary-1"}'
```

5. Verify the canary succeeds, lag falls, the queue does not grow, and repeated
   collection creates no duplicate snapshots, versions, alert events, or deliveries.
   Resume normal scheduling only after that check.

### Notification failure and dead-letter replay

1. Identify whether email, webhook, or both are affected. Review `notifier` logs and
   the operations dead-letter list. Quiet-hours events whose `available_after` is in
   the future are intentionally deferred.
2. Correct SMTP/network/recipient or HTTPS endpoint/signature configuration. Keep
   secrets in environment/secret storage. Redirects and non-2xx webhook responses fail
   by design.
3. In local acceptance only, `EMAIL_DRY_RUN=true` or `WEBHOOK_DRY_RUN=true` exercises
   the normal outbox without contacting the destination. Delivery-enabled and dry-run
   flags default to `false`; restore intended values after testing.
4. Replay only after fixing the cause, through the operations console or API:

```bash
curl -fsS -X POST \
  http://localhost:3000/api/v1/operations/dead-letters/DELIVERY_ID/replay \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Replay reuses the original delivery and alert-event identity, resets transient fields,
increments replay count, and writes an audit event. HTTP 404 means it is missing or no
longer dead-lettered; do not manufacture a new event. Replay a small canary first, then
process remaining known-good failures while monitoring queue age and destination health.

## Recovery verification and closure

Do not close the incident until applicable checks pass:

- `/healthz` and `/readyz` return HTTP 200 through the edge.
- Expected Compose services run and no error repeats in fresh logs.
- Queue counts and oldest pending age are stable or falling; no expired-lease loop.
- One source or delivery canary completes through its normal durable path.
- No duplicate source snapshot/version, alert event, or channel delivery was created.
- The active feed contains only registration-active states; completed, cancelled,
  evaluation, awarded, and contracted packages remain excluded.
- Source health/lag recovers and notification dead letters stop increasing.
- Temporary pauses, dry-run flags, diagnostic log levels, and config rollbacks are
  restored to their intended values.
- The record includes cause, impact window, actions, IDs, verification, and follow-up.

Escalate when the cause is unknown, failure returns after one controlled recovery,
integrity cannot be proven, a secret may be exposed, upstream access policy is unclear,
or recovery would require destructive database work.

## Incident record template

```text
Incident ID / severity:
Owner:
Detected at (UTC) / source timezone:
Impact and affected services/sources/channels:
Deployed commit and configuration version (no secrets):
Request, source-run, delivery, and audit IDs:
Observed safe errors and response codes:
Root cause:
Stabilization and recovery actions:
Canary and integrity verification:
Resolved at (UTC):
Follow-up owner and due date:
```
