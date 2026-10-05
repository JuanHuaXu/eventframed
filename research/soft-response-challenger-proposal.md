# Soft-response challenger: next research lead

Status: the regularized logistic component is now implemented and numerically
verified; see [component results](ridge-challenger-component.md). Fresh quality
and composite integration remain untested. The variational Bayesian extension
remains proposed. V116 remains FAIL.
V117's consumed-data diagnosis points to fitted-model error: in all36 terminal
cells its generic full-input excess exceeds the absolute view and routing terms
combined. That does not prove the current learner can never fit these tasks or
that routing has no remaining value. It motivates a learner-level comparison
before another delayed-weight update.

## Evidence-linked choices

1. Add a bounded regularized logistic challenger. Unlike separate cell estimates,
   an additive parameterization can share evidence about one coordinate across
   many input combinations. This is a hypothesis about sample efficiency, not a
   promised result, and a linear logit cannot represent arbitrary interactions.
2. Compare a variational Bayesian logistic version only as a separately named
   approximation, with the same feature set and evidence. Do not call the MAP
   plug-in forecast a posterior predictive or claim approximate variance is
   calibrated without testing it.
3. Retain the existing context-tree learner as a structural comparison for
   hierarchical responses, and retain generic/Boolean controls for interactions.
   No learner gets the generator ID, relevance coordinates or change time.

## Primary research consulted

[Friedman, Hastie and Tibshirani (2010), Sections2-3](https://web.stanford.edu/~hastie/Papers/glmnet.pdf)
provide penalized logistic objectives and quadratic/coordinate-descent fitting,
including ridge as an elastic-net special case. Their Section3 describes weighted
quadratic approximations and warns about saturated fits. That supports a small
regularized fitting control; it does not establish EventFrame drift recovery or
provide our tuning parameters. A fixed-penalty bounded implementation would not
be a reproduction of their full regularization-path algorithm.

[Jaakkola and Jordan (1997), Sections2-4 and AppendixA](https://proceedings.mlr.press/r1/jaakkola97a/jaakkola97a.pdf)
derive a quadratic likelihood lower bound giving approximate Gaussian updates
under a Gaussian prior. The posterior predictive integrates parameters; it is
not simply the sigmoid at their mean. Their comparison explicitly observes
underestimated posterior variance. Accordingly any implementation needs an
explicit approximation budget and independent numerical checks; closed-form
updates are not evidence of calibrated uncertainty or robust concept drift.
These sections were inspected in the primary PDFs, not inferred from abstracts.

## Required implementation and falsifiers

Start with the cheaper regularized control to isolate representation from
posterior approximation. Freeze all penalties, feature scaling, initial state,
iteration limits, convergence tests and fallback behavior before quality runs.
Use all nine input coordinates plus an intercept, not the teacher's true subset.
Clarify the penalty's sum-versus-average likelihood convention and treatment of
the intercept. Keep the same as-of64/32 labels, feedback delays and fit cadence.

Check gradients/Hessians or updates against independent finite differences and
small numerical references. Test constant outcomes, rank-deficient designs,
extreme logits, label complement and feature permutation, deterministic replay,
finite probabilities and bounded failure. A convergence failure must be recorded,
not silently converted into a successful fit. Compile probabilities to the same
frozen input-measure interface and verify partial/full forecast coherence.

First compare full-input learner risk at equal evidence on all transfer families,
not only additive cases. Retain the original Boolean regimes as protection tests.
A logistic-only failure on interactions is not grounds to remove those tasks.
Then test an explicit, budget-matched composite through the adaptive observer;
direct learner success alone cannot close directions1/2/4. Preserve incumbent
specialists rather than deleting them to make room for a favored task family.

Use fresh disjoint quality seeds and retain the .01 harm/.005 gain criteria.
V116/V117 may guide the hypothesis but cannot serve as untouched confirmation.
Count parameter memory, fit/update cost, table compilation and partial observation
cost separately. No hot-path training, production promotion or whitepaper change
is justified until integration and performance evidence support it.

This lead remains falsifiable: a better full-input learner that loses its advantage
under partial observations or delayed feedback is not an integrated rescue. A
failed regularized model is not evidence that the Bayesian approximation will
fix it; that needs its own distinct test and error accounting.
