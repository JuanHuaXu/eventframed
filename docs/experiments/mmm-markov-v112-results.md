# v112 delayed Markov advice: protected, insufficient recovery

**FAIL the complete quality screen:705/718 gates.** All672 non-harm gates pass;
33/46 gain gates pass. The remaining13 failures are delayed late recovery:
six parity-to-majority comparisons in each phase, plus confirmation
majority-to-parity against original log/no-neutral.

This is a useful distinction from v111's failed protection, but not promotion.
The experiment changes the calculation of delayed evidence without establishing
all required gains. Exact inference for a declared model does not imply that
the model is sufficiently useful or well specified.

## Complete frozen experiment

The [protocol](mmm-markov-v112-protocol.md) covers768 fresh latent trajectories,
two paired schedules (1,536 runs),12 cases, both phase-disjoint rule pools,
256 scored frames and eight policies. The seven original generic/Brier/log
controls remain unchanged. Alpha=.001 was fixed before generation; there was
no outcome-driven rate selection, interim tuning or optional sample enlargement.

Intervals are paired mean +/-3.5SE over32 trajectories. They are approximate
fixed-sample screens, not confidence sequences or history-wide coverage.
Schedules sharing a latent trajectory are not independent replicates.

## Confirmation, delayed late segment

Brier is lower-is-better. Late means every frame128-255, not just the final
well-adapted block.

| Case | Generic | Conservative | Arrival Brier | Original log | Markov |
| --- | ---: | ---: | ---: | ---: | ---: |
| Stable parity4 | .071506 | .069472 | .058334 | .050957 | .051431 |
| Majority to parity | .257822 | .254492 | .255628 | .219778 | .215848 |
| Parity to majority | .229523 | .228700 | .230773 | .223884 | .220822 |

The Markov gain over original log in majority-to-parity is.003930 with
interval[-.000829,.008688]. Its reverse-switch gain is.003062 with
interval[.000306,.005819]. The latter has a positive interval, but its mean
still falls short of the predeclared.005 requirement. Neither is a complete
required improvement. Against generic on the reverse switch, mean gain.008700
has interval[-.003312,.020712], so uncertainty remains even with a larger mean.

Expected accuracy on these particular delayed late trajectories is94.43% for
stable parity,68.76% for majority-to-parity and71.36% for parity-to-majority.
Original log gives94.43%,68.63% and70.81%. These are finite synthetic accuracies,
not agent-task accuracy or a universal94.7% guarantee. The small stationary
Brier regression remains visible rather than being hidden by equal accuracy.

The13 failed gains consist of design reverse-switch comparisons against
controls0..5, confirmation reverse-switch comparisons against controls0..5,
and confirmation forward-switch comparison against control5. Every failed gain
and every passed protection condition is retained in the
[independent summary](mmm-markov-v112-summary.json).

## Verification

- Seven-control compatibility on eight consumed cases plus effective-seed
  audits PASS under race in6.424s. The final artifact environment-variable name
  was corrected before fresh generation; no inference kernel changed.
- Full generation PASS194.27s. All1,536 records replay exactly in197.65s.
- Independent reconstruction checks37 source hashes, all issued scores/costs,
  as-of training origins, feedback clocks and journal accounting. Regenerated
  summary is byte-identical.
- All196,608 immediate-feedback forecasts, masks and costs match original
  log/no-neutral exactly across all768 immediate trajectories. The candidate's
  divergence from that policy is confined to delayed/missing feedback here.
- Package vet and whitespace checks pass. Component state-path enumeration,
  dense filtering, checkpointing, rollback and race tests are recorded in the
  [component report](../../research/delayed-markov-component-results.md).

The [raw artifact](mmm-markov-v112.json) is347,029,783 bytes, created0600.
SHA256 `8b386bbd6114e7890301742f32b37cea311513b58af87416765575bf6684e0f1`.
No record, gate, source kernel or outcome was edited to obtain a pass.

## Performance

After quality, full replay and independent verification were terminal:

```sh
go test ./internal/observationlearners -run '^$' -bench '^BenchmarkMarkovV112WholeFixture$' -benchmem -benchtime=1x -count=3
```

Apple M4, darwin/arm64, Go1.27.1, suffix10. The whole eight-policy fixture on
consumed design parity4 takes124.23-151.20ms immediate and122.18-124.42ms
delayed/missing, allocating about26.9MB per operation. All six repeats, including
the first/cold repeat, are retained in the [raw output](mmm-markov-v112-benchmarks.txt).
This includes fits,256 frames and all eight policies, not one serving request.
Benchmark source SHA256:
`febb09a8dfc493134901179f894cbe28972abd70dd57210ba12f986f18c55d1a`.

The earlier paired component benchmark already showed the extra O(D*M)
refiltering cost. The complete quality improvement remains insufficient to
justify production adoption solely from these results.

## Next lead, not a rescue claim

Investigate [uncertainty over the transition rate](../../research/hazard-mixture-advice-proposal.md)
using a fixed model family and joint-state likelihood, rather than choosing a
favorable alpha from this dataset. The proposal explicitly preserves likelihood
normalizers, delayed ownership and existing gate authority. A wrong fixed rate
is only a hypothesis for the remaining failure; model and acquisition limits
remain alternative explanations.

All seven research directions remain open. No production/OpenClaw calls,
private-data expansion, commits, pushes or whitepaper promotion.
