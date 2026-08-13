# CANKONIX Opportunity Intelligence

Internal opportunity intelligence platform for discovering, normalizing, classifying, scoring, and monitoring public procurement packages from SPSE/LPSE sources.

Repository: https://github.com/cankonix-squad/hunter

> Build. Integrate. Automate.

## Product boundary

COI is a decision-support system. It collects only publicly available information, provides a canonical link to the official SPSE page, and never performs supplier login, registration, document upload, bid submission, CAPTCHA solving, or access-control bypass.

## Technology baseline

- Frontend: React + TypeScript
- Backend and workers: Go
- Database: PostgreSQL
- Queue: Redis-compatible queue; PostgreSQL queue may be evaluated for a smaller MVP footprint
- Runtime: Docker Compose for local and staging
- Edge: Nginx, TLS, security headers, and request limits
- AI: provider-agnostic structured-output adapter
- Observability: metrics, structured logs, traces, dashboards, and alerts

## Repository target

```text
.
├── AGENTS.md
├── README.md
├── .env.example
├── compose.yaml
├── apps/
│   └── web/
├── services/
│   ├── api/
│   ├── collector/
│   └── ai-worker/
├── internal/
│   ├── domain/
│   ├── adapters/
│   └── platform/
├── migrations/
├── docs/
└── tests/fixtures/
```

## Documentation map

- [Living task list](tasklist.md)
- [Product requirements](docs/PRD.md)
- [System architecture](docs/ARCHITECTURE.md)
- [Crawling policy](docs/CRAWLING_POLICY.md)
- [AI agent design](docs/AI_AGENT.md)
- [MVP roadmap](docs/ROADMAP.md)
- [Architecture decisions](docs/DECISIONS.md)
- [Failure recovery runbook](docs/FAILURE_RECOVERY_RUNBOOK.md)
- [Security release review](docs/SECURITY_REVIEW.md)
- [Performance and pagination gate](docs/PERFORMANCE.md)
- [PostgreSQL backup and restore](docs/BACKUP_RESTORE.md)
- [Local acceptance](docs/LOCAL_ACCEPTANCE.md)
- [Codex execution rules](AGENTS.md)

## MVP completion signal

The MVP is ready when at least 20 priority SPSE sources run stably for seven days, the active feed excludes completed packages, classification precision reaches at least 80% on an approved golden dataset, recommendations are explainable, alerts are deduplicated, and security and restore checks pass.

## Local development target

The foundation stack now exists. Start it with:

```bash
cp .env.example .env
docker compose up --build
```

Expected endpoints:

- Web: `http://localhost:3000`
- API: `http://localhost:8080`
- Health: `http://localhost:8080/healthz`
- Readiness: `http://localhost:8080/readyz`

After the stack becomes healthy, verify the complete local runtime contract:

```bash
./scripts/local-acceptance.sh
```

The gate checks all six services, Docker health, public and protected HTTP behavior,
internal metrics, the current migration, 20 scheduled sources, lifecycle validity, and
the absence of recent collector dead letters.

The browser-facing edge is `http://localhost:3000`; it routes `/api/*`,
`/healthz`, and `/readyz` to the Go API. PostgreSQL is intentionally not
published to the host. The initial schema is applied only when a fresh local
database volume is created.

The reserved production application URL is `https://hunter.cankonix.com`.
This is a deployment target, not a live-status claim; production availability
still depends on the release preflight, stability gate, and hosted UAT approval.

## Current implementation

- React + TypeScript login and protected dashboard
- Browser session persistence, proactive access-token refresh, and logout
- Go HTTP service with structured logs, request IDs, safe recovery, graceful
  shutdown, and dependency-aware readiness
- PostgreSQL foundation for sources, durable source runs, and audit events
- JWT access tokens, rotating refresh sessions, reuse-family revocation, and RBAC
- Audited administrator source creation and authenticated source listing
- Audited source status/policy updates and durable source-run queue
- Independent collector worker with lease recovery, bounded retry, dead-letter,
  health degradation, and circuit breaker state
- Versioned fixture adapter with stable fingerprints and idempotent snapshots
- Policy-gated public SPSE listing adapter covering 20 controlled tenants with conservative request limits
- Canonical opportunity normalization with lifecycle-safe active feed and version history
- Versioned deterministic IT taxonomy with auditable include/exclude evidence
- Versioned five-dimension scoring with integer weights and reviewable breakdown
- Provider-neutral AI classification contracts with strict schema/evidence validation
- Versioned AI prompt, fingerprint cache, usage persistence, and cost-budget foundation
- Authenticated opportunity list/detail API with search, status filter, and pagination
- Relevance, minimum-score, and needs-review API filters with a review queue
- Responsive opportunity dashboard with official-source detail panel
- Classification evidence and five-dimension score breakdown in the opportunity UI
- Nginx reverse proxy with baseline security headers and request-size limit
- Docker Compose health checks and service dependency ordering
- OpenAPI contract for the implemented endpoints

The local runtime and automated quality suites are operational. Remaining MVP release
gates are the approved classification golden dataset and seven consecutive stable days
for the 20 priority sources; hosted UAT and production deployment are tracked separately.

## Useful commands

```bash
docker compose config
docker compose up --build
./scripts/local-acceptance.sh
curl http://localhost:3000/healthz
curl http://localhost:3000/readyz
curl http://localhost:3000/api/v1/system
```

CI runs Go formatting/tests/vet, frontend tests/build/audit, fresh-database
migrations, API authorization/contract integration and collector E2E tests, and
Compose image builds.
To verify migrations against an empty disposable PostgreSQL database, set
`DATABASE_URL` and run `./scripts/check-migrations.sh`. The integration suite uses the
same URL with `go test -race -tags=integration ./internal/httpapi` from `services/api`.

Open `http://localhost:3000` to use the browser login, or log in through the
API with the bootstrap administrator configured in `.env`:

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@cankonix.local","password":"your-password"}'
```

Bootstrap credentials create the first administrator only; changing those
environment values does not overwrite an existing user's password.

To rebuild a disposable local database from migrations, remove the Compose
volume explicitly and restart the stack. Never do this against staging or
production data.
