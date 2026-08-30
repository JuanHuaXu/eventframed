# Semantic Dual-Calibration Design Results

## Status

This is an exploratory design-block result, not confirmation evidence. The
semantic model, baseline calibrator, and belief-conditioned calibrator were all
selected or fitted using the same design block. The previously consumed
confirmation sessions were not scored again.

## Configuration

- 1,286 retrospective downstream-use cases from nine independent Codex session
  clusters.
- Local `nomic-embed-text` embeddings through Ollama, with the model-declared
  `search_document: ` and `search_query: ` prefixes.
- Baseline logit calibration: scale `2.7372846009`, bias `-5.6200788843`.
- Belief-conditioned logit calibration: scale `4.8446965465`, bias
  `-8.1000760964`.
- Both probability floors: `0.000001`.
- The bounded frontier is exhaustive within each replay segment. Raw transcript
  text, production timestamps, paths, and tool payloads are not exported.

## Results

| Metric | Calibrated semantic baseline | EventFrame | Design result |
|---|---:|---:|---|
| Brier loss | `0.20179` | `0.15423` | Improved by `0.04755`; cluster-bootstrap 95% interval `[0.04489, 0.05858]` |
| Expected calibration error | `0.01544` | `0.04414` | EventFrame is less calibrated than the separately fitted baseline, but much better than its uncalibrated `0.50401` |
| Recall@10 | `0.41597` | `0.58798` | Improved by `0.17201`; cluster-bootstrap 95% interval `[0.07355, 0.18666]` |
| Recall@50 | `0.90824` | `0.97768` | Improved descriptively |
| Mean reciprocal rank | `0.73506` | `0.88488` | Improved descriptively |
| High-priority miss rate@10 | `0.04933` | `0.01171` | Lower by `0.03763` |

The shuffled-feedback control had Brier `0.20208`, Recall@10 `0.40042`, and MRR
`0.73435`. It did not reproduce the outcome-aligned update gains. EventFrame and
the update-all variant were effectively identical, so the residual layer again
added no material benefit in this replay.

## Interpretation

Semantic embeddings substantially improve candidate ordering, and separate
calibration contracts let the baseline-only and belief-conditioned laws retain
appropriate probability scales. Compared with the uncalibrated semantic run,
EventFrame Brier fell from `0.42951` to `0.15423` and ECE from `0.50401` to
`0.04414`. Recall@10 moved from `0.59539` to `0.58798` while MRR rose from
`0.87444` to `0.88488`; this reflects transitions between the baseline-only and
certified-belief branches, not a within-branch reversal by either monotone map.

These figures show design fit, not expected production performance. A new
chronological holdout or prospective session stream must freeze both calibrators
before scoring. The old confirmation block must remain retired.
