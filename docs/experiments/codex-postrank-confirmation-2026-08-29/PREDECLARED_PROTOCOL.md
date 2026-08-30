# Corrected Rank-Delta Confirmation Protocol

Frozen before scoring the confirmation block on 2026-08-29.

## Data Boundary

- Source: read-only local Codex session JSONL.
- Exclusive cutoff: `2026-08-28T00:00:00Z`.
- Split: the existing chronological session-level 80/20 split after eligibility
  filtering and before outcome scoring.
- Confirmation block: the three sessions reserved by the uncapped design run.
- The earlier design-only commands emitted zero confirmation cases and did not
  score this block.

## Frozen Runtime

- LibraVDB 1.9.36-beta.5 public `SearchTextCollections` and
  `RankCandidates` contracts.
- `nomic-embed-text`, 768 dimensions, frozen document/query prefixes.
- Baseline calibration: scale `2.7372846009279144`, bias
  `-5.620078884327766`, floor `0.000001`.
- Belief-conditioned calibration: scale `4.844696546547345`, bias
  `-8.100076096425193`, floor `0.000001`.
- Corrected order: LibraVDB base ranking, then bounded EventFrame delta, then
  final sorting and packing.
- Primary candidate: current EventFrame. Baseline and shuffled-feedback variants
  are controls. Other ablations are secondary and cannot replace a failed
  primary result.

## Primary Acceptance Criteria

Current EventFrame is confirmed on this block only if all conditions hold:

1. At least three independent confirmation trajectories are accepted.
2. Brier gain over the LibraVDB baseline is positive and its trajectory-cluster
   bootstrap 95% lower bound is above zero.
3. Recall@10 gain is positive and its trajectory-cluster bootstrap 95% lower
   bound is above zero.
4. Packed-recall gain is positive and its trajectory-cluster bootstrap 95%
   lower bound is above zero.
5. High-priority miss rate at 10 does not increase by more than one percentage
   point over baseline.

Failure of any condition is reported as failed or inconclusive, without changing
thresholds after inspection. Recall@50 and MRR are descriptive secondary
metrics. The shuffled-feedback control is expected not to reproduce the full
aligned-feedback lift, but this is diagnostic rather than a primary gate.

## Limits

The weak-label outcome remains a retrieval-usefulness proxy, not a causal task
success measure. Passing this protocol confirms the corrected ranking effect on
the reserved Codex block; it does not validate 5W1H superiority, Anti-Pigeon
coverage, omitted-influence coverage, or large-corpus latency.
