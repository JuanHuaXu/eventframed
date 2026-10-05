# Primary-source check after mass-calibration failures

Read2026-09-15. These papers motivate distinct tests, not a claim that our
implementation reproduces their empirical results or inherits their guarantees.

## Target Distribution

Bickford Smith, Kirsch, Farquhar, Gal, Foster and Rainforth,
[Prediction-Oriented Bayesian Active Learning](https://proceedings.mlr.press/v206/bickfordsmith23a.html),
AISTATS2023. Full PDF sections3-4, especially equations3-5 and section4.1,
were inspected. EPIG averages information about target predictions over a target
input distribution, rather than maximizing parameter information. The target
samples need not themselves be label-acquisition candidates.

Our Bernoulli Brier-risk reduction is not their entropy/KL criterion. The common
design issue is declaring a relevant target distribution. The visible161/recent32
experiment changes empirical target coverage only. A later entropy-versus-Brier
comparison could reuse the same conditional laws, but is a separate untested
lead. Neither criterion repairs a wrong joint model automatically.

## Correlation Quality

Wang, Sun and Grosse,
[Beyond Marginal Uncertainty: How Accurately can Bayesian Regression Models Estimate Posterior Predictive Correlations?](https://proceedings.mlr.press/v130/wang21g.html),
AISTATS2021. Full PDF sections1-3 were inspected. Their evaluation separates
predictive correlations from marginal uncertainty, uses transductive acquisition,
and describes cross-normalized likelihood and oracle metacorrelations. One
evaluation holds the prediction model fixed while varying the selection model.

Their setup is Gaussian regression; copying its normalization into our Bernoulli
regime model would require a separate derivation. Our practical inference is to
audit the conditional response between a candidate and target, not just candidate
mass MSE. Keep marginal calibration, empirical target coverage, and delayed
publication changes as separate experimental factors. A known teacher's label
probabilities are not by themselves an oracle posterior over unknown functions.

## Additional Lead, Not Formula-Verified

Tang, Sloman and Kaski,
[Representative, Informative, and De-Amplifying: Requirements for Robust Bayesian Active Learning under Model Misspecification](https://proceedings.mlr.press/v300/tang26d.html),
AISTATS2026. The official proceedings abstract identifies representativeness,
informativeness and error amplification as distinct considerations and names
R-IDeA. Full-formula review remains pending: the linked PDF failed to load,
and OpenReview requested browser verification. No verification was bypassed.
Do not invent its acquisition equation or describe an implementation as R-IDeA
until the actual method and assumptions can be inspected.

## Research Boundaries

Coverage is currently being tested with unchanged publication. If it fails,
natural-evidence changes between decision160 and publication161 remain an
independent diagnostic; so does entropy-based predictive information under
the same model. These are live leads, not validated rescues or grounds to
declare the seven-goal research program complete.
