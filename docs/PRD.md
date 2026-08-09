# Product Requirements — CANKONIX Opportunity Intelligence

Status: Draft for product and engineering review  
Version: 1.0  
MVP target: 8 weeks after scope approval

## 1. Problem

Relevant IT procurement packages are distributed across many SPSE/LPSE sources. Manual checking is slow, source structures change, package status can be stale, and raw search results contain completed, non-IT, commercially unsuitable, or invitation-only packages.

## 2. Product outcome

Provide CANKONIX with an early, explainable, and auditable opportunity feed that reduces manual source checking and shortens the time from first detection to pursue/pass decision.

## 3. Success metrics

| Outcome | MVP indicator |
|---|---|
| Early detection | At least 90% of relevant packages available to public collectors are detected within 60 minutes of a source change |
| Lower noise | At least 80% of priority-feed items are judged relevant after 30 days of tuning |
| Faster decision | Median first-seen to pursue/pass decision is under 4 working hours |
| Traceability | 100% of recommendations retain source, timestamp, score factors, and change history |
| Safe operation | No CAPTCHA, authentication, rate-limit, or access-control bypass |

## 4. Users

| Role | Primary need |
|---|---|
| Business Development | Find, shortlist, assign, and follow up on opportunities |
| Bid Manager | Verify eligibility and decide pursue/pass |
| Technical Lead | Assess scope, effort, staffing, and delivery risk |
| Executive | Review pipeline value, conversion, and bottlenecks |
| Administrator | Manage users, roles, sources, taxonomy, and policies |
| Operator | Monitor source health, lag, errors, queues, and recovery |

## 5. MVP scope

### 5.1 Source registry and collection

- Registry of SPSE sources, institution type, base URL, timezone, policy, status, and health.
- Scheduled incremental discovery with per-host limits, cache/fingerprint, retry, circuit breaker, and bounded concurrency.
- Source-specific, versioned adapters with fixture-based regression tests.
- Detail enrichment only for new, changed, or prioritized candidates.

### 5.2 Canonical package record

- Normalize package code, title, source, agency, work unit, procurement type, method, category, status, dates, location, pagu, and HPS.
- Deduplicate by source and package code, with a secondary content fingerprint.
- Persist append-only change events for status, schedule, values, method, and documents.
- Canonical lifecycle: discovered → registration_open → submission_open → evaluation → award → contract → completed/cancelled.

### 5.3 Active opportunity workspace

- Feed and full-text search with procurement type, source, region, category, value, status, deadline, and score filters.
- Detail page with summary, evidence, timeline, score breakdown, requirements, owner, checklist, and comments.
- Shortlist, pursue/pass, owner, next action, due date, and required decision reason.
- Saved views and export that honor filters and access controls.

### 5.4 Classification and scoring

- Rules exclude completed/inactive packages and enforce configured value/access constraints.
- AI classifies IT relevance and capability group, extracts requirements, summarizes risk, and explains changes.
- Score 0–100 combines capability fit, eligibility, commercial fit, time readiness, strategic value, and delivery risk.
- Users can override results; every override is audited.

### 5.5 Notification and operations

- Scheduled digest and immediate alert for priority packages, material changes, and approaching deadlines.
- MVP channels: email and webhook.
- Alert deduplication and quiet hours.
- Source health, crawl lag, parse errors, queue depth, retries, dead letters, AI latency, and AI cost visibility.

## 6. Non-goals

- Supplier login, tender registration, bidding, document upload, or submission.
- Guaranteed complete access to every SPSE under all network or access conditions.
- CAPTCHA solving, identity rotation, or security-control evasion.
- Automatic legal or final commercial decisions.

## 7. Functional requirements

| ID | Requirement | Priority |
|---|---|---|
| FR-01 | Manage source registry, policy, status, and health | Must |
| FR-02 | Run scheduled incremental discovery | Must |
| FR-03 | Normalize packages and retain immutable change events | Must |
| FR-04 | Exclude non-active lifecycle stages from active feed | Must |
| FR-05 | Classify IT relevance with confidence and evidence | Must |
| FR-06 | Produce explainable opportunity scores | Must |
| FR-07 | Support shortlist, pursue/pass, assignment, and comments | Must |
| FR-08 | Configure deduplicated alerts and digests | Must |
| FR-09 | Provide fast full-text and combined filters | Must |
| FR-10 | Export filtered results with canonical links | Should |
| FR-11 | Capture relevance and classification corrections | Should |
| FR-12 | Audit administration and business decisions | Must |

## 8. Non-functional requirements

| Area | MVP target |
|---|---|
| Availability | 99.5% for application; one degraded collector cannot take down the system |
| Search | P95 under 1 second at MVP data volume |
| Dashboard | P95 under 2 seconds |
| API reads | P95 under 500 ms excluding external fetches and AI jobs |
| Freshness | Standard discovery within 60 minutes; priority sources configurable to 15–30 minutes |
| Security | TLS, JWT rotation, RBAC, least privilege, audit, secret management |
| Resilience | Idempotency, persisted queue, bounded retries, dead-letter queue, circuit breaker |
| Backup | Daily PostgreSQL backup, 30-day retention, tested restore before production |

## 9. Acceptance scenarios

- A newly detected active IT package that fits the saved value range appears in the feed with source evidence and a score.
- A package already in evaluation remains searchable in archive but is absent from the active feed.
- An invitation-likely non-tender package is explicitly labeled and separated from open opportunities.
- A material deadline change creates a change event, updates readiness, and sends one alert.
- Human verification or blocking response stops the collector safely and marks the source degraded.
- A low-confidence AI result enters review and retains deterministic evidence.
- Duplicate observations produce one canonical opportunity with source history.

## 10. MVP definition of done

- 20 priority sources operate for 7 days without daily intervention.
- The active feed reliably filters lifecycle state, IT category, and configured pagu/HPS range.
- Classification reaches at least 80% precision on an approved golden dataset.
- Score factors are visible and authorized users can override them.
- Alerts are deduplicated and traceable to their triggering event.
- RBAC, token rotation, audit, TLS, backup/restore, monitoring, and parser regression tests pass.

