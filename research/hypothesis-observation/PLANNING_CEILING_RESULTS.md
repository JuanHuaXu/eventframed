# Known-mode planning: depth2 already reaches the optimum here

Exact population diagnostic under [the frozen protocol](PLANNING_CEILING_PROTOCOL.md).
This is NOT validation of unknown-provenance operation or completion of direction7.

## Results

Across copied05, copied20 and null, at each budget0..8, receding depth2 and
depth3 match full-horizon optimal terminal Brier to numerical precision.
Greedy target-Gini also matches except copied20 at budget4. Thus deeper-than-two
lookahead offers no measured quality headroom in this known-mode family.

| Regime / budget | Random | Report entropy | Greedy Gini | Depth2 | Exact optimum |
| --- | ---: | ---: | ---: | ---: | ---: |
| Copied05 / 3 | 0.473843 | 0.180975 | 0.132225 | 0.132225 | 0.132225 |
| Copied05 / 5 | 0.248748 | 0.180975 | 0.094711 | 0.094711 | 0.094711 |
| Copied20 / 4 | 0.625773 | 0.537600 | 0.527020 | 0.482605 | 0.482605 |
| Copied20 / 5 | 0.584525 | 0.537600 | 0.472347 | 0.472347 | 0.472347 |
| Copied20 / 8 | 0.471910 | 0.471910 | 0.471910 | 0.471910 | 0.471910 |

All null risks equal0.75. The full [artifact](planning-ceiling.json) retains
every budget, not only the rows displayed. No confidence intervals are needed
for this complete enumeration of a declared finite law; that does not confer
certainty that real data follow this law.

The copied20 four-report gain of0.044415 versus greedy is a genuine finite
planning effect, not additional evidence or teacher access. Every policy pays
for the same number of unique reports. However, the earlier v9 two-step study
already used deeper planning with unknown source modes and still failed its
broader screen. This known-mode diagnostic does not overturn that failure.

## Verification and cost scope

Each regime enumerates6,561 partial-report states; cached planning across all
policies/budgets uses17,496 subproblems and58,320 candidate branches. These are
aggregate audit work counts, not per-query CPU or loaded serving latency.
The audit checks the oracle lower bound, full-budget equality, null invariance,
mass conservation, and502 direct Gini/depth1 action comparisons per regime.

[Independent Python reference](planning_ceiling_reference.py) uses the original
experiment likelihood/update/classification APIs, sequential posterior updates,
and direct expected squared loss. All108 depth/budget/regime results agree
within3.61e-16. Its [verification artifact](planning-ceiling-reference.json)
records source hash and exact full-audit replay. The independent reference
checks depth1/2/3/full planning, not the random/entropy implementations separately.

## Next action

Do not spend the next experiment merely increasing lookahead depth. Remaining
leads are source-model validity and genuinely independent measurements. A fresh
measurement must contribute a new random draw conditional on the target and
must carry an explicit acquisition price; a different source label attached to
the same report is not renewal. Compare that option against ordinary reports
under unknown, independent, copied, mixed and misleading-provenance controls.

This changes the observation contract and therefore needs its own frozen
protocol. An oracle independence promise in a simulator is not an operational
provenance certificate. Old success criteria and failures remain intact.
