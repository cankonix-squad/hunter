# Local acceptance

Local readiness is independent from production approval. Run `scripts/local-acceptance.sh` after
`docker compose up --build -d`. It requires all six local services running; PostgreSQL, API, and
collector healthy; UI/health/readiness HTTP 200; protected API HTTP 401 without authentication;
schema migration 29; exactly 20 active scheduled sources; internal API/collector metrics healthy;
and no source dead-letter during the last hour.

The notifier is expected to stay running but disabled unless a local email/webhook dry-run is being
tested. External SPSE HTTP 403/429 remains a safe source-policy failure and is not bypassed. A local
PASS proves the installed stack is operational now; it does not replace elapsed stability,
classification approval, hosted UAT, or production infrastructure gates.
