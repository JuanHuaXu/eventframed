# Exposing existing raw heads to the Brier combiner

Freeze before scoring; consumed independent-v1 only, no new confirmation.
Full pool indices [12,0,1,2,3,10] = Markov, generic64, Boolean64, generic32,
Boolean32, segment64. Prior [.5,.1,.1,.1,.1,.1]. No-segment pool [12,0,1,2,3]
has prior [.5,.125,.125,.125,.125]. The baseline always keeps half the mass;
the remaining mass is uniform over the declared alternatives, not tuned.

Compare strong substitution and linear averaging under identical eta2 weights
for each pool. The same origin-order delayed FixedShare refiltering is retained,
alpha_j=1/(j+1),alpha_0=0. Four arms in total. No guard, new fitter, new inputs,
teacher fields, scenario route or altered arrival schedule. Directly accessing
the raw heads introduces no new fit: they already generate Markov12.

This differs from v120's old six-expert composition: that used log-likelihood
emissions, generic64 prior.95, constant.001 switching, and no composite Markov
anchor. It also differs from the immediately preceding two-expert eta2 screen
only in forecast-pool access and the explicitly redistributed challenger prior.
This is a frozen structural comparison, not evidence that priors are optimal.

Same original 672 protection/96 recovery requirements per arm, all controls,
.01 upper non-harm and .005 mean/positive-lower gain. Use eight-index +/-3.5SE
descriptively, not original32-index confirmation. Retain whole/terminal/strata
and individual window harms. No pass-count optimization or case removal.

Verify multi-state path enumeration, two-head equivalence, permutation and
outcome symmetry, static immediate bound, missing/current/future label tests,
source-control score checks and exact replay. Measure combination work only;
source fit collection remains charged545.32s. No production/Go/paper changes.
