# Release preflight

`scripts/release-preflight.sh` is the final fail-closed GO/NO-GO command. Run it from the exact
release checkout on the deployment host. It does not create approval or evidence; it verifies that
external evidence references are present and then reruns repository, classification, stability,
and HTTPS staging gates.

Provide these non-secret environment values:

- `RELEASE_COMMIT` and `HOSTED_CI_COMMIT`, both equal to checked-out `HEAD`;
- `HOSTED_CI_URL` and `STAGING_URL`, both HTTPS;
- `UAT_PRODUCT_APPROVAL_REF`, `UAT_OPERATIONS_APPROVAL_REF`, and
  `UAT_SECURITY_APPROVAL_REF` pointing to controlled signed evidence;
- `BACKUP_RESTORE_EVIDENCE_REF`, `ONCALL_OWNER`, and a tested `ROLLBACK_COMMIT`;
- immutable `API_IMAGE_DIGEST`, `COLLECTOR_IMAGE_DIGEST`, `NOTIFIER_IMAGE_DIGEST`, and
  `WEB_IMAGE_DIGEST` in `sha256:<64 lowercase hex>` form;
- corresponding `API_IMAGE_REF`, `COLLECTOR_IMAGE_REF`, `NOTIFIER_IMAGE_REF`, and `WEB_IMAGE_REF`
  in `registry/name@sha256:<digest>` form, each ending in its declared digest;
- `STABILITY_END_DATE` when evaluating a date other than the current UTC date;
- secret file path variables required by the production/staging overlays and read-only
  `DATABASE_URL_FILE`. Their values must never appear in approval references or command output.

The command also requires a clean worktree and configured `origin`, validates all local Compose/TLS/
runbook contracts, requires an approved classification dataset, requires a passing seven-day
database report, and performs CA-verified HTTPS smoke against staging. Any missing/invalid input or
failed child gate exits nonzero and means NO-GO. A GO result is evidence for the release commander,
not permission to deploy outside the approved change window.

Current repository state intentionally returns NO-GO until a remote/hosted CI, signed UAT approvals,
classification approval, real image digests, staging URL, backup evidence, on-call ownership, and
the completed stability window exist.
