# V75 anchored-prefix preflight results

2026-10-05. **Correctness/resource preflight PASS; whole-cohort equivalence and
scientific adoption NOT ESTABLISHED. All seven whole goals remain OPEN.**
Frozen [protocol](mmm-mean-anchor-v75-protocol.md). Artifacts:
`research/mean-anchor-v75-initial/` and `research/mean-anchor-v75-matched/`.
Original V74 model/reference/study stay immutable. No production or paper change.

## Exact-Computation Fork

Keep all 27 mean maps, three families, three noise states, member/trial caps,
joint same-outcome pair likelihood and raw evidence journal. Store the conditional
filter at the latest revealed ordinal, rather than the last issued ordinal.
New first evidence after that anchor permits incremental filtering; old first
arrivals and second-factor replacements still replay the revealed prefix.
Unknown/canceled suffixes do not add likelihood; next-trial predictions advance
every required issue-clock transition. Older queries still use smoothing.

The latest-row smoother can read the anchored conditional filter directly. Global
hyperstate weights, local noise marginalization, hypothetical branches, zero
support and prospective publication guards are retained. Source-clock consistency
is checked before reading/publishing; a bad anchor cannot publish a partial row.
This changes implementation, not the joint probabilistic model or quality gates.

## Independent And Lifecycle Checks

216 configurations against the separate dense V74 reference: 7,450,056 counted
scalar comparisons, maximum defect 7.772e-15 at unchanged 2e-10 tolerance.
Fixed-mean V72 limits add 8,352 comparisons (tolerance, not bitwise).
Reverse first arrivals, long unknown gaps and canceled holes add 614,124
comparisons across all three sharing modes and hazards 0, 1/16 and 1.
The public 64-trial journal audit adds 42,840 comparisons in 36 configurations
at unchanged 2e-11 tolerance. These counts identify separate checks, not
independent scientific samples or fresh confirmation trajectories.

Actual posterior branches/all-target tower, both class values, same-outcome
forecasts, no-publish, cancellation, future-boundary, epoch/cap/receipt/fault
guards and no-borrowing controls pass. Additional incremental tests verify that
an impossible noise class cannot be revived and invalid anchor/count binding
fails without state changes. Unit/race/vet all exit 0; no frozen attempt edited.

A separate matched public-API fixture uses identical 150-member/16-round
observations for V74 and V75. Final laws/posteriors and all seven query modes at
oldest and latest rows match in 36,834 scalar checks, maximum defect 4.592e-41.
This checks that particular complete fixture, not all 1,440 V74 cohort arms or
every intermediate issued law. Its own unit/race/vet also pass.

## Matched Serial Performance

Apple M4. Same complete fixture, sharing configuration, trial position and query
mode; two 100 ms benchmark repetitions. Setup is excluded from query timing.
All entries have zero bytes/op and zero allocations/op.

| Query Position / Mode | V74 us | V75 us |
| --- | ---: | ---: |
| Latest / model-class | 415.169-415.530 | 68.496-68.649 |
| Latest / noise-class | 414.575-415.247 | 68.165-68.364 |
| Latest / predictive | 1469.003-1469.269 | 1129.953-1130.135 |
| Oldest / model-class | 723.314-724.663 | 727.728-741.821 |
| Oldest / noise-class | 722.555-724.829 | 726.819-727.336 |
| Oldest / predictive | 1677.516-1681.075 | 1684.650-1687.069 |

Latest class queries are about six times faster; latest predictive queries
reduce time about 23%. Oldest queries retain replay/smoothing and are not faster;
the small overhead/variation is preserved, not hidden. These are raw repetition
ranges, not confidence intervals. Comparing a latest-row result to V74's earlier
oldest-row benchmark would overstate the speedup and is not used here.

The initial V75 benchmark measures prediction 28.003-28.760 us: no established
prediction speedup. Constructor allocation in the ordinary unit run is
6,187,520/8,244,240 bytes at 150/200, under the original 8 MiB cap; not RSS,
retained steady-state memory or loaded service performance. All four initial
commands and four matched-fixture commands terminate successfully.

## Remaining Requirements

No full controlled V75 cohort or all-arm equivalence replay has run. No core
<=400 ms gate, durable serving/freshness result or equal-total-cost superiority
is inferred from these timings. V74's failed quality/harm/recovery gates remain
failed; an equivalent computation cannot alter those scientific verdicts.

Next measure the complete cohort/cost and verify unchanged laws, selections and
metrics against frozen V74 with explicit floating-point comparison semantics.
Keep late-factor replacement as a separate proposed optimization, not an
already-tested feature. Distinct adaptive-hyperstate/observation leads, useful
valid splitting and untouched agent outcomes remain in the full objective.
No private/sealed labels, reserved seeds, production, publication or installs.
