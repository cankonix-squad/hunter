# Seven-day source stability observation

The MVP gate requires 20 priority SPSE sources to operate for seven consecutive UTC calendar days
without daily intervention. The observation is historical database evidence: it cannot be shortened,
backfilled, or passed by repeatedly running a test on one day.

## Gate criteria

For an inclusive window ending on the requested UTC date, `scripts/stability-report.sh` requires:

- exactly 20 currently active sources whose policy has `schedule_enabled=true`;
- at least one successful persisted run for every source on each of seven dates (140 source-days);
- aggregate run success rate of at least 95%;
- zero dead-letter runs in the window;
- zero current consecutive failures and zero open circuit breakers;
- zero audited `source.updated` intervention during the window.

An upstream source returning no packages may still be a successful run; stability measures collector
operation and policy compliance, not the existence of procurement notices. Pausing, degrading, or
changing a source restarts the clean observation window. Never alter run/audit data to satisfy the
gate. Lifecycle filtering, classification precision, backup, security, UAT, and hosted deployment
remain separate release gates.

## Run and retain evidence

Use the read-only staging database credential and redirect the non-secret report to the controlled
release evidence store:

```sh
DATABASE_URL_FILE=/run/secrets/read_only_database_url \
  scripts/stability-report.sh 2026-08-18 \
  > stability-report-2026-08-18.txt
```

Prefer injecting `DATABASE_URL` directly into the process from the platform secret manager rather
than exporting or printing it. The report contains aggregate counts only. Record the deployed
commit, database/environment identifier, report checksum, operator, and UTC execution time next to
the artifact; do not commit generated reports or credentials.

Exit `0` and `status=PASSED` means the observation criteria passed. Exit `2` and
`status=INCOMPLETE` means seven real days or all 20 sources are not yet represented. Exit `1` and
`status=FAILED` means instability or invalid execution; investigate through the failure-recovery
runbook and begin a new clean window after correction.

## Daily review

Operators should review source health, scheduler/collector availability, safe errors, queue age,
dead letters, host/database capacity, and clock synchronization daily without changing source
configuration. An incident is not hidden to preserve the timer: stabilize it, record the event, and
restart the observation. On day seven, attach the passing report to UAT and release approval.
