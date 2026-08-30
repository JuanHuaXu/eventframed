# Elastic Rank-Delta Real-Session Replay

Date: 2026-08-29

## Scope

This retrospective test replays the existing sanitized Codex-session corpus
through EventFrame's elastic rank-delta policy. It uses read-only local session
JSONL, the frozen `2026-08-28T00:00:00Z` data boundary, the real LibraVDB
`1.9.36-beta.5` nomination and ranking contracts, and GGUF
`nomic-embed-text-v1.5` embeddings. The LibraVDB socket and database were
isolated from production OpenClaw state.

The comparator is the previously recorded fixed-delta replay over the same case
and trajectory IDs. This is not a new untouched confirmation block: its value is
paired implementation evidence for the elastic policy. Raw text, production
timestamps, file paths, and tool payloads are not exported.

## Results

| Metric | Fixed delta | Elastic delta | Paired change | Trajectory-cluster 95% interval |
| --- | ---: | ---: | ---: | ---: |
| Design Recall@10, 1,286 cases / 9 sessions | 0.51758 | 0.56977 | +0.05218 | [+0.01571, +0.05907] |
| Design packed recall | 0.37127 | 0.40968 | +0.03842 | [+0.00564, +0.04520] |
| Confirmation Recall@10, 138 cases / 3 sessions | 0.43285 | 0.45187 | +0.01903 | [+0.01359, +0.06593] |
| Confirmation packed recall | 0.34525 | 0.35628 | +0.01102 | [+0.00349, +0.05934] |
| Confirmation Recall@50 | 0.95840 | 0.96841 | +0.01001 | Descriptive |
| Confirmation MRR | 0.91207 | 0.91441 | +0.00234 | Descriptive |
| Confirmation high-priority miss rate | 0.00794 | 0.00794 | 0.00000 | No increase |
| Confirmation Brier | 0.23237 | 0.23237 | 0.00000 | Probability law unchanged |

On confirmation, 35 cases improved Recall@10, three regressed, and 100 tied.
Packed recall improved in 27 cases, regressed in ten, and tied in 101. All three
confirmation sessions had positive mean Recall@10 and packed-recall changes.

## Attribution Checks

- All 138 case IDs and all three trajectory IDs match the fixed-delta artifact.
- All baseline metrics match; the only extra baseline field is the newly
  reported high-priority case count.
- Across 6,336 paired confirmation candidates, zero forecast probabilities
  changed and 6,001 rank scores changed.
- Brier, expected calibration error, and priority-weighted Brier are identical.

These checks are consistent with the intended mechanism: elasticity changes how
strongly a certified correction affects order and packing, not the committed
forecast law.

## Runtime Note

LibraVDB emitted repeated deferred micro-temporal `dirty-anchor generation is
stale` maintenance warnings during the replay. Foreground indexing and ranking
completed, all artifacts were produced, and the daemon exited cleanly. Those
warnings are an upstream runtime limitation and are not counted as EventFrame
accuracy evidence.

## Interpretation

The elastic policy improved real-session retrieval and packing on this
retrospective corpus without worsening probability calibration or high-priority
misses. This supports keeping the feature enabled for the next prospective test.
It does not establish population-level production gain because the confirmation
sessions were previously inspected and the usefulness labels remain a
downstream-reuse proxy rather than direct task-success judgments.
