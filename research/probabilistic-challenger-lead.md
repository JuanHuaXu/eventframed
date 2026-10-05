# Next challenger lead: average small conditional models

Status: RECOMMENDATION ONLY; no new success claim. V12's own-path tree failure
argues against merely granting that tree more control. Test structural model
averaging instead of committing to a greedily fitted tree on64 labels.

Primary-source leads (metadata/abstract checked; full derivations not yet audited):

- Willems, Shtarkov and Tjalkens (1995),
  [The Context-Tree Weighting Method: Basic Properties](https://www.cs.cmu.edu/~aarti/Class/10704_Fall16/CTW.pdf).
  The paper combines model and parameter uncertainty for binary tree sources.
  Its sequence-source guarantees cannot be imported into arbitrary EventFrame
  covariates or selective feedback.
- Chipman, George and McCulloch (1998),
  [Bayesian CART Model Search](https://www2.stat.duke.edu/~scs/Projects/Trees/BayesianCART/Chipman1998_BayesCART.pdf).
  A prior over tree structures and terminal-node parameters offers an alternative
  to selecting one greedy structure. Our proposed finite enumeration is not its
  stochastic CART search and should not be called a faithful implementation.

Proposed research adaptation, not a formula attributed to either paper:
enumerate feature subsets in the fixed nine-bit test domain; put Beta priors on
each subset's conditional outcome cells and a frozen complexity prior on subsets.
Average full-input predictions using their marginal likelihoods. Partial forecasts
still require the declared past-audit input model. Fit only admitted labels;
retain count incumbents and test equal-label budgets. No real-world truth claim
follows from a posterior model weight.

Before implementation: verify likelihood bookkeeping and order invariance,
declare the subset prior and maximum complexity independently of new test results,
and assess the 3^9 conditional-cell bound. Include XOR, irrelevant correlated
features, null outcomes and stationary protection. Then require untouched online
tests; a good fit on consumed v9-v12 trajectories is not confirmation. The
finite-domain enumeration must not be presented as a scalable general backend.
