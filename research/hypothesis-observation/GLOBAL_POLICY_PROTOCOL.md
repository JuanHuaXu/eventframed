# Global trajectory-policy risk game

Consumed-grid design, not fresh confirmation. Freeze the predictive laws and
80 worlds from risk-budget-evaluation.json. Optimize a mixture of complete
six-query policies, sampled once before a trajectory; each policy observes only
the root reports and subsequent acquired evidence, never the realized regime.
Actions may depend on sufficient counts and root pattern. No statewise safety
constraint is imposed. Keep all three controls and original thresholds.

Payoff rows: candidate final Brier minus each control minus.01 (240 rows), and
candidate area Brier minus each control (225 rows excluding mask15). Minimize
the largest row. A strictly negative optimum would satisfy the entire finite
screen; a positive valid lower bound would rule out feasibility in this frozen
forecast/policy class. Zero does not decide the strict-area requirement.

For a fixed nonnegative row-weight distribution, find an exact linear best
response by backward dynamic programming over count states. State cost is the
weighted UNNORMALIZED history joint mass times squared loss, with area divided
by6 and final loss charged only at depth6. Recursion adds both child values,
without additional probability factors: likelihood is already in each cost.
The shared count state's future options and likelihood do not depend on order.
Linear minimization therefore needs no policy dependence on the hidden world.

Run128 rounds of exponential adversary reweighting, learning rate32, uniform
initial row weights. This is a declared heuristic for approaching a saddle point,
not an assumption that128 rounds converge. Uniformly mix all generated policies.
At each round report max mean row payoff (primal upper bound) and the best
weighted best-response value so far (dual lower bound). No tolerances or gates
are relaxed. Bounds are floating-point, not rigorous outward-rounded intervals.

Strict argmin in fixed type order provides deterministic replay. All candidate
costs include the common forecast loss; control thresholds are constants outside
the oracle. Verify weighted forward scores equal backward oracle cost, dual<=
primal, and exact best response is no worse than every archived control under
the same weights. Save round risks, row weights and bounds. Replay plus tiny
exhaustive oracle tests are required. No default or production changes.
