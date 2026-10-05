# Held-out input forest component

Motivation: forcing8pairwise edges can fit noise under independence. Preserve
higher-order XOR failure explicitly; pruning cannot invent a missing interaction.

Primary source: Liu et al., Forest Density Estimation, JMLR2011:
https://jmlr.org/papers/volume12/liu11a/liu11a.pdf
The paper uses held-out likelihood to choose forest complexity and warns that
searching many structures on small held-out sets can overfit. Our discrete
adaptation searches only the training tree's9nested prefixes plus uniform.
No kernel-estimation theorem or optimality guarantee is inherited.

Split the already available sample sequence in half chronologically. First
half fits the existing smoothed input tree. Sort its8edges by fitted MI and
construct0..8edge forest laws plus fixed uniform. Select highest second-half
input log likelihood, ties within1e-12 favor simpler earlier candidate. No
outcome labels, test inputs or refit on validation enter the published law.
The outcome learner can still use all past labels; the input law pays the
training-data reduction explicitly. Minimum4samples, maximum8192.

Tests: normalized positive masses for everycandidate; exact full-support
independent and copied fixtures choose uniform/oneedge; held-out scores agree
with literal selected-law likelihood; labels irrelevant; validation actually
changes scores; invalid inputs rejected. Fit costs measured separately.

This remains a component, not a rescue. Next freeze a fresh comparison with
uniform, histogram and unpruned tree including high/noise/null/XOR cases.
Do not tune split ratio or reuse a validation score as a generalization bound.
