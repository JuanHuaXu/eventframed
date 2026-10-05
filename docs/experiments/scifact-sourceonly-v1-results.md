# Source-Only Rank Feature Ablation

2026-10-04. Two actual full-FIT studies and independent audits are terminal.
Both previously failed learners improve their mean screen when the two native
nomination/rank feature channels are zeroed. The gain is small and comes from
already-consumed FIT data, not untouched confirmation. All seven whole goals
remain OPEN.

## Controlled Comparison

[Frozen protocol](scifact-sourceonly-v1-protocol.md). Same 5183 verified stored
source records, all 351 FIT queries, all 200 candidates per query, five
whole-source-family folds, optimizer, bounded correction and stronger BM25
control. No target-dependent nomination, corpus shrinking, law/confidence change
or new production code. The mask validates the original full frontier, then
returns an owned copy with only feature channels 6/7 zeroed. Native retrieval
still contributes to nomination; this is not evidence that native retrieval
itself should be removed.

Query-macro citation retrieval:

| Arm | Recall10 | MRR10 | NDCG10 |
| --- | ---: | ---: | ---: |
| Common-frontier BM25 control | 79.81956% | 0.631353 | 0.667185 |
| Original all-pair learner | 79.08357% | 0.627632 | 0.662051 |
| Original Lambda learner | 79.51092% | 0.630939 | 0.666090 |
| Source-only all-pair learner | 80.00950% | 0.635349 | 0.670806 |
| Source-only Lambda learner | 80.00950% | 0.633845 | 0.669643 |

The new recall10 gain over control is 0.18993 percentage points, not a large
relative improvement. The common frontier recall stays 92.92498%. Pure BM25
frontier recall is 92.64008%, with identical control top-ten metrics in this FIT
screen. Citation relevance is not factual truth or agent-answer usefulness.

Equal-family macro (221 families): control recall10 79.78381%, both source-only
learners 79.85923%. Only one family improves recall and one worsens; 219 tie.
All-pair NDCG improves in 20 families and worsens in 10; Lambda improves in 18
and worsens in 10. These are mean screening results, not statistical adoption
proof or evidence of robustness to new tasks/regimes. No calibration outcome
was used to choose the primary model; the cheaper all-pair candidate was frozen
as primary, Lambda as secondary, before the separate native calibration trial.

## Correctness and Leakage Audit

Authoritative outputs:

- `research/public-task-pilot/scifact-sourceonly-pairrank-v1/audit-results.json`
- `research/public-task-pilot/scifact-sourceonly-lambdarank-v1/audit-results.json`

Both independent Node audits rebind all actual stored ID/text/provenance/clock
records, reconstruct all 70200 masked features, independently refit all six
models, and reconstruct all 351 complete held-fold rankings. Training excludes
each held source-family fold; all-FIT models use exactly the original 351 FIT
IDs. Five ranking corruption controls reject in each study. FIT inputs and
target projections are byte-identical to the corresponding original study.

Complete local test closures contain 15/20 sources. Each study records five
terminal code-zero core commands (race, vet, build, experiment, benchmark).
All eight/ten race roots actually execute three times, including owned-copy,
native-cue invariance and invalid masked-value controls. Calibration/confirmation
prediction counts are zero in these two studies. The new calibration trial has
its own separate consumption ledger and cannot inherit that zero count.

Generation caveat: the shared generator's receipt precedes a small pre-launch
repair to the new Lambda runner's inherited generator-provenance paths. The
actual final runner is frozen in its study; the receipt is not a hash of that
final runner. No scientific input, model, optimizer or prediction changed in
that path repair. Original negative studies remain immutable and failed.

## Measured Cost

| Measurement | All-Pair | Lambda |
| --- | ---: | ---: |
| Source feature-index construction | 97.74 ms | 98.33 ms |
| Full-FIT training | 496.43 ms | 863.04 ms |
| 200 features, median / p99 | 0.378 / 0.678 ms | 0.369 / 0.723 ms |
| Two complete 200-rank arms, median / p99 | 35.25 / 53.96 us | 34.13 / 50.00 us |
| Whole offline executable wall time | 3.08 s | 4.89 s |
| Executable maximum RSS | 189.58 MiB | 177.39 MiB |

The standalone single-rank benchmark is 22.69-28.27 us, 18216 B and 15
allocations/op. These cached/offline measurements exclude native query latency,
live cache invalidation, concurrent load and online learner publication. They
do not establish sub-100ms serving or goal-6 freshness. Full-fit training is
reported once as model-origin cost in the frozen-model calibration trial.

## Next Requirement

Test the frozen primary on all previously unused calibration queries using a
fresh owned native store, actual full-frontier source hydration and independent
ranking reconstruction. Preserve all losses and bootstrap uncertainty; do not
refit or choose a different winner after seeing those outcomes. Confirmation
300 stays closed until its predeclared gate is met. Other whole-goal work still
needs delayed-evidence alignment, useful downstream split effects and loaded
background-learning freshness/admission experiments.
