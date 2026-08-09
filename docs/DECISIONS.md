# Architecture and Product Decisions

## Locked

| ID | Decision |
|---|---|
| D-001 | Working name: CANKONIX Opportunity Intelligence (COI) |
| D-002 | React + TypeScript, Go REST API/workers, PostgreSQL, Docker, Nginx/TLS |
| D-003 | Internal decision-support product first; transactions remain on official SPSE |
| D-004 | Provider-agnostic AI with structured, explainable output |
| D-005 | CANKONIX monochrome palette with restrained semantic status colors |
| D-006 | MVP target is eight weeks after scope approval |
| D-007 | Collector does not bypass CAPTCHA, authentication, rate limits, or access controls |

## Proposed defaults requiring review

| ID | Topic | Proposed default | Confirmed? |
|---|---|---|---|
| R-001 | Initial sources | Top 20 LPSE by CANKONIX IT relevance and observed volume | No |
| R-002 | Default value view | Configurable; initial saved view Rp0–Rp1.5 billion | No |
| R-003 | Notification channels | Email + webhook for MVP | No |
| R-004 | Queue | Redis-compatible durable queue; benchmark PostgreSQL-only alternative | No |
| R-005 | Snapshot retention | Metadata/hash long term; raw safe response up to 30 days | No |
| R-006 | AI model | Select using precision, structured-output reliability, latency, and cost benchmark | No |
| R-007 | Deployment target | Single controlled Linux host for MVP, with external backup | No |

## Decision record template

```markdown
### ADR-NNN — Title

- Status: proposed | accepted | superseded
- Date: YYYY-MM-DD
- Context:
- Decision:
- Consequences:
- Alternatives considered:
```

