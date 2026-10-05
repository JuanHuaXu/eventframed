# Count forecasts cannot be wrapped unchanged as one joint law

Classification: confirmed compatibility limitation of the isolated finite
research estimator, not a demonstrated production defect or proof of the cause
of all acquisition failures. No existing estimator is changed.

## Prior Work Check

The earlier coherent observer requires common input marginals and additive
joint cells (`research/joint-observation-component-results.md`). Adaptive v103
already tested that design: one-step late parity accuracy94.74%, lookahead94.80%,
but deeper planning failed the additional-gain criterion. Repeating deeper
lookahead without a new discriminator is unjustified.

The current retained-subset workflow differs: count forecasts use an independent
Beta(1,1) smoothing calculation at each partial mask. Its output need not be
consistent with any common joint distribution, even before experts are mixed.
Separate conditional estimates can still be useful empirical predictors.

## Exact Counterexample

Training samples: x=0,y=1 and x=1,y=1. The existing research count model gives:

    P(Y=1)             = (2+1)/(2+2) = 3/4
    P(Y=1 | bit0=0)    = (1+1)/(1+2) = 2/3
    P(Y=1 | bit0=1)    = (1+1)/(1+2) = 2/3

For any distribution of bit0, the weighted average of the two child forecasts
is2/3, never3/4. This violates the necessary law-of-total-expectation condition.
It cannot be fixed merely by choosing a different input marginal.

The diagnostic enumerates all59049 parent/one-bit partition edges in the
nine-bit space per fixture and tests whether the parent forecast is outside
the convex hull of its child forecasts. A positive violation proves incoherence;
absence of a violation on this test alone would not prove full coherence.

| Fixture | Samples | Violating edges | Maximum defect |
|---|---:|---:|---:|
| Two positive cells |2|256|.0833333|
| Balanced full-space negative control |1024|0|0|
| Noisy parity, sparse sample |64|5247|.0833333|
| Noisy parity, larger sample |4096|5786|.0666667|

These counts weight partial states equally, not by their probability in serving.
They are not deployment error rates or proof that4096 samples are worse than64.
All past measured quality results remain unchanged.

## Rescue Boundary

Not ready to patch the original estimator. Reusing the coherent observer
requires a declared model change, not just a guide-selection wrapper.
A candidate can define one joint prior over full input/outcome cells and
marginalize it consistently. For uniform input prior u(x), total prior mass a,
and symmetric outcome prior, an observed partial context A would use:

    q(A) = [n_yes(A) + a*u(A)/2] / [n(A) + a*u(A)]

Its evidence mass is n(A)+a*u(A), which is additive under partitions. At a=2
the empty-context smoothing matches the current model, but finer contexts do
not. This is a new candidate requiring proper controls, especially for sparse
cells and changing input distributions; no accuracy improvement follows from
coherence alone. Mixtures of experts with different input marginals further
require input-evidence-dependent conditional weights, or one declared common
input model. A fixed average of their conditional forecasts does not generally
equal the conditional of their joint mixture.

Next: test a separate coherent-count component against literal full-cell
marginalization, then compare its finite-data behavior before any integration.
Keep old predictors and failed rescue artifacts intact. Do not insert the known
parity mask or truth into acquisition. No production or whitepaper changes.

## Verification

Race-enabled diagnostic PASS1.242s package; vet PASS. Deterministic seeds and
fixtures are in the test. Raw:`mmm-count-coherence-v1.json`.

SHA256:

    controller.go 3d742564146015b10c74c4a9c63e0bcf3f499ea53180ebefe634c2fbcf7e3b72
    forecast.go 59d63d232ac772ca6b3b44f35ac549002074a11e88b0a96477b5ed82a5fb05c0
    count_coherence_diagnostic_test.go 5cfde9578ed5e4e015ba0b117107406c10648c7e128fb408e0eb2bf291fa6d13
    raw c5a68be97101604af2d598227e37ef2fe09d0f26316fe88c5373fafc354ab456

All seven research directions remain open; no performance or deployment claim.
