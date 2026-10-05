# Full-input uncertainty comparison v119

**FAIL of the complete rescue requirements.** Uncertainty averaging produces
useful specialist improvements, especially with32 labels, but neither candidate
passes its MAP-replacement or broad standalone requirements. Keep every control.
No production or whitepaper promotion; all seven directions remain open.

[Protocol](mmm-soft-learners-v119-protocol.md),
[raw artifact](mmm-soft-learners-v119.json),
[independent summary](mmm-soft-learners-v119-summary.json),
[verifier](../../research/soft-learners-v119-summary.mjs),
[component contract](../../research/variational-logistic-component-contract.md).

## Complete comparison

All2688 runs finish:1344 latent trajectories,21 cases,2 phases,2 schedules,
256 frames each,10 arms. No solver failures or omitted rows. The fresh bases
are2192112000 and2196112100, with6720 distinct effective seeds checked against
the archived blocks. All9 query coordinates and the same eligible64/32 labels
are available to each learner; this is not the partial-view MMM acquisition loop.

Intervals are paired mean +/-3.5SE over32 trajectories. Non-harm allows upper
loss increase<=.01, so passing it does NOT mean zero regression. Gains require
mean>=.005 and lower>0. These are fixed-sample approximate screens, not uniform
or repeated-research population guarantees. All2648 outcomes are retained.

| Candidate | Broad non-harm | Broad gains | Additive/hierarchy target | MAP non-harm | MAP gains |
| --- | ---: | ---: | ---: | ---: | ---: |
| MAP64 | 63/336 | 9/36 | 9/12 | Not applicable | Not applicable |
| MAP32 | 59/336 | 0/36 | 0/12 | Not applicable | Not applicable |
| Tree64 | 224/336 | 3/36 | 2/12 | Not applicable | Not applicable |
| Tree32 | 210/336 | 0/36 | 0/12 | Not applicable | Not applicable |
| Variational64 | 67/336 | 10/36 | 10/12 | 168/168 | 6/40 |
| Variational32 | 66/336 | 8/36 | 8/12 | 168/168 | 36/40 |

Targets are subsets of broad gains, not extra screens. Aggregate1097/2648
requirements pass:1025/2352 non-harm and72/296 gain. Aggregate pass counts are
not an accuracy measure and cannot compensate for failed requirements.

## What improved and what did not

Both variational candidates pass their168 matched-MAP non-harm screens. Maximum
upper harm is.005242 for64 labels and.008887 for32, still nonzero. Variational32
passes36/40 gain requirements against MAP32, failing all four parity-to-majority
terminal cases. Their mean gains are negative: design immediate-.003931,
design delayed-.005525, confirmation immediate-.004522 and confirmation
delayed-.005151. Every corresponding gain interval lies below zero.

Confirmation delayed parity-to-majority terminal Brier is MAP32 .104027 versus
variational32 .109178; the gain interval is[-.007479,-.002822]. This is observed
harm inside the allowed.01 non-harm tolerance, not merely insufficient power.
Variational64 also harms that transition; most of its other failed MAP gains
are positive but below the predeclared.005 floor.

For confirmation delayed null data, all-frame Brier gain over MAP is.008458
[.007831,.009085] for64 labels and.014487[.013778,.015196] for32. That supports
reduced spurious confidence in this fixture. It does not make the model optimal:
terminal null Brier remains.277314 and.299640, respectively, above the.25 floor.
The model prior/likelihood is not the data generator's exact distribution, and
the variational approximation has no general calibration guarantee.

Variational64 passes10/12 additive structural targets; both failures are delayed
gradual changes. Confirmation's mean gain against generic64 is.006354 but its
interval[-.006496,.019204] includes harm. Variational32 passes8/12 targets and
also fails delayed gradual recovery. Linear features remain inappropriate for
the higher-order Boolean tasks; retain their specialists.

Selected confirmation delayed terminal Brier, not a cross-run comparison:

| Case | Generic64 | Boolean64 | MAP64 | Variational64 | MAP32 | Variational32 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Additive stationary | .217755 | .234552 | .204221 | .202686 | .227515 | .221799 |
| Additive abrupt | .230311 | .238044 | .217122 | .215251 | .222184 | .216865 |
| Additive gradual | .248883 | .248807 | .245612 | .242529 | .244871 | .237633 |
| Hierarchy gradual | .255356 | .251695 | .273912 | .269590 | .292658 | .281352 |
| Local-table gradual | .257130 | .254936 | .278758 | .274381 | .296769 | .285592 |
| Parity4 | .071773 | .048168 | .292012 | .286895 | .317352 | .304476 |
| Parity to majority | .112301 | .201407 | .111727 | .114290 | .104027 | .109178 |

## Audits and limits

The independent verifier checks170 source hashes, all scores and teacher laws,
eligible origins, pairing and seeds. It reconstructs1376256 MAP predictions,
1376256 variational predictions with a separate wider-domain Simpson integral,
43008 MAP fitted states and43008 variational fixed-point states. Mean/covariance
equations, inverse/symmetry, bound and finite moments pass. Maximum fresh-fit
iterations are83 and66 for64/32. This is not an exact-posterior error certificate.

Generation passes in607.04s; full2688-run exact replay passes in609.34s with
all170 source hashes. The independent summary reproduces byte-for-byte. The raw
exclusive0600 artifact is311726024 bytes, SHA256:
`98d11f0fc346f6e0ebcd3dad329e68b37ba86ae41db29601973892dde203cf57`.
Focused race/as-of/seed tests pass in11.974s and scoped vet passes. Generation
duration is not serving latency. Component timings are in the linked contract;
no quality or runtime claim is made for production integration.

## Consumed composition diagnostic

A separate [script](../../research/soft-learners-v119-composition-diagnostic.mjs)
and [output](mmm-soft-learners-v119-composition-diagnostic.json) reuse these
consumed forecasts. This is diagnosis, not untouched confirmation. It compares
the four incumbent raw experts with six-expert MAP and variational additions,
using generic prior.95, remaining mass split equally, transition.001 and an
as-of refilter of published forecasts. No family IDs or oracle enter weights.
The fixed forecast tape is not an adaptive observation or serving integration.

All2064384 diagnostic forecasts normalize and1008 unavailable-outcome mutation
checks pass; the diagnostic reproduces byte-for-byte. Variational additions meet
the.01 non-harm screen in166/168 cells
against the four-expert mixture and168/168 against MAP additions. The two failures
against four experts are delayed parity-to-majority terminal cells in both
phases. No cell meets the.005 positive-lower gain rule against MAP additions.
Confirmation delayed stationary-additive gain versus the four-expert mixture is
.009938[.003742,.016134], but the extra benefit versus MAP additions is only
.002054[.000169,.003940]. This is not an established integrated rescue.

Next investigate the distinct [segment-posterior lead](../../research/segment-posterior-lead.md)
for stale-evidence recovery. Its tiny exhaustive check validates only a finite
recurrence, not richer learners or quality. Preserve the specialist benefit and
the failed transition; do not tune the ridge prior on these confirmation data.
