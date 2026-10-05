# Nonlinear partition challenger v29: isolated mathematical preflight

2026-10-02. V28's correctly implemented target-risk objective failed its
fresh external screens. Test a different model family next, not another
observation-only adjustment. No v29 outcome cohort has been collected, and
there is no claim of better recovery, retrieval or calibration yet.

## Declared Joint Family

Inputs are2..200 baseline probabilities strictly inside(0,1) and pre-evidence
coordinates in[0,1]. The coordinate is a declared representation, not a truth
label or authenticated causal variable. A first study may use normalized
frontier position, but that is a benchmark proxy, not semantic agent evidence.

There are two prior branches. With probability .99 use independent member
rates phi_i~Beta(2*b_i,2*(1-b_i)). With probability .01 choose a dyadic
partition of the coordinate domain. Each nonterminal node stops with
probability .5 or splits at its fixed midpoint with probability .5; depth3
must stop. This gives26 possible partitions with normalized prior mass,
not26 equally probable trees.

For each partition leaf, draw psi_leaf~Beta(1,1), independently. Conditional
on psi_leaf draw member rates phi_i~Beta(2*psi_leaf,2*(1-psi_leaf)). Training
and future labels are Bernoulli(phi_i), independent given phi_i. Support one
training label per member. Integrating distinct member rates yields a
Bernoulli(psi_leaf) evidence likelihood; leaf counts have an ordinary Beta
posterior. Tree posterior weights update by the same evidence marginal.

Future means are E[psi_leaf|evidence] for unseen members and
(2*E[psi_leaf|evidence]+y_i)/3 for observed members. Baseline-branch future
means are b_i and (2*b_i+y_i)/3, respectively. Average these under the
posterior branch/partition weights. Evidence and forecasts therefore come
from one joint family, including member heterogeneity.

Cold prediction is .99*b_i+.005, NOT exactly b_i. Do not silently claim
v27's identity property or restore it by an unrelated decoder. The strong
baseline prior is defeasible via likelihood ratios, not immutable authority.
Its .99/.01 weight and the depth/leaf priors are declared before new cohorts;
they are not tuned on v29 outcomes.

## Tests and Scope

The isolated `internal/researchpartition` module passes ordinary and race
tests. Controls verify all26 partitions and their ancestor maps, prior
normalization, input ownership, supported coordinates/probabilities,
duplicate/out-of-range rejection without mutation, and selector exhaustion.
An independent integrated likelihood uses products of Beta-function ratios
for leaf label counts and direct baseline likelihoods. Posterior weights
and future means agree within2e-14 at successive mixed-label histories.
There is no hidden label or truth rate in the model's API.

Three100ms Apple M4 benchmark repetitions:150 future predictions cost
5.166-5.383us, zero allocations; construction1.402-1.422us,6448 bytes and
100 allocations. The allocation count reflects bounded partition enumeration,
not a claim of an optimized runtime constructor. These are component timings,
not loaded service latency, end-to-end freshness or acquisition cost.

Bayesian partition averaging is inspired by
[Chipman, George and McCulloch (1998)](https://www.rob-mcculloch.org/some_papers_and_talks/papers/published/cartfinal.pdf).
This finite fixed-coordinate enumeration is not their stochastic CART search,
not BART, and inherits none of their reported empirical gains. It can express
nonlinear stepwise patterns without inserting the exact consumed sine curve,
but will still fail when important distinctions are absent from the coordinate
or smaller than the declared resolution.

Posterior tree groups remain modeling hypotheses, not Anti-Pigeon certificates
or authority to merge durable groups. Conditional independence, provenance,
nomination bias, real agent outcomes and temporal shifts remain unverified.
Next freeze fresh fixed-label/matched-cost comparisons with the strong
baseline, local Beta and v27 model; retain independent, calibrated, shifted
peaks and sub-bin/coordinate-irrelevant controls. No integration before those
results. All seven whole goals remain OPEN; production and whitepaper untouched.
