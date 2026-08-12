# Classification golden dataset approval

Classification approval is a product/domain decision, not an engineering toggle. The named
CANKONIX domain owner must check whether each title, procurement type, disposition, and category
reflects the intended opportunity policy and whether the dataset covers important false-positive
and false-negative risks.

## Prepare the review packet

From `services/api`, generate a deterministic Markdown packet:

```sh
go run ./cmd/evaluate -dataset testdata/classification_golden_v1.json -format review \
  > classification-review-v1.md
```

The reviewer marks every row APPROVE or proposes a corrected label/category, then comments on
coverage gaps such as mixed IT/non-IT scope, hardware/network ambiguity, consulting, training,
office supplies containing IT words, and new terminology. Keep the approved packet in the
controlled release evidence store; it contains no credentials and should not contain personal data.

## Record approval

Engineering applies the accepted corrections in one reviewed commit. Only after all rows and
coverage are accepted, change `review_status` to `approved` and add:

```json
"reviewed_by": "Full name and accountable role",
"reviewed_at": "2026-08-19T09:00:00Z",
"approval_reference": "controlled-evidence-or-ticket-id"
```

The validator rejects approval without all three fields, invalid/future timestamps, or approval
metadata on a candidate dataset. Reviewer identity and reference are audit metadata; never put a
password, token, email credential, or secret URL in them.

Run `scripts/classification-release-gate.sh`. It requires approved status and macro precision of at
least 80% across relevant, excluded, and needs-review labels. Attach its JSON report and commit SHA
to UAT. A later taxonomy/rule change requires a new dataset version and renewed domain approval;
do not silently reuse an approval for changed semantics.
