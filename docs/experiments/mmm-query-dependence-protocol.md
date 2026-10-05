# Marginal-preserving Bernoulli dependence calibration

## Source and Derivation

Wang, Sun and Grosse, AISTATS2021, sections1 and3:
https://proceedings.mlr.press/v130/wang21g/wang21g.pdf
separate predictive dependence from marginal uncertainty and use Gaussian
cross-normalized likelihood. This experiment is not their XLL implementation.
It is a binary, marginal-preserving mixture derived for the current artifacts.

Let p=P(Yquery=1), f_y=P(Ytarget=1|Yquery=y), b=(1-p)f0+p*f1.
Define Q_lambda=(1-lambda)(Q_query x Q_target)+lambda Q, 0<=lambda<=1.
Then f_y,lambda=b+lambda(f_y-b). Both marginals are unchanged; the joint law
remains normalized and nonnegative. Target aging is applied consistently first.
Pairwise covariance is multiplied by lambda. Model Brier-information gains
are multiplied by lambda^2, so a positive global lambda cannot change their
exact mathematical ordering. Zero yields all ties. Do not call this a ranking
rescue, nor assume the pairwise mixture defines a full multivariate process.

## Frozen Design

Use all2688 consumed source/coverage records. Nonempty delayed pools provide
query origins152..159; all31 target frames161..191 are evaluated. Weight each
history equally, its candidates equally, and its targets equally. Complete
histories have no candidates and remain explicit zero-pair records.

Fit one lambda from actual query and target binary labels in phase0 only, using
weighted conditional log loss. Query marginal log loss is constant in lambda,
so this also minimizes the joint log loss. Fit by convex derivative bisection
on[0,1],80iterations, with exact endpoint tests. No teacher probabilities, case
identities, phase1 labels, or true model parameters enter the fitter. The archived
phase0 query labels include originally missing/delayed evidence: this is offline
supervision, not free online feedback. Phases have already been consumed; phase1
is a transfer screen, not untouched confirmation. Never count pairs as
independent trajectories.

Compare lambda0 (independence), lambda1 (original), and the training fit. Evaluate
actual joint log loss, conditional log loss, realized target Brier, target-law
expected conditional scores given the actual query answer, and teacher-product
expected joint log loss as a supplementary fixed-generator expectation. The
last quantity is not an oracle posterior over unknown teachers. Report all84
cells, phase0/phase1 delayed means, and paired trajectory contrasts.

The descriptive improvement screen requires phase1 lower gain bound >=-.001
on every cell and >0 on transition cases19,20, versus BOTH independence and
original. Apply separately to actual joint log loss and target-expected Brier.
Use the existing mean +/-3.5SE procedure; no sequential significance claim.
Require identity/relabeling/optimization tests, phase1/teacher isolation,
source hashes, independent aggregation and exact replay. No production or
whitepaper changes. This tests pairwise response calibration only.
