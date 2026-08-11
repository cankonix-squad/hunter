# Notification delivery model

Alert events are immutable matches between a canonical opportunity version and a
user's alert configuration. Delivery adapters consume those events later; they do
not recreate matching decisions.

## Deduplication invariant

At most one event may exist for `(user_id, opportunity_version_id, event_type)`.
The collector first selects one matching rule per user. When several rules match,
the rule with the most source/relevance/minimum-score filters wins, followed by the
oldest rule as a deterministic tie-breaker. A database unique constraint remains
the final guard against concurrent collectors and retries.

This means users do not receive duplicate notifications merely because they have
overlapping rules. Different canonical versions remain separate events, and a new
event type for the same version is also separate. Channel-level delivery retries
must reuse the alert event identity rather than create another event.

Quiet hours and email/webhook delivery are subsequent layers. They may change
delivery state, but must not weaken or bypass this invariant.

## Quiet hours

Each user may define an enabled quiet-hours range and IANA timezone. Both same-day
and overnight ranges are supported. Matching remains immediate and immutable, but
events created during the range receive an `available_after` timestamp at the next
local end time. Delivery workers must select only pending events whose
`available_after <= now()`; this preserves alerts without sending during quiet hours.

## Email delivery

The notifier process materializes one email outbox row per alert event and claims due
rows with a database lease. Successful sends become `delivered`; failures use bounded
exponential retry and become `dead_letter` after five attempts. Safe errors are capped
and credentials are never stored in delivery records.

SMTP delivery requires STARTTLS with TLS 1.2 or newer. Configure it only through
environment variables. `EMAIL_DRY_RUN=true` uses the same outbox/claim/status path but
does not contact an external server; it is intended only for local acceptance. Both
`EMAIL_DELIVERY_ENABLED` and dry-run default to false.

## Webhook delivery

The same notifier/outbox pipeline can deliver a structured JSON payload to one
deployment-level HTTPS endpoint. Requests include `Idempotency-Key: alert-event-{id}`
and `X-Cankonix-Signature: sha256={HMAC}` over the exact body. The shared secret must
contain at least 32 characters, redirects are not followed, response bodies are bounded,
and only HTTP 2xx is successful. Response status codes are retained for operations while
body content is discarded. Webhook delivery also supports explicit local dry-run mode.
