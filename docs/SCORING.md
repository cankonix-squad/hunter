# Explainable Scoring

Scoring runs after deterministic classification in the same collector transaction.
Each result is tied to an opportunity fingerprint, classification result, and
`explainable-v1` profile so it remains reproducible and reviewable.

## Version 1 dimensions

| Dimension | Weight | Purpose |
|---|---:|---|
| Capability fit | 40% | Deterministic match to the CANKONIX taxonomy |
| Value | 20% | Integer-minor-unit IDR value band |
| Location | 10% | Operational location availability and future fit rules |
| Deadline | 15% | Remaining response window at the source observation time |
| Risk | 15% | Classification ambiguity or deterministic exclusion risk |

Weights, raw dimension scores, contributions, total score, and confidence use
integer basis points (`0..10000`). Currency calculations use integer minor units;
no floating point is used for money.

## Review behavior

A result enters `needs_review` when deterministic classification is ambiguous or
score confidence is below 7000. Missing location or budget is represented as an
explicit `unknown` reason with a neutral/default raw score; the engine never
invents missing input. Each dimension stores its reason code, human-readable
explanation, and evidence.

Deadline scoring uses the source timestamp when available. This keeps repeat
evaluation of an unchanged source fingerprint deterministic instead of allowing
wall-clock time to silently change its score.
