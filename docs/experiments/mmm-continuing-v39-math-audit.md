# V39 Mathematical Boundary Audit

2026-10-03. Component proof/check notes, not a performance or quality result.
This does not amend the frozen experiment or grant deployment authority.

## One Coherent Family

For member i and nomination ordinal j, let Z_ij have21 states, q_z=(z+.5)/21.
Initial mass pi_i is the declared normalized discrete Beta-shaped prior.
Transition T_i(z'|z)=(1-h)1[z'=z]+h*pi_i(z'), with0<=h<1; each row sums1.
Given the path, Y_ij has Bernoulli(q_Zij) law. Evidence and next-label prediction
are marginals of this SAME joint family; no unrelated predictive kernel attached
to a fitted posterior. Independent member paths are a MODEL assumption.

At an as-of boundary, arrived label j has emission g_j(z)=q_z^y(1-q_z)^(1-y).
Unarrived/cancelled j has emission1. The forward recurrence is

```
f_0(z) = normalize(pi_i(z) g_0(z))
f_j(z') = normalize(g_j(z') sum_z T_i(z'|z) f_(j-1)(z))
q_next = sum_z' q_z' sum_z T_i(z'|z) f_(n-1)(z)
```

For n=0 use the initial prior predictive, NOT an extra hidden nomination.
Calling Predict repeatedly does not advance latent time; successful Issue
does. Missing-emission steps still advance latent time once per nomination.
For a label arriving late at j, f_(j-1) is unchanged because it is a filtered
prefix, not a smoothed posterior conditional on future positions. Recompute
j..n-1 with all CURRENTLY arrived emissions; the final filtered law equals
the complete as-of batch product. Already-issued q_j is an immutable PRIVATE
ledger fact and never replaced by the recomputed q_next. Cancel is not Y=0.

All q_z lie strictly between0 and1 and pi_i(z)>0 on the accepted baseline
range; finite normalization sums remain positive for the capped64-step family.
This is not a theorem about arbitrary floating-point horizons or other priors.
Independent21^3path enumeration, hazard0 batch likelihood, and a separate
log-prior/unnormalized batch recurrence are evidence for the implemented rules.

## Assumptions That Cannot Be Smuggled Into Claims

- Unit emissions require non-informative missingness/cancellation for this
  likelihood. Uniform/fixed experimental delays meet that design assumption.
  Outcome-dependent delays or selective cancellation need an explicit model.
- The noise10 fixtures deliberately misspecify the Bernoulli evidence model
  relative to true usefulness. Better numerical filtering does not correct
  an unknown label-corruption mechanism or authenticate sources.
- Nomination is externally fixed here. This is NOT selection-conditioned
  inference for the daemon's learned/retrieved frontier.
- Independent local chains avoid global coupling but cannot automatically
  borrow useful information from compatible members. Sample efficiency is
  a measured outcome, not guaranteed by the prior or latent-state arithmetic.
- The geometric reset returns to the initial baseline-shaped prior, not a
  learned recurring-template library. It can retain bad baseline influence.
- Member evidence/history is capped64; epoch reset explicitly discards it.
  Persistent lifelong learning and useful retained abstractions are separate.
- No fast splitting, valid Anti-Pigeon error control, source independence,
  causal effect, neural grokking, or agent-answer improvement follows.

## Cost And Ownership

Forecast O(21); new nomination includes inherited ledger Predict work and one
21-state row. Late resolution replays only one suffix, O(64*21), with private
scratch rows before publication. Space O(M*64*21) plus inherited ledger/Shape
allocation, for M<=200. No database/network/embedding opens in this model.
It is single-owner research code; race-tool success is NOT concurrency safety
for sharing one learner across goroutines. Constructor bytes, full collector
elapsed and separate worst-suffix benchmarks must be measured, not derived
from asymptotic notation. Serving queues/storage remain outside these numbers.

Related primary inspiration: [Adams and MacKay](https://arxiv.org/pdf/0710.3742)
marginalize run-length uncertainty and use a geometric hazard; our direct
finite-rate chain is a distinct bounded model, not their algorithm. It does
not inherit their empirical results or a general convergence theorem.
