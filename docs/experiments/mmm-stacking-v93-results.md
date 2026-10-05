# Exact LOO stacking v93 results

## Verdict

**FAIL the regularized-stacking advance gate.** All12 targeted excess-risk
conditions pass, but only56/60 case/phase/sample-count cells pass non-harm.
Unregularized stacking passes12/12 and54/60. The skeptical-BMA comparison
passes12/12 and60/60 on these fresh data; it was not a selectable alternative
winner for v93, and its earlier v92 failure remains part of the evidence.

The correct conclusion is mixed protection evidence for skeptical BMA and a
failed stacking upgrade, not a retroactive pass of either earlier experiment.
No delayed-stream rescue, production adoption or whole-direction completion.

There are1,280 independent fresh training sets and3,840 nested-prefix records,
each with four paired arms. Expected scores integrate all512 raw inputs and
fresh simulator outcome probabilities. These are finite component results,
not observed agent accuracy, grokking or adaptive MMM acquisition performance.

## Regularized confirmation results

| Rule | Labels | Generic Brier | Stacking Brier | Generic expected accuracy | Stacking expected accuracy |
| --- | ---: | ---: | ---: | ---: | ---: |
| Parity4 | 16 | 0.26213 | 0.15687 | 51.12% | 77.91% |
| Parity4 | 32 | 0.15861 | 0.05782 | 76.74% | 95.00% |
| Parity4 | 64 | 0.07371 | 0.05289 | 92.39% | 95.00% |
| Majority3 | 16 | 0.19042 | 0.19992 | 73.65% | 72.18% |
| Majority3 | 32 | 0.10332 | 0.10387 | 88.96% | 88.80% |
| Majority3 | 64 | 0.05635 | 0.05635 | 94.87% | 94.87% |
| Multiplexer3 | 16 | 0.19133 | 0.19662 | 73.55% | 72.73% |
| Multiplexer3 | 64 | 0.05682 | 0.05699 | 94.70% | 94.62% |

At32 labels, parity4 gain0.10079 has approximate mean +/-3.5SE interval
[0.08288,0.11870]. However, confirmation n16 majority mean harm0.00951 has
upper bound0.02491 and worst observed fit harm0.19154. Confirmation n16
multiplexer mean harm0.00529 has upper bound0.01710 and worst fit harm0.19507.
These are failed protection conditions, not arithmetic failures of the optimizer.

The four failed regularized cells are design n16 multiplexer, confirmation n16
majority, confirmation n16 multiplexer, and confirmation n16 null. Majority also
fails the partial-view bound. Full-input null's mean harm0.00252 has upper
bound0.01149; expected null accuracy remains50%. These normal bounds are not
confidence sequences or per-event safety guarantees.

## Consumed-tail diagnostic

The original v92 majority index37 record and its source hashes were reproduced
before scoring new methods. This is consumed evidence, explicitly not fresh
confirmation. Generic Brier is0.23695 and skeptical BMA0.40899. Unregularized
stacking gives weight1 and Brier0.45988; lambda1 gives weight0.81282 and
Brier0.40119. Both have50% expected accuracy. Thus leave-one-out exclusion
does not supply new independent information to resolve this misleading sample.
Do not remove the witness or add a mask-specific exception.

LOO validity and statistical sufficiency are different. All held-out predictions
match explicit refits, and flipping an excluded label does not change its LOO
prediction. Nevertheless, the small-window validation signal is misleading.
The primary stacking paper already warns about small-sample weight instability;
this experiment does not refute its general results.

## Performance: shortcut works, upgrade still costs more

Apple M4, darwin/arm64, Go1.27.1, benchmark suffix10;500ms and three repeats,
same fixtures, no concurrent experiment:

| Component | Time range | Allocation |
| --- | ---: | ---: |
| LOO shortcut,16 labels | 7.320-7.340ms | ~677,376 bytes,6 allocations |
| Explicit LOO refits,16 | 102.053-102.199ms | 5,572,320 bytes,77 allocations |
| LOO shortcut,64 | 11.276-11.291ms | 678,144 bytes,6 allocations |
| Explicit LOO refits,64 | 477.147-478.475ms | 22,310,880 bytes,317 allocations |
| Skeptical full fit,64 | 7.519-7.526ms | 667,648 bytes,4 allocations |
| Stacking full fit,64 | 11.357-11.382ms | 997,632 bytes,7 allocations |
| Stacking full fit,256 | 26.886-26.996ms | 1,000,704 bytes,7 allocations |
| Compiled stacking lookup | 6.713-6.889ns | 0 bytes,0 allocations |

The exact shortcut is about42 times faster than explicit64-fold refits. It also
retains full-data component predictions and coefficient statistics, while the
explicit baseline returns only LOO predictions. Full stacking still costs
about51% more than skeptical BMA at64 labels, without better non-harm results.
The probability lookup benchmark excludes observation, queueing and serving.

After the two bounded fits, exact LOO adds O(n*d*2^d) work for cell indexing and
subset evidence/prediction averaging, rather than n complete O(4^d) predictive
compilations. The final3^d conditional snapshot remains. These are finite d=9
research bounds, not a high-dimensional complexity claim. No production benefit
is inferred merely from the exact shortcut's speedup.

## Verification and next lead

Exact-LOO/race contracts PASS2.309 seconds; generator/race PASS1.678 seconds.
Generation PASS89.708 seconds. Consumed-tail reproduction PASS0.239 seconds.
Vet and diff checks pass. The independent evaluator checks20 source/protocol
hashes, all records, all three candidate variants and worst per-fit harms.
Full3,840-record replay PASS89.602 seconds.
v90-v93 independent summaries reproduce, and the consumed-tail source hashes
remain intact. No predecessor source was changed to obtain these results.

Next: [online Brier aggregation with an explicit cumulative-loss budget](../../research/online-brier-aggregation-proposal.md).
This is a different prospective protocol, not a static-weight rescue. The
proposal derives a conservative immediate-feedback bound and explicitly excludes
delayed/skipped feedback, resets and unmatched acquisition comparators until
separately justified. No claim that the existing journal already implements it.

Artifacts: [protocol](mmm-stacking-v93-protocol.md), [records](mmm-stacking-v93.json),
[summary](mmm-stacking-v93-summary.json), [tail diagnostic](mmm-stacking-v93-consumed-tail.json),
[benchmarks](mmm-stacking-v93-benchmarks.txt),
[benchmark metadata](mmm-stacking-v93-benchmark-metadata.json).
Raw SHA256: `df58cd5044406e1ccfd3135e10f8a54565243be9eeff48b6623077ce4f0a034b`.
All seven directions remain open. No production/OpenClaw, commit or push changes.
