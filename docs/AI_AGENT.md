# AI Agent Design

## 1. Role

The AI layer enriches decision-making; it is not the only decision maker. Deterministic rules handle lifecycle exclusions, configured value bounds, known keywords, and validated mappings before any model call.

## 2. Capabilities

| Capability | Input | Structured output |
|---|---|---|
| IT classifier | Title, category, description, KBLI/KBKI | is-IT, capability group, confidence, evidence |
| Requirement extractor | Announcement, schedule, qualification text | requirement type, normalized value, evidence, confidence |
| Opportunity assessor | Package data and CANKONIX capability profile | factor signals, risks, verification questions |
| Executive summarizer | Canonical package and recent changes | concise summary, risks, next checks |
| Change analyst | Before/after fields | materiality, eligibility impact, timeline impact |

## 3. Taxonomy

- Application & Software
- Integration & Data
- Cloud & Infrastructure
- Network & Connectivity
- Cybersecurity & Identity
- End-user Computing
- Managed Services
- Consulting & Expert Services

Capability fit also maps to CANKONIX pillars: Cloud, Automation, Network, Knowledge, Operations, Next-Generation, Integration, and Xperience.

## 4. Structured result envelope

Every AI result must include:

```json
{
  "schema_version": "1.0",
  "task": "classify_it",
  "result": {},
  "confidence": 0.0,
  "evidence": [
    {"field": "title", "snippet": "..."}
  ],
  "model": "provider/model",
  "prompt_version": "classify-it/v1",
  "input_fingerprint": "sha256:...",
  "created_at": "RFC3339 timestamp"
}
```

Reject responses that fail schema validation. Retry once with a repair instruction, then route to deterministic fallback or human review.

## 5. Opportunity score

Initial weights:

| Factor | Weight | Basis |
|---|---:|---|
| Capability fit | 25% | Scope versus CANKONIX capabilities |
| Eligibility | 20% | KBLI, experience, certifications, staffing |
| Commercial fit | 15% | Pagu/HPS, guarantees, working capital, indicative margin |
| Time readiness | 15% | Remaining time versus bid-preparation effort |
| Strategic value | 15% | Target institution, repeatability, reference value |
| Delivery risk | 10% | Location, complexity, SLA, principal and integration dependencies |

The scoring service, not the model, calculates the final numeric score from validated factor inputs. It stores the policy version and factor breakdown.

## 6. Guardrails

- Auto-reject only through approved deterministic rules.
- Low confidence cannot permanently hide an item.
- Evidence must point to provided source fields; the model must not invent qualifications.
- Never send secrets, credentials, irrelevant PII, or unbounded document content to a provider.
- Users can inspect, correct, and override AI results.
- Provider failure must leave the package available with a pending or rule-only state.

## 7. Evaluation

Maintain versioned golden datasets for:

- IT/non-IT classification
- capability taxonomy
- requirement extraction
- material change detection
- summary factuality

Release gates:

- classification precision at least 80% for MVP
- schema-valid output at least 99%
- evidence coverage at least 95% for positive classifications
- no critical deterministic exclusion regression

Report precision, recall, confusion matrix, extraction field accuracy, latency, token usage, and estimated cost by prompt/model version.

## 8. Cost and latency controls

- Cache results by task, prompt version, and input fingerprint.
- Run cheap deterministic extraction before model calls.
- Use small models for classification/extraction and escalate only ambiguous cases.
- Batch where provider and latency requirements allow.
- Enforce daily and per-source budgets with graceful degradation.

## 9. Provider adapter

Expose one internal interface for structured generation. Provider-specific SDKs remain behind adapters. Business tables store provider-neutral task results and retain model metadata for audit and regression.

