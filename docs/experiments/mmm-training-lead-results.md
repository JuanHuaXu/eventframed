# Effective training lead

Status: mechanism diagnostic complete. Not an efficacy result. All2688 consumed
records matched; each paid policy makes41664 queries across the delayed runs.
Immediate runs make none. Windows64 and32 have identical first-inclusion counts.

| Policy | Missing label added | Earlier than natural | Already naturally available at first fit | No fit in horizon |
| --- | ---: | ---: | ---: | ---: |
| Random | 8535 | 8689 | 19064 | 5376 |
| Entropy | 8294 | 8701 | 19293 | 5376 |
| Disagreement | 8374 | 8611 | 19303 | 5376 |

Only41.34%,40.79%,40.77% respectively provide new/earlier TRAINING evidence.
The others may still accelerate mixer feedback. Do not call them universally
wasted labels or infer their causal effect on accuracy from this count.

For disagreement, average wait from paid reveal to first fit is18.56 frames
among included queries. By query clock modulo32:

| Phase | Wait after reveal | Fraction giving new/earlier training evidence |
| --- | ---: | ---: |
| 8 | 23 frames | 26.59% |
| 16 | 15 frames | 46.71% |
| 24 | 7 frames | 66.85% |
| 0, just after publication | 31 frames | 20.37% |

At phase0, none of the naturally delivered queried labels reaches training ahead
of natural delivery. The benefit there comes solely from originally missing
labels. Four late queries per delayed trajectory never enter a fit because the
last publication is224; they can still affect the mixer before scoring ends.

This supports an isolated timing experiment, not increasing budgets or claiming
that moving queries will improve predictions. Shift only the7 boundary queries
from clocks32m to32m-1, after that clock's prediction and before the next fit.
Keep their same origin block [32m-8,32m), now fully visible including the current
input; do not consume its unknown outcome during selection. Other query clocks,
learner priors, caps and publication times stay fixed. Check equal actual costs.

Artifacts: [protocol](mmm-training-lead-protocol.md), [results](mmm-training-lead.json),
[replay](mmm-training-lead-replay.json). Class totals and reveal-before-fit rules
checked across every query; no inference of independent samples from these counts.
Full diagnostic replay is byte-identical (`cmp` exit0).
