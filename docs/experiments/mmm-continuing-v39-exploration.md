# Continuing Local Rate Exploration V39

2026-10-03. Research prototype only, not adopted or empirically validated.
All seven whole goals OPEN. This addresses goals1/2/4 alongside unresolved
loaded serving and untouched-task retrieval. No whitepaper/production changes.

## Why This Lead

V38's global two-state reversal mixture could not represent partial/unrelated
changes, and four-label frozen templates lost stationary learning. Those are
structural restrictions, not evidence that a threshold needs retuning. Competing
causes include scant labels, template bias, global coupling and misspecified
transition dynamics. Replace point-template reversal with a finite latent rate
distribution PER MEMBER, updated throughout the evidence stream. This grants
no sharing, causal, authenticity or Anti-Pigeon certificate authority.

## Declared Joint Model

21 rates q_z=(z+1/2)/21. Initial member prior pi_i(z) is normalized discrete
Beta-shaped mass q_z^(4b_i-1)(1-q_z)^(4(1-b_i)-1), using only the supplied
baseline b_i. This is a DISCRETE prior, not exact continuous Beta quadrature.
T_i(z'|z)=(1-h)1[z'=z]+h*pi_i(z'). Each member's latent state advances once
per actual nomination, not global tick or observation arrival. Conditional
labels Y_ij are Bernoulli(q_Zij); the same joint model defines evidence
likelihood and next-outcome forecast. Exact forward messages for this declared
family integrate rate uncertainty, not just a plug-in estimate.

Before issue j, predict sum_z q_z*T_i f_(j-1), using the initial prior for j=0.
Unarrived/cancelled labels have unit emission. A late label is inserted at its
ORIGINAL member nomination position and only that member's suffix is replayed.
Forecasts already issued remain privately bound to the existing opaque ledger
ticket/epoch; today’s posterior never replaces the scored issued law. Inputs
contain no evaluator true rate/regime or future outcome. Pending/history and
monotone as-of boundaries retain existing DelayedShape semantics.

Single-owner research learner,<=200members/64nominations each,21atoms. Forecast
O(21), suffix update O(64*21), storage O(members*64*21) plus inherited ledger.
The constructor includes unused inherited ShapeModel allocation; measurements
must count it, not claim an optimal or constant-cost global learner. No general
convergence/stationary protection or cross-member sample-efficiency guarantee.

## Primary Research And Limits

[Adams and MacKay, Bayesian Online Changepoint Detection](https://arxiv.org/pdf/0710.3742):
equations1-5 integrate predictive distributions over latent runs and specify a
constant geometric hazard. This prototype instead marginalizes a finite rate
chain; it is NOT the paper's run-length algorithm and inherits no empirical
guarantee. The reset law/independent local chains are explicit model choices.
[Mastrototaro and Olsson, Online VSMC](https://proceedings.mlr.press/v235/mastrototaro24a.html)
motivates continuing parameter adaptation; this prototype uses neither particles
nor variational gradients and inherits none of their convergence results.

## Checks Before Outcome Experiments

Independent21^3latent-path enumeration for partial and out-of-order arrivals;
hazard0 batch likelihood equivalence; member isolation; future-prefix issued
law identity; spoofed forecast ignored, foreign/replayed/premature/cancelled/
old-epoch rejection; capped pending/history and atomic failed updates. Race,
full related package tests, vet and forecast150/late-suffix64 benchmarks.
These are mathematical/component checks, NOT a prediction-quality confirmation.

Next freeze NEW broad cohorts and compare full/adaptive/anchor/orientation
controls at equal labels, with stationary harm, partial/recurring/late/asynchronous
regimes, delayed evidence, priority/packet quality and cost gates. h=0/1/32/1/16
are predeclared candidate controls for exploration; do not select a winner on
confirmation then retroactively certify it. Actual frozen experiment protocol
must exist before new outcomes are generated. Real untouched agent tasks remain
separate requirements, not replaceable by this finite synthetic family.
