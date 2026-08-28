# Runtime architecture

## Ownership boundary

`eventframed` owns storage, embeddings, retrieval, scoring, packing, versioned
state, and future Bayesian/graph policy. The OpenClaw plugin owns hook translation,
prompt injection, and authority decisions. LibraVDB is embedded in the daemon.

This split keeps the agent integration replaceable without moving correctness
rules into JavaScript. It also avoids a third persistent process and a second RPC
boundary between policy and vector storage.

## Current data path

1. The adapter observes a successful user/assistant turn.
2. It derives a deterministic event ID and sends an idempotent observation.
3. The daemon validates field provenance and `occurred <= observed <= available`.
4. LibraVDB stores the event and its canonical vector.
5. Recall requires an explicit `as_of` timestamp.
6. Storage excludes unavailable events before satisfying the candidate budget.
7. The service reranks up to `recall_k` candidates.
8. Packing independently admits at most `pack_k` within the token budget.
9. The adapter escapes records and labels them untrusted history.

The storage adapter expands LibraVDB probes when post-ANN availability filtering
does not yield enough valid candidates. This prevents future-dated records from
crowding valid history out before reranking. The expansion is bounded by live
collection cardinality and must be included in tail-latency benchmarks.

## Trust boundary

Stored content can contain arbitrary user, model, or tool text. It is data, not
authority. Prompt envelopes are escaped and include a direct instruction not to
execute recalled content. A production release must also apply tenant identity at
the socket boundary and enforce deletion/retention policy through every index.

## Selective Bayesian staging

The current Phase 3A frontier is shadow-only. It computes the declared weighted
activation score over bounded vector candidates, leaves typed slots for
sheaf-inspired and as-of graph compatibility, lowers the threshold for critical
priority, enforces an active-count cap, and chooses inactive audits through a
score-independent deterministic Bernoulli draw. Declared hypotheses without an
available evidence frame may be nominated but cannot activate.

Because vector-nomination support is not yet estimated, every report sets
`selection_support_certified=false`; the posterior key remains event-specific and
the scored retrieval law is unchanged. This is intentional. Anti-Pigeon sharing
and production posterior updates remain disabled until their external
certificates and durable propensity journal are implemented.
