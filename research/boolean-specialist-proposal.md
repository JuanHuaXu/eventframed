# Provisional Boolean interaction specialist

Status: research proposal after v89 diagnostics, not implemented or validated.
This is an additional model family, not an oracle rule or replacement for the
retained general learner. Freeze a separate experiment before obtaining results.

## Motivation and source boundaries

O'Donnell's [Analysis of Boolean Functions, Chapter1, Theorem1.5](https://www.cs.cmu.edu/~odonnell/papers/Analysis-of-Boolean-Functions-by-Ryan-ODonnell.pdf)
establishes parity characters as an orthonormal basis under the uniform Boolean
measure. This motivates compact interaction features, not a claim that one
parity rule represents arbitrary Boolean behavior. The model below is our
bounded Bayesian specialist, not an implementation of the book's learning
algorithms or unrestricted Fourier regression.

Blum, Kalai and Wasserman's [Noise-Tolerant Learning, the Parity Problem, and the Statistical Query Model (2003)](https://www.microsoft.com/en-us/research/wp-content/uploads/2016/11/2003-Noise-Tolerant_Learning.pdf)
studies noisy parity learning and its relationship to statistical-query learning.
Its results are not inherited by enumeration in nine dimensions. Enumerating
all subsets is exponential in dimension; fixing512 candidates is the explicit
research boundary, not a general scalable solution to noisy parity.

## Coherent candidate model

For nine binary coordinates and any subset S, including the empty subset, define

$$
h_S(x)=\left(\sum_{j\in S}x_j\right)\bmod 2.
$$

Declare a normalized prior pi_S over all512 subsets, with no favored generator
mask. A frozen sparsity prior is one option. For each candidate, theta is the
probability that an observed outcome agrees with its rule, with a proper
Beta(a,b) prior. Use one joint model:

$$
P_{S,\theta}(X=x,Y=y)
=2^{-9}\theta^{\mathbf 1[y=h_S(x)]}
(1-\theta)^{\mathbf 1[y\ne h_S(x)]}.
$$

For n received eligible audit pairs and agreement count c_S, conditional
likelihood and prediction come from this same model:

$$
w_S=\frac{\pi_S B(a+c_S,b+n-c_S)/B(a,b)}
{\sum_U\pi_U B(a+c_U,b+n-c_U)/B(a,b)},
\qquad \bar\theta_S=\frac{a+c_S}{a+b+n}.
$$

There is no binomial coefficient: these are likelihoods of the actual labeled
sequence, not probabilities of an unordered agreement count. Agreement counts
are sufficient statistics for that likelihood. The full-input predictive law is

$$
P(Y=1\mid x,\mathcal D)
=\sum_S w_S\big[h_S(x)\bar\theta_S+
(1-h_S(x))(1-\bar\theta_S)\big].
$$

For a partial observed mask M, integrate the unobserved coordinates using the
declared input law. Under uniform independent bits, a component contributes
0.5 if S is not contained in M; otherwise evaluate its observed rule. The empty
subset is a learned constant predictor. The uniform partial-input law is not
correct automatically under dependent inputs; retain those mismatch controls.

The construction is a convex predictive mixture of noisy single-rule models,
not a universal Boolean representation. Majority, multiplexers and non-parity
interactions must remain controls, with the existing general learner retained.

## Proposed integration and falsifiers

- Fit only already received eligible audits, under a frozen bounded window and
  the existing refit opportunities. No true rule, change point, target mask,
  scenario name or hidden current outcome is an input to fitting or selection.
- Compile a complete predictive law into the existing conditional forecast
  representation. Keep the original observer for the first test; separate
  representation gains from any later lookahead-policy change.
- Keep the incumbent, generic count/subset and age models. Add or explicitly
  budget the new specialist; do not silently replace an expert or transfer
  certificates from a different predictive law.
- Verify posterior normalization, complement symmetry with a symmetric noise
  prior, variable-permutation equivariance, constant/null controls, valid
  partial marginalization and no future-label use before integration testing.
- Use fresh unseen subsets and coordinate permutations that respect the
  acquisition budget, not only the failed bits1..4 example. Test non-parity
  controls, dependent inputs, fitted-base variation, missing/delayed feedback
  and recurring changes. Preserve failed results and the v88 adoption threshold.
- Measure model fit cost, prediction cost, memory and source/tape replay. No
  theorem, calibrated confidence, causal identification, grokking or production
  readiness claim follows from a specialist pilot pass.
