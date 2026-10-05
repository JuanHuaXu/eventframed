# Noise-channel challenger V53

Research only; all seven full criteria unchanged. V52 was PROGRESS but did not
rescue learning. Confirmed model mismatch: existing noise10 generators flip
observations while scoring true usefulness; the moment learner assumes clean
Bernoulli observations. This is not a software bug in its declared working
model. Other credible causes are stale shared families and scarce local samples.
This new isolated trial tests the measurement-model hypothesis only.

## Joint Model

Retain the21-rate-atom,28-shared-family, per-member Markov reset model, moment
prior strength2, hazard1/16 and memoized initialization. For latent clean
outcome Y, observed label W and rate p, declare

$$P(Y=1\mid p)=p,\quad P(W\ne Y\mid Y,\eta)=\eta,$$
$$P(W=1\mid p,\eta)=\eta+(1-2\eta)p.$$

The likelihood uses W; the scored clean forecast marginalizes Y. Report the
observed-label predictive separately. No inverse clipping of an empirical
frequency, ground-truth rate, noise-case flag or future label enters prediction.
Fixed variants eta0/eta10/eta20 use0/.1/.2 in EVERY case. learned_eta has one
shared latent eta in{0,.1,.2}, prior(.8,.1,.1), and posterior proportional to
the FULL joint marginal evidence, not repeated one-step predictive factors.
Its observed forecast is sum_eta w_eta[eta+(1-2eta)q_eta], NOT a product of
averaged eta and averaged q. Rate/family/noise parameters can be confounded;
no universal identifiability or evidence authentication is claimed.

## Frozen Evaluation

Fresh diagnostic base2026105307, design2026105309 and confirmation2026105311
reserved. Diagnostic n1:ALL original14regimes x2geometries x3arrival schedules,
plus stationary/partial IIDnoise20, stationary correlated round-noise10 and
unrelated IIDnoise10.18regimes/36worlds/86400distinct labels. Correlated noise
deliberately violates conditional-independence assumptions; retain its failures.
Separate fixed/learned noise effects. Same-world Full, Adaptive and original
rich_moment2 controls, ALL2400nominations per world/schedule, no pruning.

Original score/protection/recovery gates remain: >=.01 Full issued-risk gain
with positive lower confidence bound, <=.01 Adaptive harm, >=10% shifted recovery
gain with positive lower bound, constructor8MiB and complete-loop400ms. n1 cannot
establish intervals/confirmation/adoption. Report all risks, priority, final
top10 usefulness, recovery, per-cell wins/losses, ALL costs and failed gates.
This is challenger validation, not yet an integrated MMM, real-agent or loaded
serving/freshness result. Integration follows only a promising broad screen.

## Required Invariants And Falsifier

eta0 exactly matches original non-cost predictions/receipts/state behavior.
Independent full-history reference and explicit latent paths verify likelihood,
clean/observed forecasts and marginal-evidence/model-comparison coherence.
Original-position delayed replay, ticket ownership/epoch/caps/cancellation,
finite weights and whole-bundle fencing must pass. Future labels cannot change
pre-reveal outputs; contrary revealed forks must change later outputs.
Corruption controls reject changed forecasts, noise weights, labels, identities,
coverage and cost. Derive ACTUAL compiler/test dependency closure BEFORE freeze;
record Go runtime, source copies and command logs. Preserve failures/new roots.

Falsifier: correcting the likelihood does not improve broad recovery/protection
and usefulness at acceptable total model cost, even if a noisy cell improves.
Negative results remain evidence; no retuning of consumed confirmation seeds.

## Source

[Patrini et al., CVPR2017](https://openaccess.thecvf.com/content_cvpr_2017/html/Patrini_Making_Deep_Neural_CVPR_2017_paper.html)
motivates forward transition-matrix correction. Our finite hierarchical model
and discrete noise-prior comparison are adaptations, not the paper's neural
training algorithm, noise estimator or an inherited guarantee.
