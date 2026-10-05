# Two-feature dependence diagnostic, frozen before execution

Use the original Bernoulli logistic likelihood with intercept N(0,1) and
two independent coefficient priors (1-pi)*delta0+pi*N(0,1), pi=1/255.
Balanced deterministic designs at n16 and n64: (a) identical observed
features and (b) all four feature combinations equally represented.
The observed outcome is the first feature's bit. No benchmark teacher or
screen outcome determines a hyperparameter. Score all four possible queries,
including discordant queries absent from the identical-feature design.

Enumerate all four inclusion states. Integrate original likelihood times
the Gaussian priors directly on [-10,10], with midpoint128 and192 panels
per active dimension (always integrate intercept). Compare inclusion-state
probabilities and predictive probabilities to2e-7 across resolutions.
This is numerical reference integration, not a formal tail/error certificate.

Fit the existing two-feature VI in forward and reverse coordinate order with
the SAME prior, initialization and1,024-cycle cap. Report both forecasts and
their equal-weight average. The average is only a two-fit ensemble diagnostic,
NOT a claim that equal weights recover the full Bayesian posterior. No
selection using the reference probabilities is allowed.

Falsifiers: reference prediction on identical-feature discordant queries
must be.5 by feature exchange and joint sign/label symmetry; reference single-
active state probabilities must agree. The independent-feature case is a
negative control: it has information distinguishing the features, so the
diagnostic must not force every query to.5. Compare evidence numerically
before explaining failures as prior mismatch or variational mode selection.
