# Retained learner v8 results

2026-09-12. [Protocol](mmm-retained-v8-protocol.md),
[summary](mmm-retained-v8-summary.json), [journal](mmm-retained-v8.json.gz).

480 streams, four paired arms, three base-fit seeds independent of the
evaluation seeds. For each scenario/fit, the same fitted base is reused across
design and confirmation; only evaluation stream seeds change between phases.
The study uses synthetic six-coordinate observation with equal post-forecast
full-field audit budgets.

## Frozen verdict

Adaptive retained and static retained PASSED every frozen criterion. Replacement
FAILED on interaction post Brier harm, not recurring harm in this particular
run. The previous v7 recurring failure remains valid and unchanged.

Confirmation Brier (24 streams/case; stable/null full, other rows post):

| Scenario | Fixed count | Replacement | Adaptive retained | Static retained |
| --- | ---: | ---: | ---: | ---: |
| Stable05 | 0.048373 | 0.048741 | 0.048373 | 0.048368 |
| Stable20 | 0.166188 | 0.168086 | 0.166278 | 0.166248 |
| Shift128 | 0.199944 | 0.176113 | 0.175999 | 0.182653 |
| Shift256 | 0.238102 | 0.218240 | 0.220176 | 0.222008 |
| Shift384 | 0.263635 | 0.263021 | 0.263744 | 0.260787 |
| Gradual | 0.202362 | 0.198002 | 0.196560 | 0.196375 |
| Recurring | 0.193158 | 0.199174 | 0.194221 | 0.192530 |
| Delayed/missing | 0.270514 | 0.265905 | 0.268621 | 0.265744 |
| Interaction | 0.236062 | 0.247308 | 0.237738 | 0.240463 |
| Null | 0.251928 | 0.251701 | 0.251823 | 0.251791 |

Adaptive primary gains: 0.023946 [0.009134,0.038757] and
0.017926 [0.004418,0.031435]. Static primary gains: 0.017291
[0.007210,0.027372] and 0.016095 [0.007131,0.025058]. Intervals follow the frozen
paired approximate z=3.6 procedure, conditional on three fitted incumbents.
All fit-group primary means improve. Both candidates pass both stationary
protections and all scenario/window mean-harm limits.

Stationary05 accuracy remains about 94.99%. Adaptive early-shift accuracy is
75.54% versus fixed 67.85%; midstream 65.74% versus 58.64%. This does not restore
stationary accuracy after arbitrary changes. Late shifts and absent/delayed
feedback remain weak. Interaction ability is preserved better, not solved.

## Interpretation

Retaining the short-count learner is supported as a rescue bundle. Static blend
also passes; adaptive weighting is not established as necessary or uniformly
superior. Adaptive does slightly worse than static on some cases. Keep both
candidates for independent replication rather than choosing from raw averages.

Replacement interaction harm is 0.011245 with approximate interval
[0.003298,0.019193], violating the same .01 mean guard. The failed case differs
from v7, illustrating why one scenario or seed is inadequate.

No evidence here establishes a target-law diameter certificate, real-text
calibration, causal identification, actual agent-task gain or production shadow
latency. Tree missing-field integration still assumes independent fair bits.
Items 1/2/4 are advanced, not globally completed or exhausted.

## Verification

Targeted race tests verify exact v7 control forecasts and views across five
scenario types; replay of inner pre-outcome probabilities and delayed updates;
and budget6. Vet passed. The complete artifact replay verifies source hashes,
availability accounting, views, inner journals, scores and verdicts, excluding
wall-clock timing. No v7 source or evidence was modified.

Next broaden the observation-model/generator assumptions, test explicit
falsification acquisition, and validate on fresh outcome-verifiable agent tasks.
Production shadow validation and Anti-Pigeon gate integration remain separate.
