# Frozen visible-input coverage test

Isolate target-probe coverage without changing publication timing or fitting.
Use the same2688 consumed v120 records, query pool152..159 unknown at160,
latest63 known labels and reserved query slot. Keep the existing .01 hazard,
.95 generic-family mass, paid reveal161, latest64 publication, and31-frame
evaluation. No teacher information enters selection.

The original joint rule uses eight visible probes153..160. Compute the same
decision-time conditional laws for all512 possible9-bit inputs. Compare two
frozen empirical target distributions: all161 visible inputs0..160, and the
latest32 visible inputs129..160. Repeated inputs retain their multiplicity.
These distributions use visible X only, not future X, labels, teacher fields,
case identity, or the generator's known population distribution.

For each candidate j and input x, retain p_x and conditional p_{x|y}, including
the existing extra hazard step to virtual time161. Score
sum_x w_x sum_y P(Y_j=y|E_160)*(p_{x|y}-p_x)^2.
Verify it equals the reduction in Bernoulli Bayes risk under this declared
decision-time joint model. This is still not a certificate of actual gain after
natural publication evidence changes. Test that separately, not in this run.

Collect full512-input base/conditional arrays, answer masses, support, input
histograms and chosen origins. Use explicit separate refits as a reference
implementation, not an optimized replacement for the existing eight-probe batch.
Account for one base fit plus two fits per candidate on nonempty pools. Empty
pools do not fit or buy. No new threshold, learner, calibration or parameter fit.

Match old eight-probe and disjoint-probe conditionals and choices against the
existing artifacts. Independently rescore all512-input identities and both
histograms. Evaluate new selections using already fitted actual-answer branches,
both-answer sample expected risks and population expected risks. Preserve old
noquery/random/entropy/joint/disjoint controls and equal unit query costs.

Primary screen remains all21 phase1 delayed cells with lower gain bound>=-.001
against random and entropy, plus positive bounds for cases19/20 against both.
Mean +/-3.5SE across32 trajectories is descriptive, not a simultaneous/anytime
guarantee. Report both candidates and all84 cells; no winner selection or tuning
on this consumed evaluation, no untouched or whole-goal claim.

Before collection: guard bounds, duplicate-input weighting, mutation ownership,
unavailable/future-label and future-input poisoning, teacher/identity isolation,
race checks and independent old-conditionals equality. Hash sources and retain
failures. Time full reference collection separately from cheap histogram scoring;
do not claim sub100ms service latency from an offline component. No production,
paper, dependency, commit or push changes.
