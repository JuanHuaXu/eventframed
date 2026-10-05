# Predictive-risk acquisition result

Status: FAIL.840/840 non-harm checks pass;0/80 required gain checks pass. All2688
consumed trajectories retained. This does not validate an active-learning upgrade.

Phase1 delayed terminal64 expected Brier:

| Case | Natural | Random | Exact state information | Predictive risk |
| --- | ---: | ---: | ---: | ---: |
| Additive stationary | .221747 | .221572 | .221389 | .221452 |
| Parity4 | .048364 | .048320 | .048304 | .048303 |
| Null | .258712 | .258441 | .258261 | .258335 |
| Majority to parity | .057088 | .056392 | .056522 | .056203 |
| Parity to majority | .098247 | .094506 | .095908 | .095440 |

Paid policies use the same31 queries per listed trajectory. The predictive-risk
criterion improves the two shift means compared with exact state information,
but does not consistently beat random or meet the declared improvement gates.

Verification:361 explicit joint-outcome Brier identity tests,252 sampled as-of
checks, per-candidate conditional-risk identities and matching query counts pass.
Immediate-complete schedule has no queries and identical forecasts. Full replay
is byte-identical (`cmp` exit0). Initial complete offline job10.28s wall; no
serving-latency claim. No expert is retrained by the acquired labels in this replay.

## Limits and next work

Checked an alternative explanation: none of the80 mean-gain requirements is
ruled out by the per-frame convex-hull oracle. For every delayed terminal shift
cell, each control's mean Brier minus the matching oracle hull is at least.005.
This uses `mmm-switch-budget-v120.json` matched by phase/case/schedule/segment.
The gates therefore are not trivially impossible for this fixed expert family.
That is a hindsight ceiling, not proof that the gain is learnable with31 queries.

[Roy and McCallum (2001), author's manuscript](https://groups.csail.mit.edu/rrg/papers/icml01.pdf)
motivates directly targeting expected prediction error rather than parameter
uncertainty. Our finite Brier/virtual-probe calculation is an adaptation, not a
reproduction of their document-classification learner or empirical claims.

Three targeted acquisition candidates fail their broad added-value requirements:
disagreement, exact current-state information, and predictive risk; entropy
selection was retained as a comparator throughout.
Stop tuning this fixed-tape acquisition family for now. A more informative next
test must allow acquired labels to update the underlying experts, keeping
natural/random/targeted controls, matched budgets and publication timing. Existing
Go `softV120Run` fits generic, Boolean and segmented predictors every32 frames
from arrived evidence; it is the appropriate starting point for that test.
Freeze a new protocol before outcomes and use fresh seeds only after component
and consumed-data checks. All seven full research directions remain open.

Artifacts: [protocol](mmm-predictive-query-v120-protocol.md),
[results](mmm-predictive-query-v120.json), [replay](mmm-predictive-query-v120-replay.json),
[timing](mmm-predictive-query-v120-timing.txt).
