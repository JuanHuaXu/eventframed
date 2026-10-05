# Tree-structured input law, not a new outcome tree

Prior check: context-tree v70 averages outcome predictors and failed noisy
interactions. Here only P(X) changes; existing subset outcome learning remains
available. No Chow-Liu input component was found in the local research tree.

Primary sources:
- Chow and Liu1968, original abstract at
  https://research.ibm.com/publications/approximating-discrete-probability-distributions-with-dependence-trees
- Bhattacharyya et al., Near-Optimal Learning of Tree-Structured Distributions
  by Chow-Liu,2021, introduction algorithm description:
  https://arxiv.org/html/2011.04144

Fit pairwise mutual information and maximum spanning tree from input fields
only. Our frozen adaptation adds .5 to each binary pair cell, consistent with
singleton pseudo-count1 per outcome. Root at field0, deterministic ties.
The resulting product law is positive and normalized; compile512weights for
the existing finite conditional-table interface. Labels are ignored.

No empirical-MLE or finite-sample theorem is inherited for this smoothed
adaptation. No causal interpretation, arbitrary higher-order dependence claim,
online error guarantee or validated adaptive-observation benefit. Tree topology
can overfit weak pairwise effects. A distribution with only higher-order input
dependence is an important future negative control.

Component checks: normalization, uniform limit, label-flip invariance, no
cycles, independent Prim/Kruskal objective comparison, smoothed edge marginals,
copied-field fixture, invalid inputs. Original models remain unchanged.
Next finite-data screen must retain independent/null/high-noise cases before
fresh integration; do not replace a failed outcome model with another tree by name.
