# Classification evaluation v2

Dataset: `classification-testcases-v2` (candidate). Provenance: synthetic, authored for regression; not independently labeled procurement samples.

Owner master terms: 131 discovery phrases across 12 categories. Discovery-only matches do not decide relevance.

Cases: 120. Label accuracy: 100.00%. Label plus category exact accuracy: 100.00%.

## Confusion matrix

Rows are expected; columns are predicted.

| Expected \ predicted | relevant | excluded | needs_review |
|---|---:|---:|---:|
| relevant | 50 | 0 | 0 |
| excluded | 0 | 40 | 0 |
| needs_review | 0 | 0 | 30 |

## Precision and recall

| Label | Precision | Recall | TP | FP | FN |
|---|---:|---:|---:|---:|---:|
| relevant | 100.00% | 100.00% | 50 | 0 | 0 |
| excluded | 100.00% | 100.00% | 40 | 0 | 0 |
| needs_review | 100.00% | 100.00% | 30 | 0 | 0 |

## Label or category mismatches

No mismatches on this synthetic candidate set.

## False positives and false negatives

A false positive predicts `relevant` for an expected `excluded` or `needs_review` case. A false negative expects `relevant` but predicts another label. Category mismatches are listed above separately.

| Kind | Case IDs |
|---|---|
| false positive | none |
| false negative | none |

## Production interpretation

These cases are synthetic and authored alongside the rules. Their metrics are regression feedback, not production readiness evidence. An independent reviewer must label a real procurement sample, including hard negatives and mixed scopes, before release. The collector and production classifier remain on v1.

## Coverage and next validation

The dataset retains 15 owner-approved v1 labels as a temporary baseline and covers every v2 category with positive, negative, and ambiguous cases. It includes keyword collisions, multiple categories, Barang IT work, KAK/TOR context, specialist training, and common abbreviation/typo variants. No real LPSE package has been independently labeled for this v2 report.

Before production integration, collect a time-separated sample of real packages from the 20 priority sources; have a reviewer label title, description, TOR, eligibility, and primary/secondary categories without seeing model output. Report precision, recall, confusion matrix, false positives, false negatives, source coverage, and confidence intervals on that held-out sample. Resolve owner corrections, then rerun this evaluator and the Go tests on an immutable taxonomy version.

## Scoring contract

The v2 candidate score uses 35/25/20/10/10 weights for capability, scope, supplier qualification, value/margin, and delivery readiness. A numeric dimension requires evidence; otherwise it remains unknown. Unknown dimensions are omitted from the weighted numerator and denominator; data completeness is reported separately and is not a calibrated confidence probability. An incomplete score is `provisional=true`. A verified unmet mandatory requirement sets `eligibility_status=ineligible` without rewriting the score. Scoring is not wired into the production collector.

| Check | Score | Band | Data completeness | Provisional | Eligibility |
|---|---:|---|---:|---|---|
| No verified dimensions | unknown | — | 0% | true | unknown |
| Capability only (90) | 90 | high | 35% | true | unknown |
| Five verified dimensions | 77 | medium | 100% | false | unknown |
| Verified mandatory failure | 90 | high | 35% | true | ineligible |
