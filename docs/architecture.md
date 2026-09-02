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

The remote boundary has one shared concurrency limiter, per-attempt deadlines,
bounded retry backoff, and a consecutive-failure circuit breaker. Remote index
mutations use a single-writer lane while bounded reads may continue. The local EventFrame
commit happens first and remains the durable reconciliation source when a remote
index response is lost. `/v1/health` reports daemon liveness, while `/v1/ready`
also requires the configured external contract connection.

The embedded LibraVDB store, pass-through ranker, and bounded packer remain for
standalone operation and deterministic tests. They are degraded-mode fallbacks
and cannot be used as evidence for contract-native OpenClaw retrieval claims.

## Current data path

1. The adapter observes a successful user/assistant turn.
2. It derives a deterministic capture ID and sends an idempotent raw-turn
   envelope through `POST /v1/turns:capture`. The adapter contract contains text
   and transport metadata, but no Who, What, Where, When, Why, or How fields.
3. After decoding that contract, `eventframed` performs bounded deterministic
   5W1H enrichment. Explicit user spans are `observed`, assistant spans are
   `synthetic`, and compositions or metadata fallbacks are `inferred`. No model
   or network call is made on this enrichment path.
4. The daemon validates field provenance and `occurred <= observed <= available`.
   The unmodified turn remains opaque payload metadata; enrichment does not establish
   causality. Direct callers of `/v1/events:observe` may still submit deliberately
   authored structured frames, which the daemon validates without rewriting.
5. EventFrame commits canonical `eventframe-5w1h-v1` text as the semantic corpus
   and retains the raw transcript in metadata for final delivery. Local vectors
   and remote `InsertText` use only that canonical text and reconcile retries by
   a representation-bound digest.
6. Recall supplies an explicit `as_of`; EventFrame derives the tenant's reserved
   collection and applies bounded exclusions internally.
7. `SearchTextCollections` nominates candidates. EventFrame resolves every
   returned ID against its tenant and availability view; unknown, cross-tenant,
   or future records fail closed.
8. The service builds and durably journals a capped Bayesian frontier. The
   default policy updates every evidence-ready member of that frontier, never
   the full corpus.
9. Only current external certificates may promote a cached usefulness posterior
   into the bounded score mixture; otherwise the baseline score is unchanged.
10. Exact-key then general residual lookup may correct the complete Bernoulli law
   only when fixed-reference motion and sequential improvement gates pass.
11. The service derives an EventFrame correction relative to its frozen local
    baseline and stores that bounded, version-bound rank delta in a write-through
    RAM cache backed by SQLite.
12. EventFrame sends canonical 5W1H candidate text, a canonical partial query
    frame, LibraVDB's nomination/base signal, and preserved typed metadata to
    `RankCandidates(k1,k2)`. Raw transcripts do not cross this boundary. The
    returned value is preserved as `retrieval_score`.
13. EventFrame measures answer certainty from the normalized retrieval-score gap
    at the initial top-`pack_k` boundary. Independently, only an accepted
    Bayesian, residual, or graph-compatibility path supplies correction
    reliability. Both promotions and demotions use those two signals under the
    global hard cap.
14. EventFrame adds the resulting `rank_delta`, sorts by the final `score`, and
    then applies claim/lineage occupancy, `pack_k`, and the token budget.
    Repeated or near-duplicate claims from one provenance lineage share a hard
    occupancy budget. Different externally certified Anti-Pigeon buckets remain
    separate predictive units, and an ordinary chat timestamp cannot create an
    independent evidence lineage.
15. Only after final ranking and packing does the daemon hydrate raw transcript
    metadata. The adapter recalls through `/v1/openclaw/context:recall`, whose daemon-side
    projection omits 5W1H fields and embeddings. It escapes the remaining raw
    records and labels them untrusted history.

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

## Higher-order invariant composition

Join discovery remains an asynchronous invariant-seeker task. A proposed join is
not trusted structure: the daemon resolves every constituent at the declared
publication time, requires one current Anti-Pigeon certificate for the exact
member set, and compares the frozen snapshot again under the storage write lock.
Only then does it materialize a normal EventFrame with kind `higher_order`.

The macro frame carries the exact member IDs, one directly retrievable
representative, rule and resolution identifiers, confidence, certificate ID,
and evidence epoch. Its 5W1H projection keeps only fields invariant across all
members, plus a declared higher-order `what` label and the constituent time
envelope. Source frames remain authoritative and are neither rewritten nor
deleted. LibraVDB persists the macro in the same tenant collection, so restart,
backup, and normal nomination use the existing lifecycle.

Recall checks that the certificate still owns the exact member set. Revocation
or replacement therefore excludes a stale macro without waiting for maintenance.
Explicit decomposition physically removes only the macro and advances the
abstraction version while atomically retaining a reasoned tombstone. `coarse`
and `fine` recall modes add bounded rank-only
preferences after the frozen candidate contract and before packing; `auto`
changes nothing. This gives broad queries a macro view and detailed queries the
constituents without pretending the abstraction replaced its evidence.

