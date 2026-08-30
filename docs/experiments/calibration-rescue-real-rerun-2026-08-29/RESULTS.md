# Calibration Rescue Real-Data Rerun

Date: 2026-08-29

## Scope

This is a fresh end-to-end reproducibility replay over the same read-only Codex
session cutoff (`2026-08-28T00:00:00Z`). It used a new isolated LibraVDB
`1.9.36-beta.5` database, GGUF `nomic-embed-text-v1.5`, the real
`SearchTextCollections` and `RankCandidates` contracts, the current EventFrame
runtime, and the rescued predictive map (`scale=4.645131359113079`,
`bias=-4.221099093884319`, `floor=0.000001`). Production OpenClaw memory was not
read or modified.

The replay contains the same 1,286 design cases in nine trajectories and 138
confirmation cases in three trajectories. Because the confirmation block had
already been inspected during post-hoc calibration validation, this run is a
regression/reproducibility test rather than independent confirmation.

## Results

| Block and metric | Baseline | Full upgrade | Change |
| --- | ---: | ---: | ---: |
| Design Brier | 0.20568 | 0.47441 | -0.26872 |
| Design Recall@10 | 0.40113 | 0.47690 | +0.07577 |
| Design packed recall | 0.27868 | 0.33533 | +0.05665 |
| Confirmation Brier | 0.32094 | 0.28817 | +0.03277 |
| Confirmation priority-weighted Brier | 0.32191 | 0.28450 | +0.03741 |
| Confirmation Recall@10 | 0.39439 | 0.42851 | +0.03412 |
| Confirmation packed recall | 0.28564 | 0.32557 | +0.03993 |
| Confirmation Recall@50 | 0.92749 | 0.95073 | +0.02324 |
| Confirmation MRR | 0.84534 | 0.89884 | +0.05350 |
| Confirmation high-priority miss rate | 0.00794 | 0.00794 | 0.00000 |

The confirmation Recall@10 gain has trajectory-cluster 95% interval
`[+0.02449, +0.12161]`; packed-recall gain has interval
`[+0.00381, +0.13260]`. Both retrieval improvements are positive across this
three-trajectory inference. The confirmation Brier gain interval is
`[-0.16806, +0.11508]`, so its favorable point estimate is inconclusive.

On design, the full-upgrade Brier loss is established: its gain interval is
`[-0.29114, -0.11906]`. The rescued map therefore does not support a stationary
calibration claim across this replay. Ranking and probability quality must remain
separate claims.

## Retrieval-regime difference

This run reports nomination recall `1.0` and activation approximately `1.0`,
whereas the earlier Anti-Pigeon revision replay reported confirmation nomination
recall `0.45361`. The contract names and corpus cutoff match, but the earlier run
did not freeze enough LibraVDB gating configuration to make the frontier regimes
equivalent. The prior post-hoc calibration result and this fresh run are therefore
not interchangeable.

Calibration artifacts must bind a complete nomination/gating fingerprint, not
only a contract name. A map fitted under one frontier regime must fail closed or
return to shadow mode when that fingerprint changes. A prospective untouched
block is still required before probability-driven agency decisions can be
enabled.

LibraVDB again emitted deferred micro-temporal `dirty-anchor generation is stale`
warnings. Foreground indexing, nomination, ranking, report generation, and clean
shutdown completed successfully.
