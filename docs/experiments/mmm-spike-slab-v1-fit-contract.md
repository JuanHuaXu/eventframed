# Fixed-prior spike-and-slab fitting contract

This is an inference component, not a prior-selection or trajectory-quality
experiment. Inputs declare pi in (0,1), finite slab variance c2>0, unique
nonconstant Walsh masks, at most 64 observed labels, and a cap in [1,1024].
Initialize every inclusion probability to pi, slab mean to zero, variance to
c2, and xi to one. Masks retain their declared coordinate order.

Numerical amendment after the recorded pilot saturation failure: canonical
factors retain finite inclusion log-odds. Compute both complementary masses
and their logs stably; never obtain the small tail by subtraction from one.
Entropy, variance and predictive atoms use that same pair. This changes no
prior, variational family, stopping threshold or coordinate equation. See
`research/spike-slab-logodds-repair.md` for the regression and underflow limits.

Each cycle: fixed-xi coordinate sweep, update xi from joint moments, and
reprofile the conditional intercept. Return the complete factor/profile state
at the new xi. Record the bound before sweep, after sweep, and after reprofile.
Reject nonfinite states or a bound decrease larger than 1e-9 absolute. Do not
silently clip saturated inclusion probabilities. Keep the reference sweep
independent for comparison; no feature pruning or likelihood change occurs.

Stop only when BOTH full-cycle bound change is <=1e-6 and normalized state
motion is <=1e-6, or on the declared cap. State motion is the maximum change
in inclusion, slab mean divided by 1+abs(old mean), log slab variance, and xi
divided by 1+old xi. This is an operational numerical criterion, not proof of
a global optimum or accurate posterior. Retain and report capped fits.

Unit tests must cover exact initialization, fixed-budget prefix equality,
direct dense-trajectory agreement, complement symmetry, input isolation,
invalid budgets/priors, full-basis fitting, and monotonicity. Measure complete
fitting separately from setup plus one sweep. Forecast-law integration and
the frozen as-of pilot are now separate components; their results do not
follow merely from this fitter contract.
