# Frozen outcome-simplex guard rescue

The noise-mismatch test failed because the original guard protected a finite
conditional-law family under an assumed likelihood. This rescue changes only
the guard family, replacing it with the four outcome-simplex vertices.

Keep the predictor's .20 likelihood, priors, local baseline, optimistic target,
epsilon=.01, projection solver, all ten allocations, sixteen copy masks and
five actual noise worlds (.10,.15,.20,.25,.30). Preserve all 900 gates without
tuning: 800 nonharm (<=.01), 50 genuine gain (>=.005), 50 false-confidence
reduction (>=.05). All ten-bit report vectors and sixteen latent hypotheses
are enumerated; expectations are exact finite sums.

For Brier loss L(p,y)=sum_j(p_j-1[j=y])^2, constrain
L(p,y)-L(p0,y)<=epsilon for every y in the four-class outcome space.
The expected regret under any q in that simplex is the q-weighted sum of
these four regrets, hence also <=epsilon. This guarantee does not require
a correct measurement likelihood, source model or target prior. It remains
relative to the same baseline for the same observed history, not an oracle,
and does not establish an absolute accuracy or calibration guarantee.

The proposal target is unchanged. Source posterior authenticity is not altered.
Numerical projection feasibility/objective checks must pass with no discarded
vectors. Compare original control scores against the preceding mismatch artifact.
Replay output exactly and recompute Brier by the conditional-risk identity.
Report failure to retain useful gains, should it occur; safety alone is not the
success criterion. No production integration or empirical coverage claims.

An unknown or enlarged outcome alphabet, changed scoring rule, feedback-induced
future trajectory changes, and loaded serving cost require separate analysis.
Pointwise protection concerns this forecast, not an entire counterfactual
closed-loop policy value. This convexity argument is directly derived here.

