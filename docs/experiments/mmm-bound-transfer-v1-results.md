# Bound-label rebuild v1: component PASS, integration OPEN

2026-10-01. The [frozen contract](mmm-bound-transfer-v1-contract.md) was tested
in the research-only `internal/researchmemory` package. This is a bounded
component, **not** an automatic cross-epoch learner transfer or live bridge.

`RebuildFromBoundLabels` accepts at most 256 ordered pairs of original
forecast records and terminal feedback, checks record shape, binding presence,
tenant, paired IDs, unique source journal/event and prediction IDs, epoch,
feature range, availability order, and a caller-supplied cutoff. A required
external validator decides whether each pair survives under the target
snapshot. Any malformed input or validation error returns no model. Rejected
pairs contribute no training sample. Only approved samples are refit into
short/long count models and the tree forest; pending predictions and both
forecast-mixture weight vectors reset.

## Finite checks

- 64 alternating source-A/B pairs, with A rejected and B retained, produced
  exactly 32 samples. Across all 512 feature values, every rebuilt component
  forecast matched fresh direct fits on B only.
- Flipping rejected A outcomes changed no rebuilt forecast. Flipping retained
  B outcomes changed a forecast. All-rejected input stayed cold at the supplied
  baseline. The old epoch could not score with the rebuilt model.
- Wrong tenant, duplicate identity, mismatched original/feedback IDs, unbound
  record, bad contract, out-of-order/early/future feedback, invalid features,
  nil validator, validator error, and >256 records all failed without exposing
  a partial model.
- Three ordinary-build cost repetitions, 100 reconstructions per size, gave
  observed p95 **0.735–1.316 ms** for 64 input pairs and **0.745–0.762 ms**
  for 256 pairs. Allocations per run were 176 and 562, respectively. The
  nonmonotone small-sample timing reflects measurement noise and different
  retained counts (32 versus 128), not a claimed scaling advantage. Both
  finite p95 screens passed the frozen 100 ms component threshold.

Verification: `go test -race ./internal/researchmemory -count=1` and
`go vet ./internal/researchmemory` pass. Cost command:
`EVENTFRAME_TRANSFER_COST=1 go test ./internal/researchmemory -run '^TestBoundTransferOfflineCost$' -count=3 -v`.
The existing service mixed-mutation regression also passes. SHA-256 of the
uncommitted component source, test, and frozen contract is respectively
`ee58a3b9de107867e0c013ecb758a2a677726f268ac65f88fefa4ec4746d1e3a`,
`23ca018b6f12306e708548e70a63df1f163a53ad18f84a5c9f188b3d7d9f4712`,
and `91e9de6d4289a2ec9f543c829766a2079db10c699219ce84f4d5234eb6ec849d`.
No production path, source store, dependency, remote or serving default changed.

## What remains unproven

The test validator is an explicit synthetic source-A/B filter. The current
durable record binds a journal ID, event ID and old snapshot, but does not by
itself prove the source still exists or that old feature bits are appropriate
under a new snapshot. The Bayesian journal stores a query digest rather than
the original query; the original record does not carry an event-field
fingerprint or feature-contract version. Thus a full feature-equivalence
validator cannot be reconstructed from the current durable record alone.
Source-dependence between labels is also not addressed.

Next, establish a privacy-conscious source/feature validity witness at
admission, audit it against actual retained and deleted store events, and
recheck publication identity around the new-epoch handoff. Then test a
durable replay plus mixed-mutation load; the present result cannot be used
as evidence of goal-6 completion or sub-100-ms serving latency.
