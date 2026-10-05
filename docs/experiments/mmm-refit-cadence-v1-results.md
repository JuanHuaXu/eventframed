# Faster refitting: useful diagnostic, incomplete rescue

Completed all2688 consumed trajectories. Fitting every8 rather than32 frames
passes168/168 served-mixture non-harm screens and2/16 required delayed terminal
gain screens. The overall screen FAILS. Four times as many fit bundles are
required, so this is not an equal-compute rescue or production recommendation.
All seven full research goals remain open.

Phase1 terminal64 served Brier (lower is better):

| Case | Immediate32 | Immediate8 | Delayed32 | Delayed8 |
| --- | ---: | ---: | ---: | ---: |
| Additive stationary | .220269 | .218795 | .221755 | .221658 |
| Additive gradual | .225874 | .223369 | .239065 | .234161 |
| Parity4 | .048973 | .049191 | .049066 | .050053 |
| Null | .255557 | .257136 | .258525 | .257049 |
| Majority to parity | .049675 | .050200 | .060262 | .053252 |
| Parity to majority | .077009 | .068210 | .102272 | .098235 |

The two delayed terminal gain passes are phase0 additive gradual
(.007633[.001284,.013983]) and phase1 additive abrupt
(.006509[.000609,.012409]). They are not the same case replicated across phases.
For phase1 parity-to-majority, complete-feedback gain.008799 has a positive
lower bound. Delayed gain.004038[-.009863,.017939] remains uncertain and below
the0.005 mean target. Its delayed-minus-immediate gain interaction is
-.004761[-.020258,.010736], not evidence that this specifically rescues delay.

## Where the benefit occurs

Across840 case/schedule/window/component cells,182 gain intervals have positive
lower bounds and none has upper bound below zero. For the served mixture alone,
48/168 have positive lower bounds:45 full-trajectory and only3 terminal cells.
These are exploratory per-cell intervals, not simultaneous guarantees; lack of
a negative interval is not universal non-harm. Some point estimates worsen.

Phase1 delayed majority-to-parity full Brier improves.153678 to.132898;
parity-to-majority improves.157667 to.143396. The corresponding first192-frame
gains, derived from the paired full and terminal sums, are.025370
[.018907,.031834] and.017682[.012820,.022543]. Faster refitting helps important
portions of learning/recovery, but does not close the terminal requirements.

A post-result analytic floor check finds every one of the16 terminal targets
has at least0.005 baseline headroom above mean Q(1-Q). The smallest is.012262.
Thus the required mean gain is not ruled out by the Bayes Brier floor. This
does NOT prove that a finite-data admissible learner can reach it; Q is used
only for evaluation. No gate was weakened or removed after this check.

## Verification and cost

The independent audit checks6,881,280 original/new probabilities,86016 fast
fit bundles, complete ordered fit-origin lists, and the served Markov mixture
by direct-probability recursion. It verifies1344 matched Initial/X/Y/Q trajectory
signatures before computing cadence/feedback interactions. Three mutated full
records and12 handcrafted mutations are rejected. Underlying expert fitting is
not independently reimplemented. Scoring replay is byte-identical (`cmp` exit0).

Race contracts passed on four paired fixtures and eight poisoned prefixes.
Collection took338.13s wall,1351.16s user CPU,4.95s system CPU with four workers.
Fit bundles increase from21504 to86016 (four experts per bundle). This is an
offline experiment runtime, not hot-path latency or a measured4x CPU ratio.
No paid labels are acquired, and default cadence stays32.

Artifacts: [protocol](mmm-refit-cadence-protocol.md),
[raw](mmm-refit-cadence-v1.jsonl), [summary](mmm-refit-cadence-v1-summary.json),
[scoring replay](mmm-refit-cadence-v1-summary-replay.json),
[floor/window diagnostic](mmm-refit-cadence-v1-headroom.json),
[race checks](mmm-refit-cadence-contract-results.md), [run](mmm-refit-cadence-v1-run.txt).

## Next lead

Faster publication helps, but is neither sufficient nor cost-free. Before
another combined learner change, decompose remaining fast-cadence error into
limits of the issued forecast bank versus selection error. A hindsight best
expert/convex-hull diagnostic is not a deployable rule; it can identify whether
another selector has enough headroom or whether different predictions are needed.
Preserve the earlier32-cadence oracle work rather than treating it as evidence
for this new8-cadence bank. Do not tune acquisition thresholds to these results.
