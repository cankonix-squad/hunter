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
