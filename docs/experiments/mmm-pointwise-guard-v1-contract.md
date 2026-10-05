# Pointwise Brier guard: consumed-data diagnostic control

This is not a fresh confirmation and does not replace any research goal.
Keep the fitted forecasts and Fixed Share proposed weights unchanged; epsilon
remains .01. Do not tune it on remaining counterexamples.

For baseline b, challenger c, proposed weight v in[0,1], d=c-b, emit
p=b+w*d with the largest w in[0,v] satisfying

    d^2 w^2 + max(2db,2d(b-1)) w <= epsilon.

The left side is max over binary outcomes of Brier excess. Since expected
excess is affine in the unknown Bernoulli probability q, this bounds expected
excess for every q and hence its average on every nonempty window. It also
bounds realized excess. This is additive tolerance, not zero-harm, calibration,
consistency, recovery speed, or sublinear regret. If epsilon=0 the only changed
forecast satisfying both endpoint constraints is the baseline.

We derive this scalar bound directly; it is not an application of a borrowed
regret theorem. Related background: Bernasconi de Luca et al.,
[Conservative Online Convex Optimization](https://ecmlpkdd-storage.s3.eu-central-1.amazonaws.com/preprints/2021/sub_12975_17.pdf),
Section3 Definition1, uses a cumulative multiplicative comparator constraint;
that differs from the pointwise additive requirement here. Its assumptions
and no-regret result are not inherited by this guard.

Screen three fixed proposals under the same guard: existing Fixed Share,
constant half, and full challenger. Retain existing local-ledger/Markov
controls. These controls test whether the safety price erases useful learning
and whether adaptive weighting adds value beyond simple mixing. Report whole,
terminal64, every32-step window, stationary/shift strata, and actual changed
forecast rates. A trivial baseline fallback does not complete the goal.
