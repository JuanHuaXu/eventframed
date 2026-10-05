# v94: immediate-feedback Brier aggregation

## Verdict

Component checks PASS. This is not a passed learned-model rescue, a delayed
feedback guarantee, or completion of any whole research direction.
No serving or production path was changed.

The frozen [protocol](mmm-brier-kernel-v94-protocol.md),
[raw records](mmm-brier-kernel-v94.json), and
[independent summary](mmm-brier-kernel-v94-summary.json) contain 448 records:
224 paired streams, two variants, seven cases, and 4,096 predictions per record.
Five deterministic cases each repeat the same tape across seed indices; these
duplicates are stress checks, not independent confirmation observations.

Every recorded prefix respects its declared generic-expert cumulative Brier
bound within 1e-8. The maximum numerical Jensen defect is 2.22e-16.
The no-sharing bound is 0.102587 cumulative loss; sharing at 0.001 additionally
charges -2(N-1)log(0.999). These are realized sequential-loss bounds, not
pointwise safety, future-risk guarantees, or posterior probabilities of truth.

## Recovery and cost

In the deterministic halfway reversal, plain exponential weighting makes no
correct post-flip decisions within the remaining 2,048 steps. It retains its
cumulative guarantee while adapting too slowly. Fixed sharing first gives the
challenger majority weight after 18 post-flip outcomes, achieves 99.1211%
post-flip accuracy, and reduces whole-stream Brier from 0.499963 to 0.003795.
This is a diagnostic stress construction, not real-agent accuracy.

Sharing has a measurable protection cost: when the generic expert is better,
mean Brier is 0.040058 versus 0.040012 without sharing and 0.040000 for generic.
Under the adaptive adversary it is 0.166911 versus 0.166855 without sharing.
Do not describe the variant as pointwise non-harmful.

On Apple M4, Go 1.27.1 darwin/arm64, three 500ms benchmark repetitions measured
38.10-38.69 ns per predict/update cycle without sharing and 114.1-114.2 ns with
sharing, both zero allocations. This is a two-expert, single-owner arithmetic
component, not daemon latency; model fitting, retrieval, I/O, delayed queues,
and concurrent ownership are excluded. See [raw timing](mmm-brier-kernel-v94-benchmarks.txt).

## Verification and limits

Race-enabled lifecycle tests and full exact artifact replay PASS (3.029s).
Generation PASS (0.422s package time), vet PASS, source hashes and independent
summary checks PASS. Tests cover invalid inputs without mutation, ordered
feedback, duplicate rejection, snapshot isolation, counter overflow, and
recovery from displayed-weight underflow. The state machine rejects issuing a
second forecast before feedback rather than silently omitting the first label.
A missing-feedback counterexample exceeds the no-sharing bound.

Next freeze a prospective learned-model experiment with changing generic and
Boolean experts, ordered outcomes, and an untouched confirmation split. Test
expected Brier and recovery separately from cumulative realized loss. Preserve
the v92/v93 failed protection evidence. Delayed, missing, and censored feedback
need a separately justified algorithm and cannot inherit this result.

The conservative convex-mixture derivation is distinct from the optimal strong
aggregating algorithm in [Vovk and Zhdanov (2009)](https://www.jmlr.org/papers/v10/vovk09a.html).
Delay handling remains separate, as emphasized by
[Joulani et al. (2013)](https://proceedings.mlr.press/v28/joulani13.html).
