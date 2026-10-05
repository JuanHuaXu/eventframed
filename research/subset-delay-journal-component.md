# Conservative delayed subset journal

## Boundary and decision

This is an isolated test-only adapter in
`internal/observationgate/subset_delay_journal_test.go`. It enables a future
delayed-feedback forest comparison; it is not a demonstrated learning rescue.
No production code or existing captured experimental source was changed.

The existing subset state deliberately permits one outstanding prediction.
Overwriting it would lose historical advice; simply waiting would change the
event stream. A bounded journal separates issuance from settlement instead.
Earlier v104/v106 results show that unconditional stale-role carry can hurt
recovery; this adapter therefore starts with conservative generation scoping.

## Contract

- Single owner; 64 outstanding slots with explicit backpressure, not eviction.
- Capture original outer and inner forecasts, observed inputs and generation.
  Predictions advance the issuance clock without manufacturing feedback.
- Buffer out-of-order arrivals and settle in origin order. Only an external,
  predeclared expiry policy censors a missing origin. Ready evidence is retained.
- Publication requires increasing model versions. Published model objects must
  remain immutable; the adapter does not clone or authenticate them.
- Publication and pooled-to-local splitting advance the generation. Old advice
  is counted as stale, never used to weight current models or authorize a split.
- Delivery and expiry use copy-before-commit, so rejected operations cannot
  partially mutate state. Settled identities cannot be redelivered.
- Input availability and authentic outcome arrival are responsibilities of the
  experiment driver. The journal has no independent wall clock and cannot prove
  an externally supplied label was available when its caller delivered it.

Expiry, missingness and stale labels are not negative outcomes. The next driver
must score all issued predictions from an evaluator-only outcome tape, including
those whose labels are unavailable to the learner. Fitting must use only arrived
labels, ordered by origin; full-frame training access must be costed separately.
Keep forest and uniform-input controls on identical schedules and initial fits.

## Verification

`go test -race ./internal/observationgate -run '^TestSubsetDelay' -count=1`
passed in 1.367 seconds. Four tests cover:

- Exact 192-step immediate prediction/state parity, including an authorized split.
- Capacity rejection, duplicate rejection, expiry preserving ready evidence,
  stale publication feedback, ring reuse and complete settlement accounting.
- Split invalidation of already-issued advice, including buffered ready feedback.
- Reverse-order arrival compared with literal original-forecast updates;
  repeated publication and future expiry reject without mutation.

`go vet ./internal/observationgate` passed. Race execution does not establish
multi-owner safety: this adapter is explicitly single-owner.

## Measured cost

Apple M4, darwin/arm64, Go benchmark suffix -10. Three 200ms repetitions:

| Fixed-model prediction plus feedback | ns/op | bytes/op | allocations/op |
| --- | --- | --- | --- |
| Direct | 8513 / 8720 / 8628 | 5903 / 5903 / 5904 | 7 |
| Journal | 9009 / 8920 / 8934 | 6095 / 6095 / 6095 | 8 |

Command:
`go test ./internal/observationgate -run '^$' -bench '^BenchmarkSubsetDelayJournal$' -benchtime=200ms -count=3`

Each benchmark calibration starts from a fresh state copy. These measurements
include the existing reader/observer but exclude model fitting, storage, network,
delayed scheduling and tail latency. The bounded array copy adds work proportional
to the declared capacity, not an unbounded event-history scan. No sub-100ms
end-to-end serving or delayed-quality claim follows from this component result.

All seven research goals remain open. Next: a frozen immediate/delay/missingness
comparison with forecast replay and audited fit-origin lists, retaining failed
gains and charging model-learning costs rather than treating them as free.