## Validity-constrained property fuzzing

Property fuzzing is an explicit read-only slow path. A caller binds an exact
runtime snapshot, an as-of context of 2 to 64 EventFrames, a query, a stability
threshold, a one-sided confidence contract, and at most 512 declared
perturbations. Each perturbation names its target event, property family,
validity rule, validation kind, and one atomic synthetic replacement bundle over
5W1H fields. A declared context relocation may change only `who`, `where`, or
`when`. A semantic perturbation must copy `what`, `why`, and `how` atomically
from a distinct as-of source EventFrame in the same context; the daemon verifies
each value against that source. Missing validity provenance, unknown fields,
no-op replacements, future evidence,
cross-tenant evidence, duplicate identifiers, and stale snapshots fail closed.

The current runtime output functional is a normalized nomination law over the
bounded context. It embeds the query and each canonical EventFrame, transforms
cosine similarity into nonnegative mass, normalizes across the fixed context,
and measures total-variation movement after each valid perturbation. A property
is reported as a conditional model invariant only when the minimum trial count
is met and the Bonferroni-simultaneous one-sided Wilson lower bound on the
stable-trial probability reaches the declared requirement. Point estimates or
unadjusted per-property intervals alone cannot certify it. Every property chosen
under one audit design must be evaluated in the same request; request splitting
does not preserve family-wise coverage.

The operation does not alter stored events, vectors, posteriors, residuals,
rank deltas, graph state, or composition state. It does not use raw transcript
content. Unchanged canonical frames are embedded once per request, so its cost
is `O((N+P)D + PN)` for `P` perturbations, `N` bounded context events, and
embedding dimension `D`, so it is never called by
the recall hot path. Results are sensitivity fingerprints for ontology review
and small-scale transfer hypotheses. They are not interventions or causal
effects, and they cannot authorize a join without separate Anti-Pigeon evidence.

Low packing-boundary certainty may nominate this audit after a successful
recall. Nomination copies only the existing query vector, query digest, exact
snapshot, bounded candidate IDs, and source-EventFrame semantic bundles into a
nonblocking bounded queue. The request path never runs perturbation predictions.
A single worker checks at a configured interval and starts at most one job only
when no recall is active. Queue saturation drops the nomination rather than
blocking recall; cooldown deduplication is set-based over the bounded candidate
IDs so tied retrieval order cannot duplicate work.

The worker fails closed when the snapshot has moved, uses the captured query
vector without retaining raw query text, and records only an aggregate result
summary. Its jobs are adaptively selected model-sensitivity audits, not an
unbiased sample of all EventFrames. Population claims therefore require an
independent randomized audit stream or valid inclusion-probability correction.
The current queue is in-process and non-durable, so restart discards pending
jobs. That limitation affects audit coverage, not serving correctness.

## Predictive chain translation

Chain translation is a separate snapshot-bound, read-only slow path. It
resolves four observed trajectories: baseline and revealed chains in each of
two domains. Frozen stage maps bind the compared coordinate and exact
before/after correspondence. Every trajectory is occurrence-ordered, every
non-target 5W1H coordinate must remain unchanged, and a strict translation
candidate must commute at every aligned stage and preserve the signed
nomination-law effect within tolerance.

An erased terminal distinction with small law movement in both domains is an
invariant candidate. Mapped propagation through every stage with bounded
cross-domain effect defect is a predictive-translation candidate. Any local
mismatch, incomplete propagation, terminal disagreement, or excessive defect
is divergence. These are diagnostic labels over declared maps; graph
publication, Anti-Pigeon grouping, autonomous map discovery, and SCM-backed
causal acceptance remain separate transitions.

The evaluator orders work from cheapest to most expensive. It first validates
chronology, exact stage correspondence, and unchanged-coordinate locality. A
candidate that cannot satisfy either predictive branch returns divergence with
`prediction_evaluated=false` and never constructs the predictor. Otherwise the
service hydrates vectors through the optional audit-only store capability;
ordinary `GetEvents` remains vector-free so recall packets cannot grow silently.
Only vectors whose model key and semantic-representation marker match the active
embedder are reused. A bounded concurrent cache reuses canonical query vectors
and immutable-event raw nomination scores, with fixed-size digests for query and
generated-frame keys plus single-flight miss collapse.
Changed or mismatched EventFrames are embedded normally. Cache eviction changes
cost only and cannot change the normalized law or classification.

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

OpenClaw does not infer usefulness from successful generation or packing. An
operator-write control-plane method submits an explicit signal tied to the exact
durable recall journal and nominated event. The daemon's journal membership,
availability, and idempotency checks remain the authority at the sink.
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

An edge effect is either symmetric `compatible`, directional `supports`, or
directional `supersedes`. Supersession maps strong source support to a low target
graph feature, creating a bounded negative target rank correction without
changing the scored Bernoulli law or feeding target state back into the source.
Legacy edges with no effect remain symmetric. Split publication invalidates the
complete reverse dependency closure; it does not suppress the retrieved
frontier.

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
