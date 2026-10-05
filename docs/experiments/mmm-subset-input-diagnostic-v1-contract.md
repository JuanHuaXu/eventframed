# Subset input marginalization diagnostic

Classification: confirmed uniform-input assumption, not a software bug. Whether
it causes a material portion of current failures needs investigation. Existing
empirical-input forest experiments are prior work, not a new proposed invention.

Freeze 16 independent fits each at n64 and n128 for two input laws: uniform
nine bits and uniform raw inputs with bit0 copied from bit2. Outcome is bit2
with independent .05 label noise. Seeds are 2026092140+n*100+index, plus100000
for dependent inputs. This deliberately simple component probe is not a new
benchmark of broad adaptation or a confirmation of earlier aggregate results.

Fit identical subset outcome models with uniform, past-audit empirical
(one total pseudocount), or evaluator-only oracle input weights. Oracle weights
use a 1e-12 floor because the current constructor requires positive cells.
Exact expected Brier sums over512 raw inputs, no sampled evaluation labels.
Primary diagnostic observes only bit0. Full-input forecasts must be identical
across all three arms: changing an input marginal must not alter P(Y|X).

Partial Bayes floors are .25 for independent inputs and .0475 for copied inputs.
Inspect outcome estimation plus input marginalization separately. No promotion
threshold, parameter sweep, causal inference, deployment or whole-goal claim.
Falsifier for a strong marginalization-only explanation: even oracle weights
leave most error, or empirical weights offer negligible benefit. A positive
diagnostic still needs integrated, fresh, equal-cost testing.
