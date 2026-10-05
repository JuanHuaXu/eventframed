# Frozen budget-policy off-grid evaluation

Freeze risk-budget-policy.mjs and all its model dependencies at the hashes in
risk-budget-evaluation.json. Do not change the guard's five noise points,
posterior prior, source masks, forecasts, actions, budget or tie rule.

Evaluate actual noise .125,.175,.225,.275 (interpolation) and .05,.35
(extrapolation), each under all16 masks:96 worlds. These values are evaluation
inputs only. The compiler still receives no realized noise or mask. The source
mechanism family remains the same; this is not new semantic or actual-agent data.
The four intermediate values are outside the discrete guard, but inside its
noise span; do not call them outside the prior's continuous noise support.

Keep .01 final nonharm and strictly positive area gain (excluding mask15) against
random, archived entropy and normalized entropy. Report interpolation and
extrapolation separately, including meaningful positives (>1e-12), numerical
ties and harms. No confidence interval is needed for exact finite population
enumeration, but no continuous-noise guarantee follows from six extra points.

Verification: frozen source hashes and compiler statistics; byte-exact replay;
independent ordered likelihood and backward score checks at off-grid values.
Preserve all original failures and do not tune after inspecting these results.
