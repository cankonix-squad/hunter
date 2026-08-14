# User acceptance checklist

Run UAT against the reviewed staging commit over trusted HTTPS. Use named test accounts for every
role and synthetic/public test records only. Record UTC time, tester, deployed commit, request or
audit IDs, expected result, actual result, and a redacted screenshot/link. Never paste passwords,
tokens, private source bodies, or notification secrets into evidence.

## Entry criteria

- CI is green for the exact commit and image digests are recorded.
- Staging smoke, migration verification, backup/restore drill, and security scan pass.
- At least administrator, operator, business development, bid manager, technical lead, executive,
  and viewer test accounts exist with isolated credentials.
- A rollback owner and decision deadline are assigned; no unresolved SEV-1/SEV-2 issue exists.

Provision isolated role accounts through the API image CLI, using a password file mounted as a
Compose secret. Never pass a password as an argument or environment value:

```sh
docker compose run --rm api provision-user \
  --email operator-uat@example.com --name "UAT Operator" \
  --roles operator --password-file /run/secrets/bootstrap_admin_password
```

Supported roles are `administrator`, `operator`, `business_development`, `bid_manager`,
`technical_lead`, and `executive`. Omit `--roles` for the viewer account. Provisioning is
create-only, rejects duplicate emails and unknown roles, and records `identity.user_provisioned`
in the audit log without password material.

## Acceptance scenarios

| ID | Tester | Scenario and expected result | Required evidence |
|---|---|---|---|
| UAT-ACCESS | All roles | Login succeeds; invalid credentials fail safely; protected API is 401 without a token; privileged actions return 403 for unauthorized roles; logout revokes the refresh session. | HTTP status/request IDs and role matrix |
| UAT-DISCOVER | Operator | One approved public-source canary discovers and normalizes a package once; repeating it creates no duplicate snapshot/version. Source URL, timestamps, health, and run metrics remain visible. | Source/run IDs and before/after counts |
| UAT-LIFECYCLE | Business development | Registration-active package appears in active feed; evaluation/award/contract/completed/cancelled records remain absent from active feed but searchable with an explicit lifecycle filter. | Filter values and opportunity IDs |
| UAT-REVIEW | Bid manager | Relevant item shows deterministic evidence, classification confidence, five score dimensions, change history, and canonical source URL; low-confidence item enters review rather than silent rejection. | Opportunity/version/classification IDs |
| UAT-WORKFLOW | Business development, bid manager, technical lead, executive | Shortlist, assignment, comment, pursue/pass reason, next action, readiness, and human override obey role policy, persist after reload, and appear in activity history. | Actor/action/audit IDs; no sensitive comment data |
| UAT-ALERT | Business development | Overlapping rules produce one alert for one event; quiet hours defer delivery; one email and webhook dry-run complete through the outbox without external delivery. | Alert event/delivery IDs and dedup count |
| UAT-OPERATIONS | Operator | Source degradation, queue state, bounded safe error, dead letter, and authorized replay are visible; canary recovery preserves identity and does not duplicate records. | Run/delivery/replay IDs and queue trend |
| UAT-AUDIT | Administrator | Source administration and every business mutation above show the correct actor, action, entity, UTC time, and safe metadata; secrets and tokens are absent. | Sample audit IDs across actions |
| UAT-PERFORMANCE | Product owner | Feed/search/filter/detail remain usable at the staged MVP dataset; measured API p95 stays below the documented release budget and no error spike occurs. | Load report, sample request IDs, dashboard observation |

Any data-integrity, authorization, lifecycle-filter, audit, secret exposure, or destructive-recovery
failure blocks release. Other deviations require an owner, severity, due date, and explicit product
and operations acceptance; security controls cannot be waived silently.

## Sign-off

```text
Environment URL / deployed commit / image digests:
UAT window (UTC):
Passed scenario IDs:
Failed or deferred scenario IDs and issue links:
Open risks, owners, and due dates:
Rollback owner and tested revision:
Product owner approval: name / UTC time / decision
Operations approval: name / UTC time / decision
Security approval: name / UTC time / decision
Production promotion decision: GO / NO-GO
```
