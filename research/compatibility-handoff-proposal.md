# Forecast-compatibility handoff: prospective rescue

Subsequently tested in [v111](../docs/experiments/mmm-compatibility-v111-results.md):
both variants FAIL the complete fresh quality screen. Preserve the original
proposal below and the successful component checks; neither is a quality pass.

Implemented as a research component with [unit/lifecycle checks and paired
benchmarks](compatibility-handoff-component-results.md). The original
prospective formulation follows, motivated by the consumed
[v110 trace](../docs/experiments/mmm-handoff-v110-results.md), which reproduces
late short-window weighting lag even at matched views. Preserve v109's complete
failure and v108's rejection of unconditional age discounting.

## Defined geometry, provisional modulation

Let p_j^old(x) and p_j^new(x) be successive full-support Bernoulli forecast laws
for one fixed advice role, under a declared common input measure mu. In the
current fixture mu is uniform over512 inputs. Define their joint-law affinity:

```text
a_j = E_mu[ sqrt(p_old p_new) + sqrt((1-p_old)(1-p_new)) ]
gamma_j = a_j^32
```

The exponent32 is one predeclared publication interval, not selected by a
quality sweep. Affinity lies in[0,1]; identical laws give1. For genuinely
independent product draws, affinity tensorizes, which motivates the scale.
Actual overlapping training windows and delayed evidence are not independent
draws from this law. Consequently gamma is a compatibility modulation, NOT a
posterior probability of validity, reuse certificate or effective sample size.
Use the declared model input law, not simulator outcomes. An estimated or
misspecified input measure would require its own validation in real deployment.

[Van Erven and Harremoes, Renyi Divergence and Kullback-Leibler Divergence,
2014](https://arxiv.org/pdf/1206.2459), Definition2 and Theorem28, provide the
order-one-half divergence and product additivity. With affinity a,
D_(1/2)=-2log(a); product additivity yields a^n. They do not establish the
advice-transfer rule below or its Brier performance.

## Transfer normalized evidence, not arbitrary log offsets

For positive-prior roles with current normalized weights w and prior pi:

```text
e_j = log(w_j / pi_j)
new_w_j proportional to pi_j exp(gamma_j e_j)
```

Use normalized weights in e: scaling stored log weights by role-specific
gamma without subtracting their normalization makes behavior depend on an
arbitrary common log offset. Test that gauge invariance explicitly.
All gamma1 must preserve the existing policy; all gamma0 returns its prior.
Zero-prior roles remain excluded. The current log/no-neutral policy is the
primary base, so this does not add the harmful explicit neutral candidate.
An unchanged forecast preserves its own evidence term, although normalization
may change its relative weight when other roles change.

This is a generalized evidence-transfer heuristic, not ordinary Bayesian
conditioning or a guarantee of better calibration. A changing but correct
model could lose useful historical weight; stationary and sparse-feedback
protections are mandatory.

## Pending losses must follow the same handoff

Record per-role log compatibility at each publication. For a loss issued under
version v and arriving under u, its transfer coefficient is the product of
gamma over boundaries v+1 through u. Compute this in log space with the bounded
eight-version table, preserving exact identity at no movement.

Role-specific attenuation of raw log likelihood alone would unfairly reward
roles whose old evidence is ignored: log(1)=0 beats every negative log score.
Use the fixed neutral reference for discounted likelihood ratios instead:

```text
log_weight_j += transfer_j * [log P_issue,j(y) - log(.5)]
normalize; apply the unchanged .001 share step
```

The reference .5 is not a new competing expert. At transfer1 for every role,
the common reference cancels and the original log update is recovered. At0
there is no score contribution for that role; the declared share transition
still occurs. The fixed reference is part of this generalized update contract,
not an interchangeable loss constant. No issue is scored under a future model.

Compare publication handoff alone against handoff plus pending-loss transfer
to isolate their effects; neither may claim a certificate. Keep the original
log and Brier controls. Test products across versions, late batches, duplicates,
expiry, reentrancy, finite probabilities, gauge invariance, identity limits and
coherent observation/served law. Preserve the existing gate and its budget.

Compute compatibility only on publication (four roles times512 full inputs in
this fixture), cache version coefficients, and keep per-arrival work bounded
by the five-role cap. Measure the extra publication work and entire lifecycle.
Do not extend to arbitrary role/schema changes by silently carrying this state.

Freeze fresh seeds and the full existing protection/recovery criteria before
quality evaluation. No raw affinity threshold, exponent or prior sweep on v110
outcomes, and no promotion from a passed component or a matched-mask oracle.
