# V74 joint mean/dispersion results

2026-10-05. **Scientific adoption FAILS. All seven whole goals remain OPEN.**
Frozen [protocol](mmm-mean-joint-v74-protocol.md); raw outcomes, independent
fixture traces, command logs, readback and cost audit are in
`research/mean-joint-v74-diagnostic/`. No production, paper or publication changes.

## Complete Controlled Screen

Forty consumed worlds, twenty regimes, two geometries, three delay schedules:
3,600 arms. All 2,160 prior V72 arms match bitwise outside timing, and all forty
populations match. The 1,440 new arms use the full 27 mean maps, three rate
families and three noise states. No grid pruning, gate changes or failed-arm
removal. This is one trajectory per cell, not fresh confirmation.

All twelve new arms FAIL every original gate: mean issued Brier gain at least
.01 over Full, no cell exceeding .01 harm against Adaptive, positive mean
restricted recovery gain against Adaptive, and core time <=400 ms in every cell.
Full issued Brier is .224699096; Adaptive is .214747855. Lower Brier is better.
Recovery gain below zero means slower than Adaptive; it includes restricted
miss penalties, not just delays conditional on detected recovery.

| New Arm | Issued Brier | Harm Cells / 120 | Recovery Gain (Rounds) | Mean Core ms | Worst Core ms |
| --- | ---: | ---: | ---: | ---: | ---: |
| mean_no_pair | .220790525 | 28 | -.7000 | 1123.992 | 1169.878 |
| mean_random | .219588796 | 27 | -.8667 | 1365.056 | 1427.813 |
| mean_uncertainty | .219581548 | 27 | -.9000 | 2039.289 | 2199.443 |
| mean_falsification | .219603632 | 27 | -.8833 | 2094.959 | 2251.041 |
| meanlocal_no_pair | .220914952 | 31 | -.8833 | 1084.009 | 1134.612 |
| meanlocal_random | .220262535 | 27 | -.8833 | 1310.109 | 1386.349 |
| meanlocal_uncertainty | .220111272 | 28 | -.9000 | 1990.150 | 2173.884 |
| meanlocal_falsification | .219701898 | 28 | -.9167 | 2041.240 | 2186.213 |
| meanindividual_no_pair | .252404948 | 110 | -1.8833 | 1055.380 | 1094.319 |
| meanindividual_random | .251555738 | 110 | -1.8500 | 1276.204 | 1349.708 |
| meanindividual_uncertainty | .248475088 | 110 | -1.7167 | 1954.427 | 2093.916 |
| meanindividual_falsification | .249341227 | 112 | -1.7000 | 2008.698 | 2174.355 |

Each new arm exceeds 400 ms in all 120 cells. Runtime columns measure the full
serial synthetic learner core, not one prediction, durable loaded serving,
input/output latency or the eventframed production daemon.

## What Improved And What Did Not

The best new arm, common-noise mean_uncertainty, improves the earlier best
fixed-mean/local-noise component .245903739 -> .219581548 (10.7043% lower loss).
This is a joint-model comparison, not an isolated proof that mean error was the
unique cause. It improves Full by .005117548, short of .01; it still loses to
Adaptive by .004833694, has 27 harm cells, and recovers .9 rounds slower.
Stationary mean/max harm is .004742699/.014200052, so stationary protection
also remains unestablished. The best arm's terminal risk .190090998 and packet
usefulness .811794587 still trail Adaptive .181598705/.815037852.

Independent member mean/family models perform worse; eliminating all borrowing
is not a sufficient rescue. Common and local noise alternatives remain explicit.
Static mean maps and dispersion beliefs are not adaptive hyperstate inference.
The maps depend on synthetic member indices, not actual EventFrame semantics.

Falsification versus matched random/uncertainty policies:

| Configuration | Brier Gain vs Random | Gain vs Uncertainty | Core Cost Ratio vs Random | Ratio vs Uncertainty |
| --- | ---: | ---: | ---: | ---: |
| Common noise | -.000014836 | -.000022083 | 1.534705 | 1.027299 |
| Local noise | .000560637 | .000409374 | 1.558069 | 1.025671 |
| Individual | .002214511 | -.000866140 | 1.573963 | 1.027768 |

Every paired policy requested 400 same-outcome second measurements per arm.
Equal requests are NOT equal total cost. No Goal 7 superiority is established.
Cross-configuration contrasts in raw readback are not matched acquisition tests;
the separate cost audit reports only matching model configurations.

## Correctness Evidence And Its Boundary

Initial preflight: 216 configurations, 7,450,056 counted scalar comparisons
against an independently implemented dense-transition/urn-DP joint reference;
maximum defect 7.772e-15 under the original 2e-10 tolerance. Another 8,352
fixed-mean comparisons verify the V72 limit (tolerance, not bitwise). Actual
posterior branches/all-target tower, same-outcome likelihoods, cancellation,
no-publish, receipt/epoch/cap/numeric-fault guards and independent-member
no-borrowing checks pass. Full 64-trial journals add 36 configurations and
42,840 counted external-API comparisons at 2e-11 tolerance. Unit/race/vet pass.

Three complete 150-member fixture arms receive fresh independent issued-law,
observation-value, selection, receipt, snapshot and metric replay: 7,200 issued
packets with clean AND first-observation forecasts (14,400 scalar issued-law
comparisons, plus other checks). Nonvacuous future-data forks and all 29 injected
corruptions are detected. The cohort readback recomputes all truth risks,
recovery, snapshot usefulness/bias, observed Brier and cost arithmetic.

**There is no full independent replay of all 1,440 new cohort arms.** The earlier
V72 all-arm equivalence proof cannot be transferred to this changed model.
The three fresh fixtures and broad mathematical/journal tests support the
implementation at their stated scope; they do not turn this screen into
confirmation or all-seven-goal validation. All six frozen main commands exited 0.

## Resources And Next Action

Constructor allocation: 6,179,328/8,241,576 bytes at 150/200 members, below the
unchanged 8 MiB cap; not RSS, retained long-run memory or a loaded-service test.
Serial prediction 28.001-28.155 us, model-class 720.057-723.118 us, noise-class
718.256-721.199 us, predictive query 1666.905-1669.157 us: all zero allocation.

Cost accounting identifies repeated filtering/smoothing as viable work to
remove, not a proven unique bottleneck. For mean_uncertainty, first resolution
uses 35.98%, proposal 33.19% and issue 15.13% of elapsed core time. The separate
V75 anchored-prefix fork preserves the complete model/journal and tests exact
law equivalence before timing. A runtime rescue cannot rescue failed quality.
Further distinct quality leads are recorded in
`research/mean-joint-v74-next-leads.md`; none is adopted or certified here.

Useful error-controlled splitting, untouched labeled agent tasks, durable loaded
freshness and equal-total-cost observation superiority remain necessary. No
reserved confirmation seeds, private/sealed labels or production were touched.
