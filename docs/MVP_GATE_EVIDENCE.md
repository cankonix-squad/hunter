# MVP release gate evidence

This matrix distinguishes repository-proven behavior from evidence that requires elapsed time or
human approval. A feature being implemented does not by itself satisfy a release gate.

| Gate | Status | Durable evidence | Remaining external evidence |
|---|---|---|---|
| 20 priority sources stable for seven days | Blocked | Database-backed `stability-report.sh` requires 140 source-days, at least 95% success, and no dead letter/intervention | Passing report after a real clean window |
| Active-feed lifecycle integrity | Passed | PostgreSQL query restricts active to `announced`/`registration_open`; integration E2E seeds every inactive lifecycle and rejects any leak | Repeat as hosted UAT scenario UAT-LIFECYCLE |
| Classification macro precision at least 80% | Blocked | Versioned evaluator reports 100% on 15 candidate cases and supports `-require-approved` | Domain owner must review labels/coverage and change dataset status to `approved` |
| Explainable and human-overridable recommendation | Passed | Detail contract returns classification evidence and five score dimensions; versioned override supports reason, RBAC, update/clear, audit, and separate model result | Hosted role UAT remains required for deployment approval |
| Alert deduplication | Passed | Database unique invariant on user/version/event plus idempotent matcher acceptance prevents overlapping-rule/retry duplicates | Hosted delivery UAT remains required for deployment approval |
| Security | Passed locally | Automated dependency, repository, image, authorization, secret, and TLS gates | Hosted CI and deployed TLS/security acceptance |
| Backup/restore | Passed locally | Checksum/count/migration/integrity restore drill on an empty target | Production schedule, encryption, retention, and recent drill evidence |

The overall release remains NO-GO while any row is blocked or its required hosted/UAT evidence is
missing. Dataset approval must be performed by a named CANKONIX domain owner; engineering must not
self-approve labels merely to turn the gate green.
