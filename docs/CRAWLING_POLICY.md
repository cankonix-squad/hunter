# Public Source Collection Policy

## 1. Purpose

Define safe, respectful, reproducible collection of public procurement metadata. This document is an engineering control, not legal advice. Source owners and current terms remain authoritative.

## 2. Allowed behavior

- Access public pages and public metadata needed for opportunity monitoring.
- Use a clear user agent and contact identifier when appropriate.
- Honor source-specific terms, robots policy, and operational instructions.
- Use incremental discovery, conditional requests, cache, fingerprints, and bounded schedules.
- Store canonical source URLs and observation timestamps.
- Disable or reduce collection per source when compliance or operations requires it.

## 3. Prohibited behavior

- Supplier login, credential use, registration, bidding, upload, or submission.
- CAPTCHA solving or human-verification bypass.
- Identity or fingerprint rotation intended to evade controls.
- Rate-limit evasion, proxy cycling, or access-control circumvention.
- Execution of downloaded files or unbounded storage of page content.
- Collection of unrelated personal data.

## 4. Source policy fields

Each source must define:

- adapter name and supported version
- enabled/disabled state and compliance owner
- timezone and allowed schedule window
- minimum interval between requests
- maximum concurrent requests per host
- request timeout and maximum response size
- retry count, backoff, and circuit-breaker threshold
- allowed paths/content types
- robots/terms review date
- safe contact and user-agent value

## 5. Default conservative limits

These are initial defaults, not permission to exceed a source policy:

| Control | Default |
|---|---|
| Per-host concurrency | 1 |
| Minimum request interval | 2 seconds plus jitter |
| Request timeout | 20 seconds |
| Retries | 2 for transient failures |
| Backoff | Exponential, bounded at 15 minutes |
| Circuit open | 5 consecutive access or parser failures |
| Maximum HTML response | 5 MiB |
| Maximum document metadata probe | 25 MiB; do not download by default |
| Raw response retention | Up to 30 days unless audit requires less or more |

## 6. Human verification and blocking

When a source returns CAPTCHA, access denied, repeated 429, or equivalent verification:

1. Stop automated requests to the affected route.
2. Mark the source degraded or paused.
3. Record status code, safe diagnostic metadata, and time.
4. Notify the operator without storing challenge content unnecessarily.
5. Resume only after the policy window or an authorized manual review.

The system must never automatically attempt to defeat the challenge.

## 7. Adapter contract

An adapter must expose discovery, detail parsing, lifecycle mapping, fingerprinting, and health probes. It must also provide redacted fixtures, schema validation, parser metrics, and an explicit supported-version range.

The SPSE public-list adapter additionally bounds each run to eight keywords, ten
pages per keyword, and fifty rows per page. Its deployed Maluku policy is more
conservative: seven keywords, three pages per keyword, twenty rows per page, one
request at a time, and a three-second minimum request interval. HTTP 403 and 429
stop collection immediately; only transport errors and HTTP 5xx responses are
retried.

The first controlled source batch covers the robots-allowed tenants `malukuprov`,
`komdigi`, `jakarta`, `jabarprov`, and `jatimprov`. Expansion remains batch-based;
an allowed robots path is necessary but does not by itself authorize higher rates
or broader routes. Listing matches are verified against whole words in the package
title because the server-side search may also match agency or hidden columns.

The second controlled batch adds `jatengprov`, `jogjaprov`, `bantenprov`,
`sumutprov`, and `sulselprov`. On 10 August 2026, the official robots policy
explicitly allowed every tenant path and each public `/lelang` page returned HTTP
200. The same conservative limits apply: listing/DataTables routes only, one
request at a time, a three-second interval, and bounded keyword pagination.

The third controlled batch adds `sumbarprov`, `sumselprov`, `kalbarprov`,
`kaltimprov`, and `baliprov`. On 11 August 2026, every path remained explicitly
allowed by the official robots policy and every public `/lelang` page returned
HTTP 200. Collection remains limited to the same reviewed listing contract and
conservative request budget.

## 8. Change management

- Parser changes require fixture tests for the previous and new structure.
- A sudden increase in empty critical fields opens the circuit.
- Operators can canary a new adapter version on one source run.
- Rollback changes the adapter version; it does not erase historical observations.

## 9. Data minimization

Persist normalized public procurement metadata, evidence snippets, hashes, and the canonical link. Avoid participant PII and avoid storing entire responses when a field-level snapshot is sufficient.
