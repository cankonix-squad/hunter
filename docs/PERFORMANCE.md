# Performance and pagination gate

The MVP release budget targets an internal deployment with 10,000 canonical
opportunities. The authenticated opportunity feed must sustain 40 mixed-page requests
with 8 concurrent workers at p95 no slower than 750 ms on the CI PostgreSQL service.
This intentionally includes JWT verification, HTTP serialization, total-count query,
classification/score joins, and page retrieval.

The integration gate seeds deterministic data using set-based SQL, runs `ANALYZE`, and
verifies before measuring:

- first and last 100-row pages contain the expected rows and page totals;
- API requests above the documented maximum are capped at 100 rows;
- the needs-review filter returns the expected 20% subset;
- an exact package-title search returns one canonical identity;
- 40 requests distributed over pages 1–40 complete successfully with 8 workers.

Run it against an empty, fully migrated disposable database:

```bash
cd services/api
TEST_DATABASE_URL='postgres://...' \
  go test -race -count=1 -run TestOpportunityPaginationLoadBudget \
  -tags=integration -v ./internal/httpapi
```

On 12 August 2026, PostgreSQL 16.10 on the local Docker environment produced p50
275 ms, p95 390 ms, and maximum 408 ms with the race detector enabled. The 750 ms
CI limit leaves room for shared-runner variance while still detecting material query
regressions. Results are comparative release evidence, not an internet capacity claim.

The current feed uses bounded offset pagination (`page_size <= 100`). Before the corpus
or requested navigation grows substantially beyond the tested 10,000-row envelope,
retest with production-like hardware and consider stable cursor pagination. Run a
separate staging soak test for TLS, network latency, connection limits, and realistic
user traffic before production rollout.
