# Production domain

The reserved application URL is `https://hunter.cankonix.com`. Availability is fail-closed:
the domain is not considered live until a public client validates its trusted CA chain and receives
HTTP 200 from `/`, `/healthz`, and `/readyz`.

Run the reusable gate from outside the deployment host:

```sh
./scripts/check-production-domain.sh
```

## Current observation

On 13 August 2026 the hostname resolved to a Traefik endpoint, but it returned its self-signed
`TRAEFIK DEFAULT CERT` and HTTP 404. This proves DNS reaches a server but does not prove that a
router, trusted certificate, or CANKONIX service is attached. Production is therefore not live.

## Host integration required

The operator must choose and document one routing boundary:

1. Attach CANKONIX to the existing Traefik deployment using the host's real external network and
   certificate resolver; or
2. Route Traefik to the repository's hardened Nginx edge while keeping Nginx ports private to the
   host/network.

Do not bind the repository edge directly to host ports 80/443 while Traefik owns them. Do not use
the default certificate, disable certificate verification, or expose API/PostgreSQL ports as a
workaround. After routing and certificate issuance, run the domain gate, staging smoke, UAT, and
release preflight against the exact deployed commit.

For an existing Traefik host, the repository provides `compose.traefik.yaml`. Discover the actual
external Docker network, HTTPS entrypoint, and ACME certificate resolver from that host, then set:

```text
CANKONIX_PUBLIC_HOST=hunter.cankonix.com
CANKONIX_TRAEFIK_NETWORK=<existing external Docker network>
CANKONIX_TRAEFIK_ENTRYPOINT=<existing HTTPS entrypoint; defaults to websecure>
CANKONIX_TRAEFIK_CERT_RESOLVER=<existing ACME resolver>
```

Validate with `./scripts/check-traefik-config.sh`, then merge `compose.yaml`,
`compose.production.yaml`, `compose.traefik.yaml`, and `compose.release.yaml`. The overlays remove the local `3000:80`
publication, attaches only the edge to Traefik's external network, routes the production hostname,
ask the named resolver for TLS, disable host-side application builds, and require immutable image
references. PostgreSQL, API, collector, notifier, and web remain internal.

The Traefik overlay uses a dedicated internal Nginx configuration with production CSP, HSTS,
security headers, request-size limits, and bounded proxy timeouts. Traefik owns client-aware rate
limits: API traffic is limited to 20 requests/second with burst 40, while the higher-priority login
router is limited to 5 requests/minute with burst 3. The general web router is not subjected to the
API limit. `scripts/check-traefik-config.sh` validates router priority, middleware values, port
isolation, secret boundaries, and Nginx syntax.

On the deployment host, copy `.env.traefik.example` outside the repository or export the same
variables from the platform environment. Never put secret values in it. After creating secret files
with mode `0400` or `0600`, run `./scripts/traefik-host-preflight.sh`. It verifies the exact production
hostname, required non-empty secret files and permissions, existence of the external Docker network,
immutable `image@sha256` references, the merged Compose output, absence of build directives and
host-published ports, and absence of development credentials.
Only a passing host preflight authorizes `docker compose ... up -d`; it does not perform deployment.
