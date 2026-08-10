# System Architecture

## 1. Architectural style

Start as a modular monolith plus independently scalable workers. Preserve clear domain boundaries so collectors and AI workloads can scale without prematurely fragmenting the core API.

## 2. Logical components

| Component | Responsibility |
|---|---|
| React web | Opportunity feed, review workspace, operations console, administration |
| Go API | Auth, RBAC, domain services, search, workflow, audit, OpenAPI |
| Scheduler | Creates bounded source-run jobs according to policy |
| Collector worker | Discovery, fingerprinting, enrichment, and snapshots |
| Normalizer | Maps source payloads to canonical package and lifecycle records |
| Rule engine | Deterministic exclusions, eligibility, and score inputs |
| AI worker | Classification, extraction, summarization, and change analysis |
| Notification worker | Deduplicated digest and immediate delivery |
| PostgreSQL | Transactional state, history, search, jobs, audit, and configuration |
| Redis-compatible queue | Durable job coordination for MVP; optional after benchmark |
| Nginx | TLS termination, routing, limits, and security headers |

## 3. Processing flow

1. Scheduler creates a source run.
2. Collector fetches the public index under the source policy.
3. Fingerprints identify new or changed candidates.
4. Detail enrichment produces a source snapshot.
5. Normalizer upserts the canonical package and append-only changes.
6. Deterministic rules calculate lifecycle and exclusions.
7. AI worker enriches eligible packages using a versioned structured schema.
8. Scoring service calculates the opportunity score and factor breakdown.
9. Feed/search exposes the opportunity to authorized users.
10. Notification worker evaluates alert rules and deduplication keys.

## 4. Domain modules

- `identity`: users, roles, sessions, service accounts
- `sources`: registry, policy, adapter version, health, source runs
- `packages`: canonical record, values, schedules, documents, changes
- `classification`: taxonomy, requirements, AI runs, user corrections
- `opportunities`: score, owner, decision, checklist, comments
- `notifications`: rules, events, delivery, quiet hours
- `operations`: metrics projections, dead letters, replay, audit

## 5. Core data model

| Entity | Key data |
|---|---|
| `sources` | name, base URL, timezone, adapter, policy, status, health score |
| `source_runs` | start/end, status, pages, candidates, errors, next run |
| `source_snapshots` | package identity, fingerprint, observed time, safe raw metadata |
| `packages` | source, code, title, canonical URL, method, category, lifecycle |
| `package_values` | pagu, HPS, currency, observed time |
| `package_schedules` | stage, start, end, observed time |
| `requirements` | type, normalized value, evidence, confidence |
| `classifications` | taxonomy, is-IT, confidence, evidence, model version |
| `opportunities` | state, owner, decision, reason, next action, due time |
| `opportunity_scores` | total, factors JSON, policy version, calculation time |
| `change_events` | field, before, after, detected time, source run |
| `notifications` | event, rule, channel, recipient, status, dedupe key |
| `audit_logs` | actor, action, object, metadata, time |
| `saved_opportunity_views` | user owner, name, validated combined filters, timestamps |
| `opportunity_assignments` | opportunity owner, assigning actor, assignment timestamps |
| `opportunity_comments` | opportunity, author, immutable internal comment, timestamp |

## 6. API surface

- `GET /api/v1/opportunities`
- `GET /api/v1/opportunities/{id}`
- `PATCH /api/v1/opportunities/{id}`
- `POST /api/v1/opportunities/{id}/feedback`
- `GET /api/v1/sources`
- `GET|POST /api/v1/saved-views`
- `DELETE /api/v1/saved-views/{id}`
- `GET|PATCH /api/v1/opportunities/{id}/assignment`
- `GET /api/v1/team`
- `GET|POST /api/v1/opportunities/{id}/comments`
- `POST /api/v1/admin/sources`
- `POST /api/v1/admin/source-runs`
- `POST /api/v1/alert-rules`
- `GET /api/v1/analytics/pipeline`
- `GET /healthz`
- `GET /readyz`

## 7. Reliability model

- Jobs carry an idempotency key and source-policy version.
- External fetches use timeout, per-host concurrency, conditional request, bounded retry, and circuit breaker.
- A poison job moves to dead letter with a safe replay action.
- Database state changes and outgoing events use a transactional outbox.
- Source adapter failures degrade one source, not the API or other collectors.
- Raw response retention is bounded; normalized records and hashes are long-lived.

## 8. Search approach

Use PostgreSQL full-text search and `pg_trgm` for MVP. Introduce a separate search engine only after measured query volume or ranking needs exceed the PostgreSQL target.

## 9. Deployment environments

| Environment | Purpose | Minimum services |
|---|---|---|
| Local | Developer workflow | web, API, collector, AI worker, PostgreSQL, queue |
| Staging | Parser regression, integration, UAT | full stack, test notification, controlled AI budget |
| Production | Internal operation | full stack, TLS, backups, monitoring, alerting |

## 10. Initial architecture decisions

Decisions and unresolved choices are tracked in [DECISIONS.md](DECISIONS.md). Implementation must not silently decide a listed review item.
