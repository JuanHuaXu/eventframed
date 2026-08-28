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

## Milestone 3: selective Bayesian layer

- [x] Capped vector frontier with typed graph/sheaf score slots
- [x] Shadow activation decisions and independent deterministic audit sampling
- [x] Explicit uncertified-selection fallback that cannot alter the scored law
- [ ] Durable nomination and activation propensity journal
- Anti-Pigeon posterior-sharing guard
- Cached incremental posterior updates and changepoint invalidation
- Shadow update-all audits and omitted-influence certificates

## Milestone 4: residuals and abstraction

- Belief-conditioned forecast law and template bundle
- Residual cache with posterior-motion certificates
- Graph dependency closure and sheaf-inspired snapping
- Outcome commitments, delayed scoring, calibration, and split/merge audits

## Milestone 5: bounded agency

- Signed data-only wake/notify/schedule proposals
- Capability and consent checks in the OpenClaw authority layer
- Expiry, idempotency, causal-chain budgets, quiet hours, and kill switch
- No direct tool execution by `eventframed`

## Required evidence before production

Benchmark p50/p95/p99 latency, resident memory, index build cost, activation rate,
availability-filter expansion, omitted-influence coverage, false Anti-Pigeon
merges, calibration, and task accuracy against update-all and naive-selective
baselines. Fault tests must include duplicate delivery, crash/restart, stale
versions, deletion races, corrupt records, and unavailable daemon behavior.
