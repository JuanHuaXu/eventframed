# Runtime architecture

## Ownership boundary

`eventframed` owns storage, embeddings, retrieval, scoring, packing, versioned
state, and Bayesian/graph policy. The OpenClaw plugin owns hook translation,
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
7. The service builds and durably journals a capped Bayesian activation frontier.
8. Only current external certificates may promote a cached usefulness posterior
   into the bounded score mixture; otherwise the baseline score is unchanged.
9. Exact-key then general residual lookup may correct the complete Bernoulli law
   only when fixed-reference motion and sequential improvement gates pass.
10. The service derives the aligned template and reranks up to `recall_k`
    candidates after any certified correction.
11. Packing independently admits at most `pack_k` within the token budget.
12. The adapter escapes records and labels them untrusted history.

The storage adapter expands LibraVDB probes when post-ANN availability filtering
does not yield enough valid candidates. This prevents future-dated records from
crowding valid history out before reranking. The expansion is bounded by live
collection cardinality and must be included in tail-latency benchmarks.

## Trust boundary

Stored content can contain arbitrary user, model, or tool text. It is data, not
authority. Prompt envelopes are escaped and include a direct instruction not to
execute recalled content. A production release must also apply tenant identity at
the socket boundary and enforce deletion/retention policy through every index.

## Selective Bayesian layer

The Phase 3 frontier computes the declared weighted activation score over bounded
vector candidates, leaves typed slots for sheaf-inspired and as-of graph
compatibility, lowers the threshold for critical priority, enforces an
active-count cap, and chooses audits through a score-independent deterministic
Bernoulli draw. Declared hypotheses without an available evidence frame may be
nominated but cannot activate.

Each decision is durable before recall returns. The journal records the policy
snapshot and inclusion probabilities needed by later feedback. External
selection-support and omitted-influence certificates are both required before a
posterior can affect ranking. Anti-Pigeon sharing is a third, independent
certificate; without it, each event keeps its own posterior. Feedback is
availability checked, idempotent, inverse-propensity weighted with a hard cap,
and monitored by a capped online changepoint detector. Any absent, expired, or
epoch-mismatched certificate yields conservative baseline fallback.

## Residual specialization

Phase 4A specializes the paper's general law/template bundle to retrieval
usefulness. `useful` and `not_useful` are explicit Bernoulli branches. The
baseline score is a declared plug-in probability, the certified Beta posterior
forms the optional belief mixture, and the residual cache contains law-only
corrections. A correction moves mass between both branches and the decision
template is derived afterward, preventing point/law semantic drift.

Exact action keys bind the query digest, event ID, and horizon. General keys bind
the Anti-Pigeon posterior bucket or event-local posterior key and horizon. Cache
records retain their first base-law reference. Runtime motion is exact absolute
Bernoulli distance with zero approximation error for this implementation; the
separate improvement gate uses error spending across repeated checks and counts
only unweighted full-stream or independently audited validation trials. Delayed
outcomes score the immutable corrected law and atomically update posterior,
calibration, exact residual, general residual, and version state.

## Predictive compatibility graph

Phase 4B persists a tenant-scoped, bounded predictive compatibility graph. Nodes
may represent buckets, predictors, resolutions, or agents, but every current node
shares the concrete `retrieval-usefulness-v1` Bernoulli law space and every edge
uses the declared `identity_bernoulli` comparison map. This is deliberately a
sheaf-inspired compatibility scaffold. It neither instantiates general sheaf
laws nor assigns causal meaning to an edge.

An external slow path selects a complete candidate graph from a finite family on
a chronological design window, then evaluates it on a later untouched
confirmation window. The daemon does not trust a caller-supplied affected set.
It diffs the current and candidate graphs, expands changed nodes through the
union of old and new adjacency up to the configured radius, and derives affected
edges, event IDs, and posterior keys itself. Fixed comparison obligations are
evaluated against the candidate so deleting a difficult edge cannot silently
erase its burden.

Publication requires passing external future-diameter certificates for every
affected active bucket, compatibility certificates for every affected retained
or new edge, simultaneous coverage, proper-risk non-inferiority, bounded
unresolved burden, and positive priority gain net of resource cost. LibraVDB
publishes the graph, snap audit record, targeted posterior/residual stale marks,
and monotone versions in one transaction. Rollback republishes the previous
topology under a new version and repeats stale marking; it never reactivates old
posterior or residual state.
