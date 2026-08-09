# MVP Roadmap

Target: 8 weeks after scope approval. Each week produces a demonstrable, tested increment.

## Week 1 — Foundation

- Repository and Docker Compose baseline
- React, Go API, PostgreSQL, queue, and Nginx skeletons
- Configuration and secret conventions
- Initial migrations, JWT/refresh flow, RBAC, audit foundation
- CI checks and observability skeleton

Exit: all services start locally; health/readiness and one authenticated API call work.

## Week 2 — Collector framework

- Source registry and source-policy administration
- Scheduler, durable job model, bounded retry, dead-letter handling
- Adapter interface, request limiter, fingerprints, source-run metrics
- First SPSE adapter with redacted fixtures

Exit: one source discovers candidates repeatedly without duplicating unchanged details.

## Week 3 — Canonical data

- Package, values, schedules, and change-event schema
- Normalization, lifecycle mapping, deduplication, idempotency
- Search indexes and archive/active separation

Exit: source observations reliably produce one canonical package with history.

## Week 4 — Opportunity workspace

- CANKONIX React shell and design tokens
- Feed, combined filters, saved view baseline, detail page
- Source evidence, timeline, ownership, and comments

Exit: users can find and review active packages end to end.

## Week 5 — Rules and AI

- IT taxonomy and deterministic prefilter
- AI adapter, structured schemas, prompt registry, cache
- Classification, extraction, summary, golden-set evaluation

Exit: eligible packages receive explainable enrichment; low-confidence items enter review.

## Week 6 — Decision workflow

- Versioned opportunity scoring
- Shortlist, pursue/pass, reason, assignment, next action
- Readiness checklist and user overrides
- Audit trail and initial pipeline analytics

Exit: a complete opportunity decision can be made and traced.

## Week 7 — Notifications and operations

- Alert rules, quiet hours, deduplication
- Email and webhook adapters
- Source-health console, lag/error/queue dashboards
- Replay and failure-recovery workflow

Exit: material events produce one traceable notification and operators can recover jobs safely.

## Week 8 — Hardening and release

- Security review, performance tests, parser regression suite
- Backup and restore drill
- UAT fixes, production Compose/deployment runbook
- Controlled rollout and seven-day source stability observation begins

Exit: production release candidate meets technical gates; final MVP acceptance follows the stability window.

## Initial backlog order

| Priority | Epic |
|---:|---|
| P0 | Platform baseline and CI |
| P0 | Identity, RBAC, and audit |
| P0 | Source registry and collector runtime |
| P0 | First adapter vertical slice |
| P0 | Canonical package and lifecycle |
| P0 | Active feed and detail workspace |
| P0 | Rules, AI classification, and evidence |
| P0 | Score and pursue/pass workflow |
| P1 | Notifications and source operations |
| P1 | Analytics and export |
| P1 | Expand from one adapter to 20 priority sources |

