# V60: member-local escape passes inference, fails scientific rescue

## Verdict

The [frozen protocol](mmm-tree-v60-protocol.md) was executed completely: all
40 consumed V54 worlds, three delay schedules and eight arms, 960 arms and
96,000 underlying outcomes. Independent replay completed successfully. This
is a diagnostic n=1 per cell, not fresh confirmation. All seven WHOLE goals
remain OPEN/ACTIVE. No production, private data, whitepaper or publication changed.

V60 removes V59's fixed-range and same-baseline coupling restrictions by adding
an independent member branch at terminal leaves. That is an expressiveness and
inference success, not an accuracy result. All five paired policies still harm
Adaptive by more than .01 in 105/120 cells; all miss the mean .01 Full-relative
gain. Recovery is 2.383333 rounds slower than Adaptive over 60 changing cells,
with no recovery improvements over the new no-pair model. Scientific rescue FAILS.

## Whole-loop results

Lower Brier is better; higher top-10 utility is better. The reported issued
Brier integrates the clean next-outcome law against simulator probabilities,
not the noisy first measurement. Receipts separately preserve their appropriate
measurement forecast. A requested W2 measures the SAME Y, not a new trial.

| Arm | Issued Brier | Terminal Brier | Top-10 utility | Total loop seconds | Maximum loop ms |
| --- | ---: | ---: | ---: | ---: | ---: |
| Full | .224699096 | .208151654 | .727150389 | 7.911 | 90.262 |
| Adaptive | .214747855 | .181598705 | .815037852 | 31.249 | 322.510 |
| No pair | .251827031 | .241870341 | .637066685 | 7.253 | 79.180 |
| Random | .248930579 | .237159099 | .664686983 | 9.021 | 90.621 |
| Uncertainty | .248417739 | .236522481 | .674686983 | 16.408 | 143.947 |
| Information | .248416522 | .236599177 | .669186983 | 16.422 | 154.275 |
| Noise-class concentration | .248390678 | .236687398 | .671186983 | 16.401 | 142.461 |
| Prediction value | .248572105 | .236898173 | .668186983 | 39.090 | 359.042 |

The compute component passes: every complete loop is below the unchanged
400 ms cap; constructor allocation is 2,716,544 bytes, below 8 MiB. This is
allocated memory, not RSS, and a serial research fixture, not loaded serving
latency or background-learning freshness. Three warm benchmark repetitions:
Predict 215.1-218.5 ns, no allocations; 150-origin concentration batch
3.820-3.851 ms, no allocations; prediction-value batch 16.716-17.049 ms,
12,040 bytes/12 allocations. Constructor, issue, expiry, candidate integration,
requests, replies, snapshots and drain are included in complete-loop timing.

Concentration's issued-risk gains over random/uncertainty are only .000539901
and .000027061. It costs approximately 81.82% more than random. Prediction value
regresses uncertainty by .000154366 and costs 39.090 versus 16.408 seconds.
Equal request budgets do not establish equal TOTAL cost. Goal 7 is not validated.

## Comparisons and audit

All populations match V57/V58/V59, and all 240 Full/Adaptive arms in each
comparison are bitwise equal after removing timing fields. Against V59,
prediction value improves issued Brier by .000087204, terminal Brier by
.000547491 and utility by .012213306, winning 65/losing 55 cells. This is a
small partial change, not rescue. Every new learner wins only 3/loses 117
cells against V58. Against V57, prediction value worsens issued Brier by
.030789304 and utility by .142622681. Cross-run time changes are descriptive:
control times also moved, and the prediction-value microbenchmark is slower
than V59. No causal throughput improvement is inferred from total times.

Ten model test roots race-pass. Independent checks enumerate 26 small-tree
alternatives and 1,515 terminal latent states; delayed suffix rebuilding checks
4,656 forecasts and 120 prediction values. Three fixture roots race-pass,
including 18 future forks, 52 corruption rejections, 3,840 distinct seeds and
23,040 domain-separated channels. Full replay checks all 960 arms and 288,000
prediction-value entries. Compiler closure is 83 inputs plus 16 support files;
all 99 frozen originals and copies match. Eight serialized commands terminated
with exit 0; full collection took 145.87 seconds and audit 393.31 seconds.

Each paired policy requests 48,000 W2 packets, including unavailable or expired
ones. Missing totals: random 1,053; uncertainty 982; information 987;
concentration 998; prediction value 979. Expired totals respectively:
2,154/2,002/1,984/1,968/2,012. They remain charged and acknowledged and cannot
resurrect expired model factors. The reserved design/confirmation seeds and
sealed outcome-labeled agent cohort were not opened.

Raw evidence: [readback](../../research/tree-v60-diagnostic/readback.json),
[completion commands](../../research/tree-v60-diagnostic/completed.json),
[V59 comparison](../../research/tree-v60-diagnostic/comparison-v59.json),
[V58 comparison](../../research/tree-v60-diagnostic/comparison-v58.json),
[V57 comparison](../../research/tree-v60-diagnostic/comparison.json).

## Post-collection falsifiers

The [branch audit](../../research/tree-v60-diagnostic/branch-audit.json)
reconstructs all 108,000 final learner forecasts using observed W1 and only
requested, available, retained W2. Hidden rates are used only to score the
counterfactual, never to build beliefs. The hypothesis that the root simply
never uses the local branch is not confirmed: prediction-value mean root-stop
weight is .243761; only 7/120 cells exceed .99. Pure independent member models
are WORSE by .034411 terminal Brier on that same selected stream. A direct-root
independent alternative gains just .000575556. None selects its own evidence.

Fixed-selected-stream final-suffix probes preserve the prior and compare 600,
1,200 and 2,400 issued positions. Each member typically has four, eight or
sixteen first observations, respectively. Longer counterfactual suffixes include
old expired W2 only when actually requested, available and within the new suffix.
Probability-shape checks are NOT equality checks against the 600-position law.

| Prediction-value suffix | Terminal gain over 600 | Wins/losses | Late-regime mean gain | Recurring mean gain |
| --- | ---: | ---: | ---: | ---: |
| 1,200 | .010175798 | 103/17 | -.018155711 | -.010358669 |
| 2,400 | .010105482 | 88/32 | -.028334960 | -.012014494 |

The 2,400-position probe also harms abrupt/gradual regimes on average. Thus
increasing retention blindly is falsified as a general rescue. These probes
are final-state diagnostics ONLY: no new prequential forecasts, recovery,
runtime, policy, untouched confirmation or goal validation. Full per-cell
negative results remain in [1,200](../../research/tree-v60-diagnostic/window-probe-1200.json)
and [2,400](../../research/tree-v60-diagnostic/window-probe-2400.json).

Next: a bounded, member-specific, origin-aware retention selector. Earlier
window-bank and delayed fixed-share failures remain relevant controls, not
forgotten evidence. Do not compare raw marginal likelihoods of differently sized
suffixes, tune away failing regimes, or advertise mathematical normalization as
validated intelligence. Useful certified splits, untouched agent utility,
loaded serving/freshness and equal-total-cost observation all remain required.
