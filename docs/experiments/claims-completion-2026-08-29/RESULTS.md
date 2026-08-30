# Remaining claims completion results

Run date: 2026-08-29. The acceptance criteria were frozen in
`PREDECLARED_PROTOCOL.md` before these runs. Unfavorable outcomes were retained.

## Decision matrix

| Claim | Result | Evidence |
| --- | --- | --- |
| Structured frames | **Inconclusive** | The chronological block produced 46 cases from only 2 trajectories, below the frozen minimum of 3. No independent interpretability ratings were collected. |
| Residual reuse | **Falsified** | Heterogeneous stress Brier improved by 0.01931, 95% trajectory bootstrap CI [0.00614, 0.03280], but 635/1536 applied corrections were harmful. The false-reuse rate was 41.34%, Wilson 95% CI [38.90%, 43.82%], far above the 5% ceiling. |
| Selective Bayesian update | **Inconclusive on the primary real track; contradicted by stress evidence** | The real block had only 2 trajectories. In the synthetic stress replay, 5% activation increased Brier by 0.02891 versus update-all, exceeding the 0.01 ceiling; priority-weighted increase was 0.03137, trajectory bootstrap 95% CI [0.03130, 0.03144]. |
| Anti-Pigeon granularity | **Falsified** | Moderate and strong divergent groups split correctly in 90.63% and 100% of confirmation trajectories, but the compatible 0.8/0.8 group was uncertain in all 64 trajectories and never reached the required share decision. |
| Omitted-influence coverage | **Inconclusive; estimator not implemented** | The daemon validates externally supplied certificate fields but contains no shadow-audit estimator for divergence coverage or never-nominated events. Existing experiment certificates assert synthetic values and therefore cannot establish empirical 95% coverage. |
| Changepoint adaptation | **Validated for seeded synthetic mechanisms** | All six frozen v5 scenario criteria passed. Noiseless changes were detected perfectly; noisy abrupt, gradual, and recurring detection rates were 87.50%, 95.31%, and 79.69%. This does not establish real-stream prevalence or production thresholds. |
| Predictive snapping | **Falsified as an outcome-improvement claim** | Publishing and rollback correctly advance graph/dependency versions, but an integration test confirms that accepted snaps do not change recall scores, rank deltas, templates, or forecast laws. A positive confirmation gain is therefore impossible in the current runtime. |
| Priority weighting | **Inconclusive** | The chronological block had only 2 trajectories. Its high-priority miss rate was 7.5% for both baseline and EventFrame, so the point estimate shows no improvement, but the frozen inferential test cannot be run. |

## Artifacts

- `stress-design.json` and `stress-confirmation.json`: independent seeded
  residual, Anti-Pigeon, Bayesian group, and changepoint experiments.
- `selective-stress-dataset.json`, `selective-stress-report.json`, and
  `selective-vs-update-all-report.json`: update-all, selective, EventFrame, and
  no-update controls.
- `codex-cold-start/`: chronological post-cutoff Codex replay and ablations.
- `internal/service/graph_integration_test.go`: operational snapping and exact
  rollback invariance check.

## Interpretation

The completed sweep does not support publishing the full implementation claim
set as validated. Changepoint adaptation passes its synthetic contract, and the
Anti-Pigeon split branch works on known divergent groups. The compatible-share
branch, stale-residual suppression, certified selective-update quality,
omitted-influence estimator, and predictive use of graph snaps need engineering
work before another frozen confirmation round.

The real-session block remains useful exploratory evidence, but two trajectories
cannot support the predeclared clustered inference. Collecting more cases from
the same two sessions would not repair that independence problem.
