# Prospective Cold-Start Pilot Results

## Status

This new temporal holdout supports continued testing, but it is a pilot rather
than confirmatory evidence. It contains 36 scoreable cases from one independent
Codex task, so cross-task uncertainty cannot be estimated. The report marks
`cluster_inference_valid` false; its degenerate single-cluster bootstrap values
must not be interpreted as confidence intervals.

## Frozen protocol

- Data window: `2026-08-28T00:00:00Z` through the exclusive boundary
  `2026-08-29T12:32:56Z`.
- Cold start: records before the lower boundary were neither ingested nor used
  for posterior feedback.
- Model: local `nomic-embed-text` with frozen document/query prefixes.
- Baseline calibration: scale `2.7372846009`, bias `-5.6200788843`.
- Belief-conditioned calibration: scale `4.8446965465`, bias
  `-8.1000760964`.
- No model, threshold, weight, or calibrator was fitted to these 36 outcomes.
- Raw transcript text, paths, tool payloads, and production timestamps are not
  exported.

## Results

| Metric | Frozen baseline | EventFrame | Descriptive change |
|---|---:|---:|---:|
| Brier loss | `0.26846` | `0.23404` | `0.03442` lower |
| Expected calibration error | `0.10989` | `0.08530` | `0.02459` lower |
| Recall@10 | `0.43754` | `0.49818` | `0.06064` higher |
| Recall@50 | `0.91582` | `0.95777` | `0.04194` higher |
| Mean reciprocal rank | `0.77931` | `0.88889` | `0.10958` higher |
| High-priority miss rate@10 | `0.03333` | `0.03333` | unchanged |

The shuffled-feedback control was worse than baseline: Brier `0.27865`,
Recall@10 `0.38511`, and MRR `0.70654`. EventFrame and update-all were identical,
so this pilot again provides no evidence that residual correction adds value in
the downstream-use proxy.

## Interpretation

The direction of every principal retrieval and probability metric agrees with
the earlier retrospective experiment, and outcome-aligned updates outperform
randomly assigned updates. That is useful out-of-window evidence. It does not,
however, establish that the effect generalizes beyond this task or user because
all cases share one trajectory.

The next confirmation run should keep this protocol frozen until at least three
new qualifying tasks are available. More cases from this same task improve
within-task precision but do not solve the missing independent-cluster problem.
