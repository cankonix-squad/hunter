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

## Correction and selected public target

The initial review incorrectly treated one unavailable host/API path as a blocker
for all public-page collection. The agreed scope permits crawling public pages
without login, subject to source policy.

The first target is the Maluku tenant on spse.inaproc.id/malukuprov. Its
robots.txt explicitly allows /malukuprov, its listing is public, and the page
declares the DataTables route used by the adapter. Direct detail access returned
HTTP 403, so the adapter does not fetch or bypass detail pages.

## Decision

Do not attempt CAPTCHA handling, alternate identities, proxy rotation, hidden
endpoint discovery, or any other bypass. Do not register or use supplier
credentials for collection.

The initial approval permitted collection only against the reviewed Maluku listing routes. The
adapter uses one request at a time, a three-second minimum interval, a 20-second
timeout, a 1 MiB response limit, and stops on HTTP 403 or 429. It does not access
login, supplier transactions, CAPTCHA, or blocked detail content.

## Expansion review — 10 August 2026

The official `https://spse.inaproc.id/robots.txt` explicitly allows the tenant
paths for Jawa Tengah (`jatengprov`), DI Yogyakarta (`jogjaprov`), Banten
(`bantenprov`), Sumatera Utara (`sumutprov`), and Sulawesi Selatan
(`sulselprov`). Each corresponding public `/lelang` page returned HTTP 200 during
review. This approval remains limited to the public listing and declared
DataTables routes under the existing adapter safeguards; detail, authentication,
supplier transaction, and challenge routes remain out of scope.
