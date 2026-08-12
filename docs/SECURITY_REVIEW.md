# Security release review

Review date: 12 August 2026. Scope: repository, Go/npm dependencies, application
runtime images, Compose exposure, authentication/RBAC boundaries, HTTP edge, collector,
and notification delivery. This review is a technical release gate, not a penetration
test or authorization to expose the local stack publicly.

## Automated gates

CI fails on:

- reachable known Go vulnerabilities reported by `govulncheck`;
- npm dependency vulnerabilities at high or critical severity;
- repository dependency, configuration, or secret findings at high/critical severity;
- fixable high/critical vulnerabilities in API and web runtime images;
- Go race/vet, frontend test/build, migration, API contract, E2E, or image-build failure.

The Trivy action is pinned to the full commit of verified release v0.36.0 rather than
a mutable tag. Scanner findings may be suppressed only with a documented identifier,
technical non-applicability argument, owner, and expiry date. There are currently no
such exceptions.

## Controls verified

- PostgreSQL has no host port; only the edge publishes port 3000 locally.
- API and collector use bounded request bodies, durable idempotent work, capped safe
  errors, leases, bounded retries, and circuit breaking.
- Access tokens are signed and short lived; refresh tokens rotate and reuse revokes the
  token family. Protected routes reject missing/invalid tokens and privileged mutations
  enforce role checks in integration tests.
- The edge disables version tokens, limits request bodies, and sets content-type,
  framing, referrer, and permissions headers. The API adds no-store and request IDs.
- Runtime Go processes use a non-root user. Secrets are environment-only and `.env` is
  ignored. Repository pattern review found no tracked private key or common token form.
- SMTP requires STARTTLS/TLS 1.2+, webhook delivery requires HTTPS and HMAC, does not
  follow redirects, and discards bounded response bodies.
- The collector does not authenticate to SPSE, solve CAPTCHA, evade rate limits, or
  bypass access controls.
- Reproducible builds use Go checksums and `npm ci`; supported base images are refreshed
  before the release scan.

## Residual release blockers

- Development defaults in `compose.yaml` are not production credentials. Production uses
  `compose.production.yaml`: file-mounted secrets, ambiguous-source rejection, and startup
  entropy checks prevent those defaults from crossing the production boundary. Secret-store
  provisioning and rotation evidence remain deployment responsibilities.
- The local edge remains HTTP-only. The staging overlay enforces TLS 1.2/1.3, HSTS, forwarding
  header overwrite, request limits, and per-address API/login throttling. Production DNS,
  firewall enforcement, CA certificate issuance/renewal, and hosted acceptance evidence remain
  deployment responsibilities.
- Production backup scheduling/encryption/external retention, monitoring/alerting, UAT,
  and seven-day source stability must pass before production acceptance. The repository
  restore drill and load gate are automated and passing.
- A qualified external penetration test is recommended before internet exposure and
  required after any material authentication or deployment-architecture change.

## Release procedure

Run the complete CI workflow against the exact release commit. Rebuild images without
cache when a base-image advisory changed, scan the resulting immutable artifacts, and
record their digests. Any high/critical fixable finding blocks release. Rotate a secret
immediately if exposure is suspected; never paste the value into an issue or log.
