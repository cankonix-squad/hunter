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
