# V55 direction: exact original-position query acceleration

RECOMMENDATION, not a production defect or validated rescue. V54's full audit
passed, but its repeated hypothetical-history queries contribute about20.2s
proposal time over120 falsification arms. Quality/recovery failures remain
even if query work becomes cheap. No threshold weakening or scientific tuning
on the consumed diagnostic is proposed.

## Source And Derived Identity

Rabiner1989, [full primary paper, sectionIII and sectionV.A](https://web.mit.edu/6.435/www/Rabiner89.pdf),
uses forward/backward messages for smoothed state probabilities and explains
scaling against numerical underflow. This is ordinary finite-state inference,
not a new statistical guarantee for EventFrame.

Our transition is T(z,z')=(1-h)1[z=z']+h*pi(z'). For one component c, let
alpha_j(z) be the normalized filtered state after the REVEALED factor at j;
beta_j(z) the likelihood of later REVEALED factors given that state. Then
gamma_j(z) is proportional to alpha_j(z)*beta_j(z). A later requested W2
changes ONLY that original-position factor, so

P_c(W2=v | revealed history)
= sum_z gamma_j(z) * f_j(W1,v,z)/f_j(W1,z).

For the open rate atoms and eta in{0,.1,.2}, the single-label denominator is
positive. A component with zero marginal history likelihood has zero global
weight; its conditional query is undefined and must be skipped, not converted
into NaN or an invented confidence certificate. Recombine component queries
with CURRENT global posterior weights, not stale cached weights.

The reset transition's backward update is
beta_j(z)=(1-h)*g(z)+h*sum_z' pi(z')g(z'),
where g(z')=f_{j+1}(z')beta_{j+1}(z'). Normalize beta for numerical scale;
arbitrary shared positive scale cancels when gamma is normalized.

## Tested Preparation, Not Implementation

[Identity script](paired-v55-smoothing-identities.mjs),
[results](paired-v55-smoothing-identities.json):162 exhaustive3-atom path
comparisons across four4/5-position histories, three hazards and three noise
rates pass, including24 zero-support checks and81 unknown-evidence forks.
These are exact finite toy identities, NOT a Go learner or cost experiment.

## Required Implementation And Falsifier

Isolate the optimized learner. Keep the84 components,21 atoms, priors, hazard,
original-position semantics, revealed-history boundary, policies and costs.
Recompute only component-conditional queries; do not skip zero-weight components
in evidence updates, since numerical underflow can later reverse. A bounded
one-query-per-member cache can retain conditional probabilities with member
revision+origin slot. Issue, resolved evidence, cancellation and epoch changes
must fence relevant caches; global weights always recombine afresh. Pending
requests change availability, never supply labels.

Freeze a new complete compiler closure before timing. Verify delayed and future
fork equivalence, score/choice ties, impossible components, fail-closed numeric
faults, cache invalidation and full receipts/snapshots. Retain original V54 raw
outputs. Consumed V54 populations may test EXACT computational equivalence,
not provide fresh confirmation or tune the scientific policy. Run the full
matched diagnostic, not selected favorable cells. Respect8MiB/400ms caps and
measure complete costs, not query-only microseconds. If selection changes beyond
declared numerical equivalence or broad timing fails, preserve that failure.

Even an exact speed rescue does not establish Goal7 superiority: the tiny V54
quality gain over random needs fresh, genuinely total-cost-matched evaluation.
Adaptive recovery, source dependence, untouched task quality, useful splitting
and loaded serving/freshness are not solved by this algebra. All seven OPEN.
