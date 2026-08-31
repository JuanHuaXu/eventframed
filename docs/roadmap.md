# Implementation roadmap

## Milestone 1: executable memory slice

- Durable LibraVDB event storage
- Availability-safe recall and independent packing
- OpenClaw recall/capture adapter
- Idempotency, provenance, and Unix-socket isolation
- Replay, race, fault, and latency tests

## Milestone 2: production embeddings and lifecycle

- [x] Pluggable embedding provider and embedding-model version keys
- [x] Canonical-vector persistence with SQ8/FSQ6/PQ traversal contracts
- [x] Durable runtime, policy, graph, posterior, residual, abstraction, and contract versions
- [x] Deletion, retention, compaction, backup, and recovery propagation
- [x] Guarded Phase 1 migration with mandatory pre-migration backup
- [x] Restart, recovery, model-mismatch, and protocol compatibility tests
- [x] Daemon-owned LibraVDB contract indexing, nomination, ranking, deletion, and retention
- [x] Tenant-derived external collections and retry reconciliation without modifying `libravdb-memory`

## Milestone 3: selective Bayesian layer

- [x] Capped vector frontier with typed graph/sheaf score slots
- [x] Shadow activation decisions and independent deterministic audit sampling
- [x] Explicit uncertified-selection fallback that cannot alter the scored law
- [x] Durable nomination and activation propensity journal
- [x] Externally certified Anti-Pigeon posterior-sharing guard
- [x] Cached bounded Beta posterior updates and capped changepoint invalidation
- [x] Warm-started CUSUM drift signal, reset cooldown, and seeded confirmation benchmark
- [x] Proposal-only Bayesian shared-versus-split comparison with member evidence
- [x] Anti-Pigeon authority guard: comparisons cannot mutate active grouping
- [x] Independent shadow-audit feedback and omitted-influence certificates

## Milestone 4: residuals and abstraction

- [x] Retrieval-usefulness belief-conditioned law and aligned template bundle
- [x] Law-only residual cache with analytic posterior-motion and anytime-valid improvement certificates
- [x] Graph dependency closure and bounded sheaf-inspired predictive snapping
- [x] Immutable outcome commitments and delayed Brier calibration
- [x] External split/merge confirmation contract and monotone rollback

## Milestone 5: bounded agency

- [x] Signed data-only wake/notify/schedule proposals
- [x] Capability and consent checks in the OpenClaw authority layer
- [x] Expiry, idempotency, causal-chain budgets, quiet hours, and kill switch
- [x] Authenticated durable claim leases, bounded active projection, restart recovery, and evidence-deletion cancellation
- [x] No direct tool execution by `eventframed`

## Milestone 6: whitepaper claims validation

- [x] Leakage-rejecting chronological evaluation contract
- [x] Retrieval Brier, calibration, ranking, nomination, and activation metrics
- [x] Pre-outcome priority weighting and trajectory-cluster bootstrap intervals
- [ ] Automated policy replay from actual `eventframed` forecast journals
- [ ] Update-all and naive-selective counterfactual baselines
- [ ] Synthetic drift, hidden-subgroup, and snapping claim matrix (initial
  stable/hidden-shift Bayesian ablation complete; subgroup and snapping remain)
- [ ] Untouched OpenClaw confirmation trajectories
- [ ] Full marked next-event/no-event proper-score evaluator

## Milestone 7: background invariant incubation

- [x] Low-certainty recall nomination using the existing packing-boundary signal
- [x] Nonblocking bounded queue with set-based cooldown deduplication
- [x] Single idle-gated worker with timeout and stale-snapshot rejection
- [x] Vector-only query payload and aggregate queue-status API
- [ ] Durable pending-job recovery and independent randomized coverage stream
- [ ] Mixed-load p99 and trigger-yield confirmation benchmark

The production policy defaults to updating the complete bounded retrieval
frontier. Selective activation remains an ablation until fresh confirmation
shows that it retains enough frontier-update-all quality to justify suppression.

## Required evidence before production

Benchmark p50/p95/p99 latency, resident memory, index build cost, activation rate,
availability-filter expansion, omitted-influence coverage, false Anti-Pigeon
merges, calibration, and task accuracy against update-all and naive-selective
baselines. Fault tests must include duplicate delivery, crash/restart, stale
versions, deletion races, corrupt records, and unavailable daemon behavior.
Production evidence must also measure stale ANN candidate pressure after
contract `Delete`/`DeleteBatch`; authoritative EventFrame resolution blocks
deleted output, but the current external index may continue nominating IDs that
are already absent from authoritative metadata lookup.
