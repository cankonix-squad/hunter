# Public Source Review

## LKPP / INAPROC review — 9 August 2026

The first candidate was LKPP's own LPSE/INAPROC infrastructure. It was not
enabled for collection.

Evidence reviewed:

- LKPP's official ISB page states that access to the INAPROC API Gateway
  requires an INAPROC account and a formal request from a K/L/PD:
  <https://isb.lkpp.go.id/isb-2>
- LKPP's official public-information portal describes national procurement
  profile data as public information and links to the public profile portal,
  but does not document an unauthenticated package-feed API:
  <https://ppid.lkpp.go.id/information/public/203/portal-pengadaan-nasional>
- The current SPSE terms are published under Keputusan Deputi II No. 4/2024:
  <https://jdih.lkpp.go.id/regulation/keputusan-deputi-ii/keputusan-deputi-ii-nomor-4-tahun-2024>
- Automated research fetches to `lpse.lkpp.go.id` returned HTTP 402, and a
  direct DNS lookup from the local runtime could not resolve the host.

## Decision

Do not attempt CAPTCHA handling, alternate identities, proxy rotation, hidden
endpoint discovery, or any other bypass. Do not register or use supplier
credentials for collection.

Collection may proceed when either:

1. CANKONIX receives authorized INAPROC API access and its permitted scope,
   rate limits, retention rules, and contact owner are recorded; or
2. an individual LPSE provides a reachable public listing whose current
   robots/terms permit conservative automated metadata access.

Until then, the fixture adapter remains the only enabled SPSE adapter. Technical
HTTP/parser work for a real source must be based on a redacted response obtained
through an allowed access path, not on an invented or reverse-engineered schema.
