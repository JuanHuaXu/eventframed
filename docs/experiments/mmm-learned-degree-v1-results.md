# Learned degree v1: improved priors, incomplete optimization

All four bounded learned-kernel variants FAIL the broad gates. The experiment
completed129024 fits over2688 consumed records, including86016 learned fits.
Collection182.57s, replay183.81s, byte-identical forecasts and optimizer traces.
These are sequential offline test-body times, not serving latency or fresh
confirmation. All seven research directions remain open.

## Quality

| Candidate | Non-harm | Recovery gain | Broad verdict |
|---|---:|---:|---|
| Learned orders, fixed noise64 | 577/840 | 55/128 | FAIL |
| Learned orders, fixed noise32 | 508/840 | 41/128 | FAIL |
| Learned orders and noise64 | 320/840 | 13/128 | FAIL |
| Learned orders and noise32 | 309/840 | 34/128 | FAIL |

Five original controls and all thresholds remain unchanged. Counts are
correlated gates, not independent successes. All paired intervals are recorded
in `mmm-learned-degree-v1-summary.json` and `-contrasts.json`.

| Matched contrast | Non-harm | Recovery gain |
|---|---:|---:|
| Fixed-noise64 vs fixed degree64 | 160/168 | 28/32 |
| Fixed-noise32 vs fixed degree32 | 156/168 | 21/32 |
| Learned-noise64 vs fixed degree64 | 168/168 | 32/32 |
| Learned-noise32 vs fixed degree32 | 131/168 | 7/32 |
| Fixed-noise64 vs scalar64 | 138/168 | 28/32 |
| Fixed-noise32 vs scalar32 | 132/168 | 28/32 |
| Learned-noise64 vs scalar64 | 168/168 | 30/32 |
| Learned-noise32 vs scalar32 | 93/168 | 11/32 |
| Learned-noise64 vs fixed-noise64 | 80/168 | 6/32 |
| Learned-noise32 vs fixed-noise32 | 80/168 | 6/32 |

Passing every check against one weak fixed-prior control is not a broad rescue.
The learned-noise64 variant illustrates this distinction directly.

Historical phase1 delayed/missing terminal64 expected Brier, 64-label fits:

| Case | Fixed noise | Learned noise | Logistic | Markov |
|---|---:|---:|---:|---:|
| Additive stationary | .208872 | .225100 | .207774 | .221755 |
| Additive gradual | .232208 | .245238 | .244261 | .239065 |
| Hierarchy gradual | .244633 | .260268 | .263783 | .244456 |
| Local-table gradual | .249113 | .267008 | .272673 | .250387 |
| Parity4 stationary | .228246 | .197653 | .285755 | .049066 |
| Majority to parity | .240897 | .225993 | .288538 | .060262 |
| Parity to majority | .149301 | .145429 | .114788 | .102272 |

Learning order weights helps relative to the two preceding fixed priors, but
does not replace the stronger Boolean/Markov specialists. Learned noise is not
uniformly better than fixed noise, despite its improved training objective.
The Gaussian working likelihood and the Bernoulli forecast score remain
different objectives; no ordinary Bayesian calibration guarantee is asserted.

## Crucial optimizer limitation

85902/86016 fits (99.87%) hit the16-step cap. Only114 met projected-gradient
tolerance:40 of21504 learned-noise64 fits,74 of21504 learned-noise32 fits, and
none of the fixed-noise fits. No line-search cap occurred. Mean final projected
gradient norms are .12249/.07262 (fixed64/32) and .94628/.53715 (learned64/32),
well above1e-6. All traces descend, but descent is NOT convergence.

Mean negative-log-likelihood decreases are65.961/37.306 for fixed64/32 and
74.587/42.126 for learned64/32, excluding the parameter-independent normalizing
constant. Mean evaluation counts33.57/33.00 and48.44/45.21 fit the declared
work budget. Final mean order budgets are approximately:

