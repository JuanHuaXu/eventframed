# Runtime architecture

## Ownership boundary

In the OpenClaw production path, `eventframed` owns the complete memory contract:
observation lifecycle, tenant-scoped candidate indexing, contract-native
`SearchTextCollections` nomination, `RankCandidates` two-pass ranking,
availability checks, Bayesian and Anti-Pigeon state, residual forecasts, graph
policy, packing, and durable decision journals. The TypeScript OpenClaw adapter
only translates hooks and injects the returned context packet. The original
`libravdb-memory` plugin is a frozen contract reference and is not modified or
loaded alongside EventFrame.

The daemon calls the frozen LibraVDB RPC contracts through an optional external
endpoint. It assigns each tenant a deterministic hashed `eventframe-*`
collection; callers cannot override that scope. The embedded LibraVDB database
remains the authoritative structured EventFrame and policy store.

The embedded LibraVDB store, pass-through ranker, and bounded packer remain for
standalone operation and deterministic tests. They are degraded-mode fallbacks
and cannot be used as evidence for contract-native OpenClaw retrieval claims.

## Current data path

1. The adapter observes a successful user/assistant turn.
2. It derives a deterministic event ID and sends an idempotent observation.
3. The daemon validates field provenance and `occurred <= observed <= available`.
4. EventFrame commits the structured event, then idempotently indexes its text
   and typed metadata through `InsertText`, reconciling retries by digest.
5. Recall supplies an explicit `as_of`; EventFrame derives the tenant's reserved
   collection and applies bounded exclusions internally.
6. `SearchTextCollections` nominates candidates. EventFrame resolves every
   returned ID against its tenant and availability view; unknown, cross-tenant,
   or future records fail closed.
7. The service builds and durably journals a capped Bayesian frontier. The
   default policy updates every evidence-ready member of that frontier, never
   the full corpus.
8. Only current external certificates may promote a cached usefulness posterior
   into the bounded score mixture; otherwise the baseline score is unchanged.
9. Exact-key then general residual lookup may correct the complete Bernoulli law
   only when fixed-reference motion and sequential improvement gates pass.
10. The service derives an EventFrame correction relative to its frozen local
    baseline and stores that bounded, version-bound rank delta in a write-through
    RAM cache backed by SQLite.
11. EventFrame sends LibraVDB's nomination/base signal and preserved metadata to
    `RankCandidates(k1,k2)`. It preserves the returned value as `retrieval_score`.
12. EventFrame measures answer certainty from the normalized retrieval-score gap
    at the initial top-`pack_k` boundary. Independently, only an accepted
    Bayesian, residual, or graph-compatibility path supplies correction
    reliability. Both promotions and demotions use those two signals under the
    global hard cap.
13. EventFrame adds the resulting `rank_delta`, sorts by the final `score`, and
    only then applies `pack_k` and the token budget.
14. The adapter escapes records and labels them untrusted history.

The SQLite sidecar is a durable materialization and restart cache, not a new
authorization source. Every row is keyed by query and event, expires quickly,
and must match policy, evidence epoch, graph, posterior, residual, and
abstraction versions. A stale row cannot reactivate a posterior or residual that
failed the current recall's gates.

For answer certainty `c` and correction reliability `r`, the default elastic
scale is `[0.5 + (2.5 - 0.5)(1 - c)]r`. Thus an uncertain packing boundary can
move quickly, a certain boundary moves conservatively, and an uncertified
correction (`r = 0`) cannot move at all. The boundary is the contract-returned
order at `pack_k`; its certainty is the nonnegative inside/outside score gap
normalized by the larger score magnitude. This is rank-domain certainty, not a
calibrated probability that the answer is correct. Responses expose
`packet_answer_certainty`, `rank_delta_answer_certainty`,
`rank_delta_correction_reliability`, the scale, and the basis. The older
`packet_confidence` and `rank_delta_confidence` fields are deprecated aliases for
answer certainty. Elasticity never creates a raw delta, cannot bypass
`MaxRankDelta`, and version `eventframe-rank-delta-v2` prevents reuse of cached
rows written under the former confidence semantics.

Deletion and retention mirror IDs through `Delete` and `DeleteBatch`. If an
external deletion was interrupted after the structured record was removed, or
the external ANN index temporarily returns a deleted vector, nomination
resolution detects the stale ID, verifies it is not merely future-dated,
reissues deletion, and omits it safely. Current LibraVDB builds may retain such
ANN hits after metadata lookup no longer finds the record; many stale hits can therefore
consume nomination slots until external index maintenance completes. This is a
recall-quality risk, not an authorization path, and must be measured in
deletion-heavy workloads.

The standalone storage adapter expands embedded LibraVDB probes when post-ANN availability filtering
does not yield enough valid candidates. This prevents future-dated records from
crowding valid history out before reranking. The expansion is bounded by live
collection cardinality and must be included in tail-latency benchmarks.

## Trust boundary

Stored content can contain arbitrary user, model, or tool text. It is data, not
authority. Prompt envelopes are escaped and include a direct instruction not to
execute recalled content. A production release must also apply tenant identity at
the socket boundary. EventFrame enforces tenant isolation by deriving the only
permitted external collection from the authenticated tenant identifier.

