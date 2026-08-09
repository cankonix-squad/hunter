# Codex execution rules

These rules apply to the whole repository unless a deeper `AGENTS.md` overrides them.

## Mission

Build CANKONIX Opportunity Intelligence as an internal, auditable decision-support platform. Optimize for early detection, relevance, traceability, safe collection, and operational reliability.

## Locked stack

- React + TypeScript frontend
- Go REST API and background workers
- PostgreSQL as system of record
- Docker-based local and staging environments
- Nginx/TLS at the edge
- Provider-agnostic AI adapter with structured JSON output

Do not introduce another primary language, database, frontend framework, or deployment model without recording an architecture decision.

## Product guardrails

- Only collect public content needed for opportunity monitoring.
- Never solve CAPTCHA, rotate identity to evade controls, bypass authentication, or circumvent rate limits.
- Never perform supplier registration, bidding, upload, or other SPSE transactions.
- Always retain canonical source URL, source timestamp, first seen, last seen, and change history.
- Human users own pursue/pass decisions. AI recommendations must be reviewable and overrideable.
- A completed, cancelled, evaluation, award, or contract-stage package must not appear in the active-registration feed.

## Engineering conventions

- Keep domain logic independent of HTTP, database, source adapters, and AI providers.
- Implement every SPSE variant behind a versioned source-adapter interface.
- Prefer incremental discovery and fingerprints; do not refetch unchanged detail pages.
- Make jobs idempotent and persist retries. Use bounded exponential backoff and dead-letter handling.
- Store timestamps in UTC and render in the source or user timezone.
- Use integer minor units or PostgreSQL `numeric` for currency; never floating point.
- Validate all external input and structured AI output at boundaries.
- Use database migrations that are forward-only and safe for staged deployment.
- Generate and maintain OpenAPI documentation for public service endpoints.

## Security baseline

- JWT access tokens with rotating refresh tokens and revocation support.
- RBAC and least privilege for users and service accounts.
- Secrets only through environment variables or a secret store; never commit credentials.
- Audit administrator changes and opportunity decisions.
- Enforce TLS, security headers, request-size limits, and safe document-type restrictions.
- Do not store irrelevant participant PII or execute downloaded content.

## AI baseline

- Rules run before AI for deterministic exclusions and cheap extraction.
- Every AI result includes schema version, model/provider, prompt version, confidence, and evidence.
- Low-confidence output enters review; it must not silently reject an opportunity.
- Auto-reject is allowed only for approved deterministic rules.
- Maintain a golden dataset and regression evaluation for classification and extraction.
- Cache by input fingerprint and enforce per-run and daily cost budgets.

## Required tests

- Go unit tests for domain and parsing logic
- Fixture-based parser regression tests per adapter/version
- Repository and migration integration tests against PostgreSQL
- API authorization and contract tests
- React component tests for critical decision flows
- End-to-end test for discover → normalize → classify → score → review
- AI schema, golden-set, and fallback tests
- Security, container, backup, and restore checks before production

## Definition of done for every change

- Acceptance criteria are satisfied and covered by tests.
- Logs and metrics expose failure modes without leaking secrets or sensitive data.
- Error and degraded states are visible to operators.
- Documentation and `.env.example` reflect new configuration.
- No unrelated refactor is bundled into the change.
- Formatter, linter, tests, migration checks, and container build pass.

## Implementation sequence

1. Establish repository, Compose, CI, configuration, auth, migrations, and observability.
2. Deliver one end-to-end source adapter as a vertical slice.
3. Add canonical data, history, classification, scoring, and the opportunity feed.
4. Validate operational behavior before expanding source coverage.
5. Add adapters in small batches with source-specific health and regression tests.