- Fixed64: [.13566,.04548,.02841,.04278], noise1.
- Fixed32: [.13560,.04467,.03005,.03481], noise1.
- Learned64: [.20595,.14463,.14925,.21475], noise .24397.
- Learned32: [.22113,.17060,.17153,.22717], noise .16983.

These averages over all fits are descriptive, not recommended hyperparameters.
The quality failure rejects these bounded optimizer/learner combinations; it
does not falsify fully optimized order-weight learning or identify its optimum.

## Verification and numerical repair

The independent audit validates every86016 optimizer records' trace, parameter,
stop and work bounds. A separate pivoted Gaussian solver reconstructs1008
stratified learned fits,32256 forecasts,2016 initial/final objectives and4536
finite-difference derivatives. Maximum errors: forecast1.133e-14, objective
1.066e-12, projected-gradient norm1.340e-8. Full replay is byte-identical.
All1376256 linear control forecasts match the prior experiment exactly, and
the standard score audit checks4128768 candidate probabilities and161280
archived control metrics.

Gradient, initial fixed-model equivalence, invalid/constant/conflicting-duplicate
inputs and evaluator/future/unavailable-label poisoning tests pass under the race
detector. This verifies the research adapter, not a production asynchronous path.

The first independent audit stopped because Go's log(.01) is
-4.605170185988092 while V8's is -4.605170185988091, a one-ULP difference.
The verifier now permits only a four-epsilon, scale-adjusted floating boundary
envelope and tests rejection of1e-8 violations and NaN. It records13519 boundary
roundoffs, maximum8.882e-16. No optimizer, data, policy threshold or forecast
was changed. The failed empty `-audit.json` remains; use `-audit-verified.json`.

## Component cost and next work

Three Apple M4 repetitions: fixed-noise64 fitting4.147-4.192ms; learned-noise64
4.649-4.678ms,6896 allocated bytes and16 allocations per fit. Fixture-dependent,
initialization/stack storage excluded; no concurrent research jobs during this
benchmark. The unchanged256-feature point predictor was measured previously;
it was not independently rebenchmarked here. No e2e latency claim follows.

First separate optimizer convergence from predictive usefulness. A bounded
quasi-Newton or independently converged reference using the SAME objective,
bounds and eligible samples can reveal whether these remaining gradients matter
for forecasts. Do not respond only by tuning more fixed weights on consumed
scores. Any optimizer change needs its own derivatives/descent/boundary audit
and unchanged quality controls. Better training likelihood may still worsen
proper-score calibration. A subsequent expert-composition test could preserve
Boolean specialists while adding the genuinely new soft-law candidate, but is
not yet tested and must retain arrival-time evidence accounting.

Dependency reconnaissance: cached Gonum0.16.0 includes unconstrained BFGS/LBFGS;
it is not declared directly in this repo's go.mod and is not a bounded optimizer
by itself. Bundled Python has no SciPy. Nothing was installed, added to go.mod,
or changed globally. Neither missing SciPy nor a failed candidate exhausts the
remaining research leads.

## Hashes

- Forecasts: `c6af455f2601170ae9e50b924144cd9a16f3bd5dd797d5640c0408c90a96e41d`
- Summary: `e1891b7e9c4be88f8dbd3b64f2c715e6fa4430ac8f51e80945d353e89fce938e`
- Contrasts: `13c5771b5f005b365365dbf2e0343493465c4f07af9270a1a40f2b60503be8f2`
- Component: `93e9be45d8163f282332511e4fa5511fb83a83af906ebc23ce5f036fd4c66943`
- Collector: `e33fd73aa705085f461d57db88f11b1284e96292c9b02f266e0bb8c3408cae51`
- Independent audit: `2ad98220ec8914023e97efd4ac1bd8c6e5a7923533ae422275cff5e8e0c303ac`
- Contrast code: `a334069bb35d6f538594b1139c063564237a9a1a879e6e7b2441c38d705e1d84`
- Protocol: `05d81472256762f9b97543bf719fed62a493497d8861d0278559c588337f3d33`

Input and reused summary/feature code hashes remain those of the preceding
spectral/degree experiments. No production, paper, commit or push changes.
