# Independent repeated-measurement lead (V54, not yet an experiment)

V53 fails broadly and wrongly infers nonzero noise in some clean streams.
Do not rescue by passing true noise rates or case identities to the predictor.
The next lead changes the evidence contract rather than retuning consumed cases.

For ONE latent outcome Y~Bernoulli(p), obtain W1 and optionally W2 from distinct
conditionally independent measurement processes with symmetric error eta<.5.
The second measurement is NOT another independent latent outcome and is not a
repeated conversation, duplicated testimony or unauthenticated source vote.

Define a_y(w) = (1-eta) if w=y, else eta. The joint likelihood is

P(W1=u,W2=v|p,eta) = p*a_1(u)*a_1(v) + (1-p)*a_0(u)*a_0(v).

Its disagreement probability is d=2eta(1-eta), independent of p. Under the
declared assumptions, eta=(1-sqrt(1-2d))/2. Finite evidence requires uncertainty
bounds, not substituting an observed disagreement fraction as exact truth.
Heterogeneous/source-correlated noise invalidates this identification shortcut.

An ordinary single label has likelihood eta+(1-2eta)p. Thus p=.90,eta=.10 and
p=.82,eta=0 are observationally identical under single independent labels, but
their paired-measurement disagreement probabilities differ(.18 vs0). This is
an identification resource, not evidence that a deployed agent has access to it.

The first measurement may arrive alone, and W2 may arrive later. Replace its
original-position factor with the JOINT likelihood; equivalently multiply by
P(W2|W1,p,eta). Never multiply two marginalized single-label factors, which
incorrectly resamples Y. Unknown second measurements integrate to1. Disagreeing
pairs have ZERO likelihood for eta0; implementations must distinguish impossible
models (zero posterior mass) from numerical faults (NaN) and from all-model
failure. V53's always-positive single-emission model does not implement this.

Before implementation/evaluation, freeze a new protocol with fresh seeds, full
clean/noisy/correlated/shift/delay coverage, and equal TOTAL label acquisitions
for random, uncertainty and falsification-driven requests. Include proposal
scoring, delayed joint-factor replay, source independence audits and runtime
cost. Never select an observer using unrevealed second labels or latent rates.
Do not waive Adaptive protection, useful split outcomes, untouched agent tasks
or loaded freshness gates. Conditional independence must be a tested contract
boundary, not silently assumed from different source names.

`node research/noise-v54-identities.mjs` checks only the likelihood identities,
stagewise update equivalence, single-label nonidentifiability and paired
distinguishability. It is NOT a learned implementation, acquisition-policy
experiment, valid real-world noise certificate or completion of any goal.
