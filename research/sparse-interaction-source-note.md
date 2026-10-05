# Sparse interaction lead: source inspection

Provisional lead, not an implemented rescue or a validated claim.

Tipping, [Sparse Bayesian Learning and the Relevance Vector Machine](https://www.jmlr.org/papers/volume1/tipping01a/tipping01a.pdf),
JMLR2001, section3 and equations23-27: Bernoulli likelihood with a sigmoid of
a weighted basis expansion, one Gaussian prior precision per coefficient,
and a Laplace approximation around the penalized logistic mode. This provides
a source-grounded distinction from tying every same-degree feature's variance.
The classification model does not add the regression noise parameter.

Read the author's [errata](https://www.miketipping.com/jmlrtypo.htm): equation27
requires the working response `t_hat = Phi*w + B^-1*(t-y)`, not raw t. Do not
copy the uncorrected update. Sparse coefficient estimation is not a guarantee
of calibrated probabilities or of success on our shifted, small-window data.
Section3 also points to a variational alternative; inspect that primary source
before choosing between Laplace and variational inference.

Local comparison: `internal/observationlearners/variational_logistic.go` uses
ridgeFeatures and a fixed identity precision. The current kernel experiments
tie variances by degree. The existing Boolean specialist already averages
single parity atoms; do not rename that as a new sparse model. A genuinely new
candidate must support multiple weighted interactions together and retain
the old specialists as controls. Check the rest of the implementation before
claiming this capability is absent everywhere.

No code was added for this lead. Before implementation: settle the predictive
law and treatment of precision uncertainty, derive bounded sample-space
updates if feasible, freeze prior/optimization/pruning rules, and validate
against independent dense solves. Preserve all broad quality and cost gates;
do not choose feature masks using the simulator's teacher metadata.
