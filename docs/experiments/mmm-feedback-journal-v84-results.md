# Bounded forecast journal v84 results

PASS of the lifecycle prerequisite and immediate-feedback parity checks. This
does not establish accuracy under delayed/missing labels; that experiment is
still required. No production serving or durable storage changes were made.

[Protocol](mmm-feedback-journal-v84-protocol.md),
[stress artifact](mmm-feedback-journal-v84.json),
[journal](../../internal/observationgate/feedback_journal.go),
[tests](../../internal/observationgate/feedback_journal_test.go).

The journal reuses the frozen v82 forecast/update transitions but separates
issuance order from feedback arrival. Immutable emitted probabilities are kept
in64 bounded origin-indexed slots. A full live slot rejects issuance before any
reader access. Missing feedback requires explicit censoring or expiry; it never
becomes a negative label. One stream owner serializes all operations.

## Lifecycle evidence

Twenty seeded2048-issuance schedules produce40,960 forecasts. Delays are0..31,
20% of labels are independently missing, model versions advance every64 issues,
and entries older than the declared48-origin retention window are expired.
End-of-stream outstanding entries are explicitly censored, not assigned labels.

| Outcome | Count |
| --- | ---: |
| Issued | 40,960 |
| Applied selector updates | 24,684 |
| Consumed as stale | 7,809 |
| Censored/expired | 8,467 |
| Pending after final censoring | 0 |
| Unexpected rejected first deliveries | 0 |
| Maximum simultaneous pending entries | 35 |

Accounting holds after every operation: issued equals applied plus stale plus
censored plus pending. Every attempted duplicate delivery is rejected without
mutation. Separate unit tests fill all64 slots and verify backpressure without
reading or overwriting, then reuse slots and reject old origin IDs.

Stale feedback is substantial:7,809 of32,493 received first deliveries, about24%.
Those labels are not claimed false or useless. They are withheld from current
selector updates because their model version or pooled/local identity changed.
A separate training owner may admit observed labels under its audit policy.
Whether this conservative skip policy preserves enough learning is an open
experimental question, not a free correctness improvement.

## Invariants checked

- Enabled and disabled subset states match the frozen immediate-feedback
  forecasts and weight updates through1024 steps, including publications and
  a split. Issuance itself performs no outcome-dependent weight update.
- Out-of-order feedback updates from each origin's saved expert probabilities,
  not the newest prediction. Issuance order advances independently.
- Model publications are monotone and store immutable references. Old-version
  feedback is consumed as stale without current weight or split mutation.
- A pooled-to-local split advances a generation; pending pre-split feedback
  cannot give the new local slot credit for old pooled predictions.
- Failed reads, invalid origins, duplicate deliveries, reused ring IDs and
  invalid expiry cutoffs leave state unchanged.
- Censoring/expiry changes only pending/accounting state, never applies a label.

Feedback-arrival updates are online selector updates. They are not presented
as an exact fixed-model Bayesian posterior in original event order. Retaining
historical selector weights across model publications remains the inherited
algorithm; the journal does not prove its delayed-learning optimality or
calibration.

## Verification and cost

Artifact SHA256:
`bee90a0ff0b81f475d62d5b6191c62721203b7ae6929051c117ed98e9e1478d4`.
The artifact contains117 source hashes and per-seed prediction/update tape hashes.
Full stress execution with the race detector passes in6.409 seconds. Exact
non-race stress replay passes in0.598 seconds. Immediate/lifecycle race tests
pass in1.469 seconds; package vet passes.

Apple M4, darwin/arm64, GOMAXPROCS10, three500ms baseline-model microbenchmarks:

| Boundary | Time | Bytes/op | Allocations/op |
| --- | --- | ---: | ---: |
| Direct immediate forecast/update | 9.773-9.899 us | 11207 | 33 |
| Journal immediate forecast/delivery | 9.834-9.899 us | about11207 | 33 |

The ranges overlap. Do not claim a measured production latency bound or a
statistically established overhead percentage from these three repetitions.
These benchmarks exclude model fitting, persistence, networking and concurrent
serving. Issuance/delivery/censor lookup are constant time in the bounded ring;
expiry scans64 slots. Predictor cost remains additional. This implementation is
single-owner, in-memory research code, not thread-safe shared storage.

## Next experiment

Run the frozen retained-subset and count controls with equal delayed/missing
feedback schedules. Fit only from labels actually received; count stale-weight
skips separately from training eligibility. Freeze event/arrival order, expiry,
end-of-stream handling and model-publication timing before generating outcomes.
Require immediate-feedback parity, no ghost labels, and actual Brier/accuracy
and stationary protection on fresh streams. If skips or lag erase the earlier
gain, retain the failure and test a specifically justified policy next. All
seven research directions remain open.
