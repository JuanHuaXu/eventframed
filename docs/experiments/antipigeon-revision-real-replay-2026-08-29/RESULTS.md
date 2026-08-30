# Post-Revision Real-Session Replay

Date: 2026-08-29

## Scope

This retrospective replay used the sanitized local Codex-session corpus, the
frozen `2026-08-28T00:00:00Z` data boundary, a fresh isolated LibraVDB
`1.9.36-beta.5` database, its real `SearchTextCollections` and `RankCandidates`
contracts, GGUF `nomic-embed-text-v1.5` embeddings, and the previously frozen
baseline and predictive calibration maps. Production OpenClaw memory was not
read or modified.

The replay contains 1,286 design cases in nine trajectories and 138
confirmation cases in three trajectories. It is a regression test of the
current retrieval, posterior, elastic-reranking, and packing pipeline. It does
not directly test the new Anti-Pigeon split transition because the real replay
does not publish sharing certificates; the synthetic grouped-outcome experiment
remains the direct test of that transition.

## Results

| Block and metric | Baseline | EventFrame | Full upgrade | Full-upgrade change vs baseline |
| --- | ---: | ---: | ---: | ---: |
| Design Recall@10 | 0.40308 | 0.46244 | 0.45685 | +0.05377 |
| Design packed recall | 0.19777 | 0.24047 | 0.24048 | +0.04272 |
| Design high-priority miss rate | 0.04766 | 0.03512 | 0.03846 | -0.00920 |
| Confirmation Recall@10 | 0.34617 | 0.39809 | 0.40643 | +0.06027 |
| Confirmation packed recall | 0.19690 | 0.23370 | 0.23961 | +0.04271 |
| Confirmation Recall@50 | 0.88276 | 0.90721 | 0.90793 | +0.02517 |
| Confirmation MRR | 0.81005 | 0.90459 | 0.87977 | +0.06972 |
| Confirmation high-priority miss rate | 0.02381 | 0.02381 | 0.02381 | 0.00000 |
| Confirmation Brier | 0.55166 | 0.55617 | 0.55466 | -0.00300 |

For the full upgrade, the confirmation Recall@10 gain has a trajectory-cluster
95% interval of `[+0.03235, +0.27473]`. The packed-recall gain interval is
`[-0.00278, +0.27473]`, so the positive point estimate is not established across
trajectories. The Brier gain interval is `[-0.00461, -0.00181]`: the frozen
calibration map was measurably worse, not non-inferior, on this replay.

Against shuffled feedback rather than the no-update baseline, the full upgrade
improved confirmation Recall@10 by `+0.08465`, packed recall by `+0.06787`, and
Brier by `+0.00519`; all three cluster intervals are positive. This supports the
claim that the feedback signal is informative, while showing that its current
probability calibration is not yet correctly scaled.

## Interpretation

- **Validated in this replay:** useful feedback improves Recall@10 and does not
  increase the confirmation high-priority miss rate.
- **Inconclusive:** packed-recall improvement is positive but its three-trajectory
  interval crosses zero.
- **Falsified for the frozen maps:** proper-score/calibration non-inferiority.
  Recalibration must use a chronological design block and then be checked once on
  untouched confirmation data.
- **Bottleneck:** confirmation nomination recall is only `0.45361`. EventFrame
  cannot rerank relevant events that LibraVDB removes before the frontier, so
  sidecar gating and frontier coverage need a separately frozen contract test.
- **Not exercised:** Anti-Pigeon `split` and `split_reset`, because no real sharing
  certificates were installed.

The earlier elastic replay is not numerically interchangeable with this one:
the current runner includes the later ablation family, and the earlier protocol
did not freeze the sidecar gating configuration needed to explain its reported
nomination recall of `1.0`. This run should therefore be treated as current
regression evidence, not a paired before/after estimate of the revision gate.

LibraVDB again emitted deferred micro-temporal `dirty-anchor generation is stale`
warnings. Foreground insertion, nomination, ranking, artifact generation, and
graceful shutdown completed successfully.

## Calibration rescue

After the frozen-map failure above, a single monotone logit rescue was fitted on
the 1,286-case design block only. The calibration fitter was first corrected to
use accepted Newton steps and to exclude evaluator-only zero placeholders for
events that were never nominated. The 25,505 nominated design forecasts improved
from Brier `0.33694` to `0.18977`.

The fitted deployment map is `scale=4.645131359113079`,
`bias=-4.221099093884319`, and `floor=0.000001`. Applying that frozen map once to
the untouched 138-case confirmation block reduced full-upgrade Brier from
`0.55466` to `0.39862` and priority-weighted Brier from `0.56027` to `0.40085`.
The Brier gain over baseline is `+0.15303`, with trajectory-cluster 95% interval
`[+0.09994, +0.18196]`.

Because calibration is downstream of rank utility, Recall@10 (`0.40643`),
Recall@50 (`0.90793`), MRR (`0.87977`), packed recall (`0.23961`), and the
high-priority miss rate (`0.02381`) were unchanged. This rescues calibration
non-inferiority for this frozen replay. It does not improve the separate
nomination-recall bottleneck and does not constitute a real-data test of
Anti-Pigeon splitting. The confirmation block is now spent for this calibration
family; no additional map should be selected against it.

The machine-readable fit, input hashes, deployment map, and confirmation metrics
are in `calibration-rescue.json`.
