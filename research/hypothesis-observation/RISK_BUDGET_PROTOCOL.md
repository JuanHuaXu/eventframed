# Uniform remaining-horizon risk budget

Design diagnostic, consumed80-regime grid. Keep the source family, reference
tie_entropy policy, frozen forecasts, prior-average objective and tie rule from
REGIME_SAFE_PROTOCOL.md. Change ONLY final-loss eligibility: with r queries
remaining require f_g(s,u)<=F_g(s)+r*(0.01/6+1e-12) for every supported regime.
Keep area eligibility a_g(s,u)<=A_g(s)+r*1e-12 unchanged.

This is a uniform remaining-horizon allowance, not optimized allocation across
branches and not an empirical confidence bound. Terminal allowance is zero;
at the initial state it is .01 plus numerical slack. Reference-action feasibility
follows by the previous induction because child budgets are smaller. The actual
realized regime is not supplied to the selector. No success threshold changes.

Expose the total budget as an argument only to verify that zero exactly recovers
the archived zero-budget policy. The only nonzero experimental budget is .01;
do not tune it against results. Reject invalid, nonfinite or negative budgets.
Compare all controls and zero-budget results, record actual changes by decision
depth and retain FAIL unless every original condition passes.

Verify zero-budget action parity, byte-exact replay, archived-score parity,
separate backward scoring, all state/regime budgets and branch normalization.
No production or archived source changes. The guarantee is limited to the
declared grid and is stronger than a population-only constraint, so failure
does not exhaust trajectory-level risk allocation methods.
