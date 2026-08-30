# Remaining claims completion protocol

Frozen on 2026-08-29 before executing the experiments described here.

## Evidence tracks

1. **Chronological Codex track.** Read-only sessions with timestamps in
   `[2026-08-28T00:00:00Z, 2026-08-30T00:00:00Z)` form a cold-start block.
   Raw text and production timestamps are not exported. Labels remain the
   frozen downstream-use anchor proxy. This track tests structured versus
   unstructured frames, residual reuse, selective updating, and priority value.
2. **Independent stress track.** A design seed of `32452843` and untouched
   confirmation seed of `49979687` drive heterogeneous residual, group,
   omitted-influence, injected-drift, and snapping scenarios. Real replay
   streams may supply covariates, but injected outcomes or changes are labeled
   semi-synthetic.

No failed criterion may be replaced or retuned after confirmation inspection.
Every claim is reported as validated, falsified, or inconclusive.

## Frozen criteria

| Claim | Primary criterion |
| --- | --- |
| Structured frames | Structured minus unstructured trajectory-cluster 95% lower bound is above zero for priority-weighted Brier gain; interpretability remains untested unless independently rated. |
| Residual reuse | Heterogeneous confirmation Brier-gain 95% lower bound above zero, harmful false-reuse upper Wilson bound at most 5%, and application/maintenance counts reported. |
| Selective Bayesian update | Activation at most 60%; Brier increase versus update-all has 95% upper bound at most 0.01; naive selective and no-update controls reported. |
| Anti-Pigeon | False-merge and false-split Wilson upper bounds at most 5%, with at least 80% correct terminal decisions for compatible and divergent groups. |
| Omitted influence | At least 95% empirical certificate coverage with 95% Wilson lower bound at least 90%; never-nominated cases included and reported separately. |
| Changepoint adaptation | Existing v5 per-scenario detection, unmatched-alarm, and delay criteria apply unchanged; real-stream perturbations are labeled semi-synthetic. |
| Predictive snapping | Confirmation gain 95% lower bound above zero for beneficial snaps, harmful-snap acceptance upper Wilson bound at most 5%, and rollback restores the pre-snap forecast exactly. |
| Priority weighting | High-priority miss-rate improvement 95% lower bound above zero while overall Recall@10 degradation is no worse than one percentage point. |

All stochastic proportions use fixed-terminal-sample 95% Wilson intervals.
Paired losses and retrieval gains use trajectory-level bootstrap intervals.
These are not sequential confidence sequences.

## Limits

The Codex anchor label is a retrieval-usefulness proxy, not causal task success.
Injected drift and known-truth group experiments establish mechanism behavior,
not prevalence in production. If fewer than three post-cutoff Codex trajectories
produce scored cases, the real-session claims are automatically inconclusive.
