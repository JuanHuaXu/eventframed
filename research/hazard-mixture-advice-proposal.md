# Uncertain transition rate: prospective lead after v112

Implemented with [component checks and paired benchmarks](hazard-mixture-component-results.md).
Quality remains untested. The original prospective formulation follows.

V112 passes all672 non-harm gates, but fails13
gain gates. It improves confirmation delayed reverse recovery by only.003062
over original log advice, below the fixed.005 requirement. Do not change the
requirement or tune alpha on these now-consumed trajectories.

## Research connection and limits

[Wilson, Nassar and Gold (2010)](https://doi.org/10.1162/NECO_a_00007)
motivate inferring hazard rates rather than fixing them in advance. The publisher
abstract was accessible during this round; full-text retrieval was not. This
proposal is therefore not represented as a reproduction of their hierarchy,
equations, implementation or empirical guarantees.

The executable starting point is the finite Markov prior in Algorithm2 and
Lemma4 of [Mourtada and Maillard (2017)](https://proceedings.mlr.press/v76/mourtada17a/mourtada17a.pdf),
previously checked against literal path enumeration. We propose extending the
latent state with a fixed but unknown transition-rate index. This is our
explicit finite-model construction, not a claim of a new general theorem.

## Fixed model family, not retrospective rate selection

Before any fresh quality run, use the finite rate set:

```text
H = {0, .001} union {2^(-k): k=0,...,8}
P(alpha=.001) = .5
P(alpha=h) = .05 for each of the other ten rates
```

This spans no switching through complete mixing, includes the original rate
and uses a dyadic grid tied to the256-event horizon. It is a modeling choice,
not a discretization accuracy certificate. There is no sweep to select the
best rate on v112 or its confirmation records.

Let the joint latent state be (h,j), with four existing role states j. The
rate index remains constant along a path. At each issued event:

```text
posterior(h,j) proportional to prior(h,j) * emission_i(j)
next(h,k) = (1-h)*posterior(h,k)
            + h*role_prior(k)*sum_j posterior(h,j)
```

Emission_i(j) is the actual issued Bernoulli likelihood when the label is known
and1 otherwise. Normalize the full44-state distribution, not each rate's
four-state filter separately. Independent within-rate normalization would erase
the marginal-likelihood evidence needed to learn rate weights.

Marginalize over h to obtain the four raw-role weights used by the existing
gate and coherent observation controller. No new raw forecast family, neutral
expert, comparison-test allocation or acquisition budget is added. Refilter the
joint messages on delayed labels using the established checkpoint/expiry
contract. Old forecasts and already emitted outputs remain immutable.

This treats the rate as uncertain but constant. It does not yet infer a
time-varying hazard or certify changepoints. Long stationary histories may
concentrate on a sticky rate and still delay recovery; that is a falsifier to
test, not a reason to add an untested second switching layer silently.

## Checks before promotion

- Compare the joint filter with literal enumeration over each fixed rate and
  role sequence, including the rate posterior and predictive marginal.
- A point mass on.001 must recover v112; identical rate-conditioned likelihoods
  must not spuriously change rate weights. Verify missing labels, arrival-order
  equivalence with fixed issued evidence, checkpointing and numerical support.
- Keep likelihood normalizers when comparing rates. Do not count a label both
  in the joint-state likelihood and in a second rate update. Do not refit old
  forecasts under future data or treat fitting labels as held-out evaluation.
- Evaluate both directions of change and all stationary cases on fresh data,
  retaining the full quality criteria and fixed-rate control. Score Brier;
  a log-mixture identity is not a delayed, gated Brier non-harm theorem.
- Measure O(D*H*M) message work and bounded storage, with H=11, M=4 and the
  existing retained-suffix cap. Marginalization leaves foreground model count4,
  but does not make the extra inference work free. No serving-tail claim follows
  from a microbenchmark.

The hypothesis is that uncertainty over switching speed can improve recovery
without unconditional forgetting. V112 does not prove that a wrong fixed rate
caused its remaining failures; model misspecification and acquisition effects
remain live alternatives. A negative fresh result must be retained rather than
repaired by selecting a favorable rate or window afterward.

All seven research directions remain open. No production integration, private
data expansion, commit, push or whitepaper promotion is authorized here.
