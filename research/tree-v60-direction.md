# V60 prospective direction: member-local escape from terminal pooling

RECOMMENDATION, not a production defect or tested empirical rescue. Preserve
V59's frozen source and rejection. This direction follows two independently
audited mathematical obstructions, not a convenient successful subset.

## Pre-Patch Reasoning

Symptoms: anchoring preserves the initial baseline but broad risk worsens;
predictive acquisition also exceeds its compute cap. Confirmed upstream model
limitations: (1) a fixed conditional-rate interval excludes 27.99375% of tested
probabilities, (2) shared terminal paths impose within-leaf monotone ordering
even when outcomes disagree. Short suffixes, noisy-source mismatch and weighting
remain competing explanations; these two effects are not the sole root cause.
Range and snapshot-order audits are independent views and use hidden rates only
after collection. They must not inform online candidate scores or priors.

Related-fix check: this is a new isolated model, not an upstream software bug.
[Kontoyiannis et al. (2022)](https://rss.onlinelibrary.wiley.com/doi/10.1111/rssb.12511),
sections 2.2/3.1, integrates independent terminal parameters and recursively
averages stop/split marginal likelihoods. We borrow that mechanism, not its
Markov-chain model, coding guarantee, asymptotics or runtime theorem. Our public-
feature tree, paired single-outcome noisy measurements and finite rolling suffix
are a different construction. Identity preservation is methodologically inspired
by [Kull et al. (2017)](https://proceedings.mlr.press/v54/kull17a/kull17a.pdf),
not their empirical calibration result.

Falsifier: independent branches remain unable to represent member disagreement,
break initial baseline preservation, or fail ANY original broad quality,
stationary, recovery, measurement or complete-cost requirement. A correct family
may still be too sample-hungry or expensive. No narrower success, gate relaxation,
new seed after failure, or retrospective confirmation claim.

## Declared Model To Test

Retain V59's shared conditional-rate node family and noise prior. Add an
independent member branch at EVERY terminal leaf, including depth0. Let
G={0,.05,.10,...,1}. The selected preflight alternative uses

```
Z_b ~ beta-binomial(N=20, alpha=b, beta=1-b),
v_b(g) = P(Z_b=20*g),
E[Z_b/20] = b.
```

Compute the finite masses using the adjacent-count recurrence, and independently
verify them by a 20-step Polya-urn count distribution. One prior unit is frozen;
no outcomes select the concentration. The individual prior is .8 mass at p=b
and .2*v_b(g) on grid atoms (22 atoms, including the identity). This finite
beta-binomial PRIOR does not make the noisy evidence posterior a continuous
conjugate beta posterior. All masses are positive and mean exactly b; this is
not fitted reliability, error coverage, or a truth certificate.

The initially proposed exponentially tilted grid also preserves b and support,
but allocates little mass to a low-rate revision of a high baseline. Its source
and complete preflight remain in `tree-v60-identities.mjs/json`. The broader
finite Polya prior is `tree-v60-beta-identities.mjs/json`; it is an explicit
design alternative before a Go cohort freeze, not a tuned confirmation run.

For retained evidence E_i and a SHARED noise value eta, let

```
I_i(eta) = .8 L(E_i | p=b_i,eta)
         + .2 sum_g v_(b_i)(g) L(E_i | p=g,eta).
I_leaf(eta) = product_(i in leaf) I_i(eta).
T_leaf(eta) = .5 L_pool,leaf(eta) + .5 I_leaf(eta).
T_node(eta) = .5 L_pool,node(eta) + .5 T_left(eta) T_right(eta).
```

Empty groups have evidence 1. The posterior noise weights are proportional
to their prior times T_root(eta). Conditional on eta, target forecasts use
the pool/independent posterior branch weights, individual posterior means,
and recursive ancestor weights. Do not independently normalize the noise model
per member or multiply separately noise-marginalized I_i: eta is common.
W1/W2 are two measurements of ONE latent Y; replacing its first factor with
the joint factor does not create two independent Y observations.

Zero-evidence forecasts remain exactly b_i in every component. At the
independent terminal branch all member grid combinations have positive mass,
so arbitrary member disagreement is allowed even at identical public scores.
Nearest-grid approximation has squared probability error at most .025^2=.000625
per member, for any static true probability vector in [0,1]. This is an
oracle representational bound, NOT a guarantee that finite noisy data selects
the good parameters, or that drift/noise assumptions hold. Full prior support
does not imply fast adaptation or valid frequentist error control.

For one FIXED retained evidence set, mixture likelihood is at least the prior
mass times any contained component likelihood. The resulting log-likelihood
penalty includes every branch and individual grid prior cost; it may be large.
It is NOT a prequential trajectory-regret or convergence theorem after window
eviction. Anti-Pigeon retains external target-law authority; posterior branches
are hypotheses, not certificates or automatic graph edits.

## Lifecycle And Cost Requirements

Happy path: issue journals the current forecast, then advances expiry; first
evidence updates one individual and its ancestor pool factors; paired evidence
replaces its original factor; expiry removes it from BOTH paths. Failed updates
must commit neither path. Exact zeros need integer zero-support counts; never
subtract -Inf. Prepare individual plus changed root-to-leaf states in bounded
scratch, normalize, then publish. Keep owner/epoch/time/cap fences and expired-
reply accounting. No production or concurrency claim from a serial prototype.

Individual parameters are bounded by frontier size, not corpus size. Cache each
member's independent log evidence and maintain terminal products without
recounting all members per hypothetical observation. Approximate update cost
O(3*(21*depth+22)); forecast also evaluates member-local means. Measure setup,
evidence changes, nomination integration and all-target prediction value rather
than hiding these terms. Dense prediction acquisition can still exceed its cap;
amortization is a separate exactness-tested lead, not a quality rescue.

Tests before dispatch: baseline/support identities; full joint enumeration of
small two-member pool/individual models under all three noise values; clean vs
observed laws; paired-factor replacement; zero-support revival/eviction; same-
baseline divergent evidence; permutation/ties; failed-operation atomicity;
independent ledger reconstruction and full small-tree pruning enumeration.
Compare all frozen 40 worlds/schedules and unchanged Full/Adaptive controls.
Do not open reserved seeds or sealed agent labels before a justified follow-up.

## Current Status

Both preflights pass 28 initial-mean identities, 1,848 paired mass/marginal
checks, 10,001 nearest-grid checks and 18 full joint comparisons (1,515 latent
states each). Independent count recurrence agrees with Polya-urn dynamic
programming for the broader prior. The recursive terminal mixture agrees with
full joint enumeration, including impossible eta0 branches and revival after
removing contradictory paired evidence. Incorrectly marginalizing eta separately
per member is rejected by a nonvacuous likelihood discrepancy.

Two members initially at .925, with four all-zero/all-one paired outcomes,
forecast (.580937,.852696) under the tilted prior versus (.184612,.930747)
under the finite Polya prior. After twenty such outcomes, the latter forecasts
(.025502,.960241) and pool weight 2.33e-11. These idealized deterministic examples
demonstrate representational and prior-response mechanics ONLY. They are not
stationary non-harm, noisy-generator, adaptive recovery, Anti-Pigeon coverage,
agent utility, performance or equal-total-cost evidence.

The first prototype normalized an impossible noise branch and produced NaN;
failed source and local repair are retained in `tree-v60-preflight/`. A second
test used an absolute likelihood-gap check on small probabilities; it is now
dimensionless, without changing joint-equality tolerance or research gates.

No Go implementation, full cohort, serving integration, confidence coverage or
empirical rescue is claimed here. Next implement BOTH individual and pooled
factor lifecycles with independent reconstruction before full dispatch.
All seven WHOLE research goals remain OPEN/ACTIVE. Production, private corpora,
whitepaper, public publishing and previous source/results remain untouched.
