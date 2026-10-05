# Endpoint objective: line-direction limit exposed

The [frozen endpoint objective](POPULATION_ENDPOINT_PROTOCOL.md) optimizes
all-genuine risk at .30 noise rather than averaging genuine risk over the
interval. It retains the same target, baseline, line family, population
constraints and all900 gates.

Result:873/900 (FAIL), below the average objective's888/900. All800 population
protection and50 false-confidence checks pass;27 gain checks fail.

| Actual noise | Pass /180 | Genuine gain failures /10 | Worst population harm |
| --- | ---: | ---: | ---: |
| .10 | 172 | 8 | .007032 |
| .15 | 175 | 5 | .008641 |
| .20 | 175 | 5 | .009553 |
| .25 | 176 | 4 | .009927 |
| .30 | 175 | 5 | .010000 |

Endpoint emphasis rescues three .30 gains relative to the average objective but
loses other gains. This is not minimax optimization or evidence that the
endpoint is worst for every possible rule.

## Numerical ceilings exclude five allocations in this family

Let v* be the minimum endpoint regret over the declared lambda box and
population coefficient constraints. The valid dual lower bound D satisfies
D<=v*, hence any attainable endpoint gain is at most -D. The numerical upper
bounds below include an additional1e-10 reporting cushion:

| Renewal counts [0,1,2,7] | Endpoint gain upper bound | Required |
| --- | ---: | ---: |
| [1,1,2,2] | .003292118 | .005 |
| [1,1,3,1] | .000001928 | .005 |
| [1,2,2,1] | .001873382 | .005 |
| [1,3,1,1] | .001933696 | .005 |
| [2,1,2,1] | .003576841 | .005 |

These ceilings are based on floating-point dual calculations, not formal
interval arithmetic. The gap to the requirement is far larger than the
reported numerical tolerances. [Exact stored values](population-endpoint-ceilings.json).

Thus changing only the objective or step-size allocation cannot make this SAME
line-restricted family pass every gain requirement. The excluded family is
p(x)=p0(x)+lambda(x)(target(x)-p0(x)), lambda in[0,1], with the existing uniform-
noise optimistic target and conservative population coefficient constraints.
This does not exclude different forecast directions, less conservative valid
risk descriptions, new observations or other models.

## Verification and local risk

All ten solves finish within707 sweeps, meeting the unchanged1e-8 dual-gap and
primal tolerances. The [full output](population-endpoint-objective.json) includes
all lambda mappings, solver diagnostics and gate failures. Full optimization
replay is byte-exact; original controls match across every world.

Alternate scoring with archived lambdas agrees on3200 Brier values within
8.88e-16. Eight hundred coefficient/direct-risk checks agree within3.69e-15.
Six analytic cases, twenty feasible-grid comparisons and one rejection test pass.
These checks do not constitute an independently implemented optimizer.
[Verification](population-endpoint-verification.json).

Worst SINGLE history/outcome loss increase is .287691. Population protection
does not imply per-event or priority-sensitive protection. This remains an
important limitation, not a benefit to conceal behind the gate count.

## Next meaningful experiment

Allow the full four-class forecast vector to vary under the same population
budget, rather than restricting it to a fixed line toward one target. Keep
all original quality tests and distinguish average, endpoint and worst-regime
objectives. This tests a genuinely larger decision family instead of continuing
to tune a numerically excluded step-size family.

All tests remain finite model-based research on consumed scenarios. No fresh
real-world confirmation, acquisition-speed validation, production change,
performance claim, whitepaper promotion or publication. All seven goals open.

