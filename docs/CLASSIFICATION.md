# Deterministic Classification

Classification runs inside the collector transaction after canonical normalization
and before any future AI adapter. The result is tied to the opportunity fingerprint,
taxonomy version, and rule version so a decision can be reproduced and audited.

## Taxonomy `cankonix-it-v1`

| Code | Scope |
|---|---|
| `software-engineering` | Applications, information systems, and system integration |
| `data-ai` | Data integration/platforms, analytics, machine learning, and AI |
| `cloud-infrastructure` | Cloud platforms, DevOps, and container platforms |
| `cybersecurity` | Cybersecurity assessment and services |

The taxonomy and exact keyword rules are seeded by migration 005. Any semantic
change must use a new taxonomy or rule version; deployed versions must not be
silently reinterpreted.

## Outcomes and safety

- `relevant`: one or more include rules match and no exclude rule matches.
- `excluded`: one or more exclude rules match and no include rule matches.
- `needs_review`: no rule matches, or include and exclude evidence conflicts.

Confidence is stored as integer basis points (`0..10000`). Evidence records the
rule, keyword, source field, and matched field value. An excluded result only sets
`auto_rejected` when its deterministic rule is explicitly marked
`approved_auto_reject`; conflicting rules can never auto-reject.

This foundation does not remove opportunities from the active feed yet. Feed/API
filters and the human review queue belong to the classification/scoring UI slice.

## Golden-set regression evaluation

The versioned candidate baseline lives at `services/api/testdata/classification_golden_v1.json`.
It covers relevant categories, approved deterministic exclusions, unmatched cases,
and conflicting evidence. Run the regression gate from `services/api`:

```sh
go run ./cmd/evaluate -dataset testdata/classification_golden_v1.json -minimum-precision-bp 8000
```

The runner validates dataset and taxonomy versions, reports a confusion matrix,
per-label precision/recall, accuracy, failures, and macro precision. It exits with
status 1 when macro precision is below the configured threshold and status 2 when
the dataset or command is invalid. Dataset changes require human review; do not
change labels merely to make a regression pass.

The initial dataset has `review_status: candidate`. After domain-owner review, change
that status to `approved`; release CI can then add `-require-approved`. A candidate
result is useful for regression development but does not satisfy the MVP approval gate.