## Selective Bayesian layer

The Phase 3 frontier computes the declared weighted activation score over bounded
vector candidates and leaves typed slots for sheaf-inspired and as-of graph
compatibility. The production default sets both activation thresholds to zero
and the active-count cap to `recall_k`, updating every evidence-ready frontier
member. Explicit selective policies may raise thresholds or lower the cap for
ablations and constrained deployments. Declared hypotheses without an available
evidence frame may be nominated but cannot activate. Audits use a
score-independent deterministic Bernoulli draw.

Each decision is durable before recall returns. The journal records the policy
snapshot and inclusion probabilities needed by later feedback. External
selection-support and omitted-influence certificates are both required before a
posterior can affect ranking. Anti-Pigeon sharing is a third, independent
certificate; without it, each event keeps its own posterior. Feedback is
availability checked, idempotent, inverse-propensity weighted with a hard cap,
and preserves per-event evidence even inside a certified shared posterior.
Certified sharing discounts only the pooled posterior's effective outcome weight
(default `0.5`). Direct member sufficient statistics retain full audit weight so
the group comparison and any later split do not lose divergence evidence.

The slow-path group comparison evaluates a shared Beta-Bernoulli model against
independent member models using marginal likelihood. It emits only `share`,
`split`, or `uncertain`; it cannot publish a certificate, mutate a posterior key,
or change an abstraction. Anti-Pigeon external certification remains the sole
authority for posterior sharing.

Outcomes are monitored by capped Bayesian online changepoint state plus a
constant-time, warm-started two-sided CUSUM. The CUSUM absorbs small predictive
deviations, accumulates sustained directional motion, and has a bounded reset
cooldown. Any accepted change resets the affected posterior and advances the
evidence, graph, posterior, residual, and abstraction versions. Any absent,
expired, or epoch-mismatched certificate yields conservative baseline fallback.

## Residual specialization

Phase 4A specializes the paper's general law/template bundle to retrieval
usefulness. `useful` and `not_useful` are explicit Bernoulli branches. The raw
baseline retrieval score and certified Beta posterior form the optional bounded
belief-conditioned score. A frozen predictive map calibrates that complete score
into the pre-residual Bernoulli law, while a separately fitted baseline map is
used when no certified belief is accepted and for the reported base law. Both
maps are policy-versioned. The residual cache contains law-only corrections. A
correction moves mass between both branches and the decision template is derived
afterward, preventing point/law semantic drift.

Calibration fitting uses only nominated, sink-visible forecasts. Evaluator-only
zero placeholders for events outside the retrieval frontier remain part of the
end-to-end score, but cannot train the emitted-probability map. The fitter uses
accepted Newton steps and falls back to the existing map when the declared
design Brier score would worsen; calibration remains monotone and cannot reorder
candidates.

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

## Bounded agency authority

Phase 5 separates proposal generation from authority. An authenticated local
planner may submit a data-only draft to the daemon. `eventframed` validates the
action, evidence availability, utility, priority, timing, causal ancestry, chain
budget, and tenant queue bound before signing the exact JSON payload with
Ed25519. Proposal state, leases, terminal resolutions, and `agency_version` are
durable. Deleting or retaining cited evidence atomically rejects any pending or
claimed dependent proposal.

The daemon's issuer token, authority token, and private key stay mode 0600. A
trusted planner receives the issuer token; the OpenClaw adapter receives the
separate authority token and public key. Claim and resolution fail before storage
access when authority authentication fails. The adapter verifies the exact signed payload and then
applies a second, independent authority policy: configured tenant, lease holder,
session prefix, capability, explicit action consent, causal-depth cap, expiry,
quiet hours, and kill switch. Failure at any gate rejects the proposal. The
allowed sink is OpenClaw's session-turn scheduler, never a tool API.

The authority sequence is:

1. Trusted planner issues a wake, notify, or schedule draft with existing evidence.
2. The daemon validates, signs, persists, and later leases the proposal.
3. The adapter verifies the signature and local consent policy.
4. The adapter schedules a normal agent turn under a deterministic proposal tag.
5. The adapter records approval and scheduler handle in the daemon.

The scheduler and LibraVDB do not share a transaction. If durable approval fails,
the adapter first removes the scheduled tag and then records rejection. If that
rollback is uncertain, it leaves the lease unresolved so a retry begins by
removing the same deterministic tag. A crash after execution but before durable
resolution can still produce at-least-once behavior; this limitation must be
measured in restart and duplicate-delivery fault tests.

Canonical proposal records remain in the tenant-independent agency archive for
audit and idempotency. Contract 7 also maintains a tenant-scoped active
projection containing only pending and claimed records. Authority polling scans
that projection, bounded by the 1,000-record tenant queue limit, rather than
lifetime history. Issue, claim, resolution, evidence deletion, active projection,
and version updates share LibraVDB transactions. Contract-6 databases rebuild
the projection before publishing the contract-7 marker.
