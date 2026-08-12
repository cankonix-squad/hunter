# Staging deployment

Staging is the release-candidate environment for integration checks and UAT. It uses the base,
production-secret, and staging TLS overlays together. This repository configures the deployment
boundary; DNS records, firewall rules, certificate issuance, and the actual host remain operator
owned and must be evidenced before promotion.

## Host prerequisites

- A dedicated Linux host with Docker Engine and Compose v2, patched and restricted to operators.
- An A/AAAA record whose hostname exactly matches `CANKONIX_PUBLIC_HOST`.
- Inbound TCP 80 and 443 only; SSH restricted to the administration network. PostgreSQL and all
  application containers remain unpublished.
- A trusted CA certificate and private key stored outside the repository with least-privilege
  ownership. Port 80 exists only for HTTPS redirect and certificate-renewal integration.
- The production secret files described in `SECRETS.md`.

The edge is the only trusted HTTP proxy. It overwrites inbound forwarding headers with the peer
address instead of trusting client-supplied `X-Forwarded-*` values. If a managed load balancer or
CDN is later placed in front, restrict edge ingress to its documented networks and introduce an
explicit `set_real_ip_from` allowlist; never enable recursive trust for arbitrary addresses.

## Validate and deploy

Set file-path variables without putting their contents in shell history:

```text
CANKONIX_PUBLIC_HOST=staging.example.internal
CANKONIX_TLS_CERTIFICATE_FILE=/secure/path/fullchain.pem
CANKONIX_TLS_PRIVATE_KEY_FILE=/secure/path/privkey.pem
CANKONIX_POSTGRES_PASSWORD_SECRET_FILE=/secure/path/postgres_password
CANKONIX_DATABASE_URL_SECRET_FILE=/secure/path/database_url
CANKONIX_JWT_SECRET_FILE=/secure/path/jwt_secret
```

Validate the merged configuration, then deploy the exact reviewed commit:

```sh
docker compose -f compose.yaml -f compose.production.yaml -f compose.staging.yaml config --quiet
docker compose -f compose.yaml -f compose.production.yaml -f compose.staging.yaml up -d
```

The edge enforces TLS 1.2/1.3, HTTP-to-HTTPS redirect, HSTS, CSP and existing browser-security
headers, a 1 MiB request limit, bounded proxy timeouts, 20 requests/second per address for the API,
and a stricter login rate. Do not enable HSTS on a hostname until valid HTTPS is continuously
available for it and its intended subdomains.

## Acceptance and rollback

Before UAT, verify certificate hostname/chain/expiry, TLS 1.2 and 1.3, rejection of TLS 1.0/1.1,
HTTP 308 redirect, `/healthz` and `/readyz` over HTTPS, security headers, login throttling, absence
of published database/API ports, container health, migration version, and a backup/restore drill.
Run `scripts/check-staging-config.sh` in CI and on the deployment host before rollout.

Rollback by redeploying the previous immutable commit and its compatible secret versions. Database
migrations are forward-only, so do not restore an older database merely to roll back application
containers. If TLS renewal fails, retain the last valid certificate, alert before expiry, and do
not expose an HTTP-only fallback.
