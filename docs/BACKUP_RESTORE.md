# PostgreSQL backup and restore

Production policy requires one encrypted daily PostgreSQL backup, 30-day retention,
and a restore drill before release. Store backups outside the application host and
restrict access to the operations role. The repository scripts create and verify the
database artifact; encryption, remote transfer, retention deletion, and scheduling
belong to the production backup platform and must not rely on repository credentials.

## Create a backup

Use a PostgreSQL client with the same major version as the server. Point
`DATABASE_URL` at the source and choose a protected destination. Pause the API write
window, collector, and notifier first; confirm no writer remains, then acknowledge the
stable dataset explicitly:

```bash
BACKUP_WRITERS_QUIESCED=true DATABASE_URL='postgres://...' \
  ./scripts/backup-postgres.sh /protected/backups
```

The script writes a compressed custom-format dump through a temporary file and renames
it only after success. It produces three files with one UTC timestamp:

- `.dump`: the restore artifact;
- `.sha256`: checksum for corruption detection;
- `.manifest`: migration state and row count for every public table.

Transfer all three together. Encrypt at rest with the production backup service and
test retrieval independently. Never commit a dump; it may contain user and operational
data. Resume writers only after all three artifacts exist and the checksum file matches.

## Restore drill

Create a new empty disposable database. The drill refuses a target containing any
public table and never drops or cleans objects:

```bash
RESTORE_DATABASE_URL='postgres://.../empty_restore_target?sslmode=require' \
  ./scripts/restore-drill.sh /protected/backups/cankonix-oi-TIMESTAMP.dump
```

It verifies SHA-256 and archive readability before restore, uses `--exit-on-error`,
checks migration count/max version, compares every public-table count with the backup
manifest, and validates core opportunity foreign-key/lifecycle integrity.

After a drill, run application readiness and a read-only feed canary against the restored
database, record start/end times, artifact timestamp, database version, checksum, row
counts, RPO/RTO achieved, owner, and deletion evidence for the disposable target.

## Production recovery rules

- Restore into a new database/cluster; never overwrite the only copy of an incident
  database.
- Stop affected writers and preserve evidence before cutover.
- Confirm the application version is compatible with the restored migration version.
- Rotate credentials used during recovery and audit the cutover.
- Do not delete the old database until integrity, health/readiness, feed lifecycle,
  login/RBAC, collector idempotency, and notification deduplication checks pass.
- Escalate checksum mismatch, partial archive, missing manifest, schema mismatch, or
  count/integrity mismatch; do not waive these failures during a production restore.
