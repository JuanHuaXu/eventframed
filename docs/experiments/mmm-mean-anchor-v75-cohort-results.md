# V75 complete-cohort results

2026-10-05. **Cost improves; strict all-arm trace equivalence FAILS. Scientific
adoption FAILS. All seven whole goals remain OPEN.**
Frozen [protocol](mmm-mean-anchor-v75-cohort-protocol.md), raw records and command
logs: `research/mean-anchor-v75-diagnostic/`. No production or publication change.

## Complete Controlled Collection

All 3,600 arms finished: forty consumed worlds, twenty regimes, two geometries,
three delay schedules. All forty populations and all 2,160 old controls match
V74 exactly outside costs. The 1,440 new arms retain all 27 means, three rate
families and three noise states. No pruning, gate changes or failed-arm removal.
This is one trajectory per cell, not untouched confirmation.

All twelve new arms still FAIL all original scientific and runtime gates: .01
mean issued-Brier improvement over Full, no cell with more than .01 harm against
Adaptive, positive restricted recovery gain, and <=400 ms core in every cell.
Each new arm exceeds 400 ms in all 120 cells. Best issued Brier remains
.21958154845158395, versus Full .2246990964333006 and Adaptive .2147478547980669.
The best arm has 27 harm cells and recovers .9 rounds slower than Adaptive.

| Arm | Issued Brier | Mean Core ms | Worst Core ms |
| --- | ---: | ---: | ---: |
| mean_no_pair | .220790525 | 760.554 | 808.727 |
| mean_random | .219588796 | 934.826 | 1041.198 |
| mean_uncertainty | .219581548 | 1216.294 | 1605.951 |
| mean_falsification | .219603632 | 1267.085 | 1645.440 |
| meanlocal_no_pair | .220914952 | 681.212 | 732.129 |
| meanlocal_random | .220262535 | 841.334 | 944.422 |
| meanlocal_uncertainty | .220111272 | 1123.458 | 1489.645 |
| meanlocal_falsification | .219701898 | 1174.409 | 1553.489 |
| meanindividual_no_pair | .252404948 | 651.996 | 693.707 |
| meanindividual_random | .251555738 | 807.112 | 912.341 |
| meanindividual_uncertainty | .248475088 | 1089.984 | 1459.128 |
| meanindividual_falsification | .249341227 | 1141.045 | 1520.540 |

Common-noise uncertainty mean core falls from V74's 2039.289 to 1216.294 ms
(40.36% reduction); worst core falls from 2199.443 to 1605.951 ms. These are
serial synthetic full-core measurements, not loaded daemon serving or freshness.

## Equivalence Boundary And Near-Tie Diagnosis

The frozen recursive comparator checks every non-cost scalar at absolute
tolerance 2e-10 and every discrete field/order exactly. Its eleven self-tests
pass. Across 54,449,760 scalar checks, 1,280 new arms pass the full comparison;
160 FAIL. Native nomination order differs in 1,696 rounds. A separate,
postcollection diagnostic finds changed candidate sets in 155 arms / 1,593
rounds; 103 changed rounds are order-only. **This is not blanket policy-trace
equivalence, and the diagnostic does not sort away or erase failures.**

Before the first changed nomination, 6,434,470 clean/first-observation issued-law
comparisons have maximum defect 4.441e-16. Across all 6,912,000 issued-law
comparisons, including diverged evidence histories, maximum defect is 6.583e-12;
maximum snapshot-law defect is 3.553e-14 and final-risk defect 9.493e-15.
Changed-round query-score defect reaches 2.334e-11; the maximum score span among
swapped members is 6.172e-14. This supports near-tie numerical sensitivity as a
diagnosis, not permission to replace the frozen selection rule retrospectively.

Matching audit forecasts by trial yields 420,584 diagnostic comparisons with
maximum defect 9.865e-13; 11,416 audit trials are unmatched because nominations
changed. Joined metrics cannot prove complete trace equivalence. The recursive
comparator's maximum mixed-field defect of 438 is a discrete trial-ID difference,
NOT a probability defect.

Nomination readback: `nomination-audit.json`; exact script and both raw-data
hashes are recorded there. Core readback independently recomputes losses,
recovery, observations, receipts, snapshot utility and cost sums. There is no
fresh dense independent replay of all 1,440 new arms and no fresh detailed
three-arm reference replay in this collection. Earlier unit/reference evidence
retains only its original scope.

## Cost And Remaining Requirements

For common-noise uncertainty, issue takes 22.57%, first resolution 33.13%,
nomination 23.22%, second resolution 10.85%, snapshot 5.77% of core elapsed time.
These are instrumented phase fractions, not a causal profiler or single proven
bottleneck. Constructor allocation 6,187,520 / 8,244,224 bytes at 150 / 200
members fits the original 8 MiB cap; not RSS or long-run retained memory.

| Falsification Configuration | Brier Gain vs Random | vs Uncertainty | Core Cost Ratio vs Random | vs Uncertainty |
| --- | ---: | ---: | ---: | ---: |
| Common noise | -.000014836 | -.000022083 | 1.355424 | 1.041758 |
| Local noise | .000560637 | .000409374 | 1.395889 | 1.045352 |
| Individual | .002214511 | -.000866140 | 1.413738 | 1.046845 |

Equal 400 second-measurement requests are not equal TOTAL cost. No Goal 7
superiority follows. All four frozen collection commands exited zero. A separate
V76 at-anchor conditional-factor replacement tests a cost lead without changing
the probability model; its preflight cannot rescue failed quality criteria.
Useful valid splits, untouched labeled agent tasks, adaptive stability and
durable loaded freshness remain necessary. Reserved confirmation seeds, sealed
labels, private sessions, production and whitepaper/publication remain untouched.
