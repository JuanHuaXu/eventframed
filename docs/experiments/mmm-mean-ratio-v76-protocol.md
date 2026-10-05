# V76 exact at-anchor pair replacement preflight

Isolated proposal, not production adoption or a scientific-quality rescue.
Freeze before tests; dispatch after V75 timed collection is terminal.

## Patch Reasoning Gate

Confirmed repeated work: V75 replays known history even when only a second
measurement at the anchor is added. An at-anchor first-only conditional filter
already includes the first factor. Likelihood-ratio replacement may remove that
replay exactly. It does not remove global hyperstate marginalization, all-target
predictive work, numeric guards or older-than-anchor replay. Those are alternative
cost sources; one favorable microbenchmark cannot establish complete-core cost.
No upstream patch applies to this unpublished research model.

For fixed hyperstate and anchored rate atom, let L1 be the original first
likelihood and L12 the joint pair likelihood for ONE latent outcome. Replace
the conditional filter p by normalized `p * L12/L1`, and add the log normalizer
to its existing evidence log. This is not a product of independent outcomes.
An old zero likelihood must have zero filter mass; positive mass there is an
error, not grounds to silently skip it. Existing unsupported hyperstates stay
unsupported. The new pair cannot change the first value. Point rates remain
implicit, current/free grids and full27 means remain unchanged.

Apply only when the first is revealed, second not yet observed, and the trial
ordinal equals the current anchor. Older pairs retain full replay; newer first
reveals retain incremental filtering; epoch/cancellation/publication clocks
remain unchanged. Prepare every dependent belief/cache/odds object before any
row publication. Hypothetical queries must not publish.

Falsifiers: disagreement with independent dense full-joint reference, different
actual branch laws, failed tower identity, revived impossible noise class,
NaN/0/0 handling, changed shared/local/individual semantics, partial publication,
future-data dependence, or incorrect old-row fallback. Tests must cover both
second outcomes at the anchor and older rows with unknown suffixes, hazards
0/1/16/1, full64 journals, configuration limits and malformed support.
Use the existing 2e-10/2e-11 tolerances and original constructor cap (8 MiB at
both150/200). Run unit/race/vet before serial matched-position benchmarks.

This preflight does not establish full-cohort equivalence, <=400 ms complete
core, useful valid splitting, fresh labeled agent utility, durable freshness,
loaded serving or equal-total-cost observation superiority. All seven whole
goals remain active. Keep private/sealed labels, reserved confirmation seeds,
production, paper and publication untouched. Preserve every failed attempt.

## Resolved-Symbol Identity

For a fixed member/hyperstate, write D for all its currently revealed factors,
a for its latest revealed ordinal, and `p_a(z) = P(R_a=z | D, hyperstate)`.
The first observation at a is part of D; its realized value is w1. The stored
filter is at a even when `count > a+1`; it is NOT the next-rate prediction cache.
Let `g(z)=P(W2=w2 | W1=w1,R_a=z,hyperstate)=L12(z)/L1(z)` wherever L1>0.
Then `s=sum_z p_a(z) g(z)` is the conditional second-observation probability,
`p_a^+(z)=p_a(z)g(z)/s`, and the member evidence normalizer changes by s.
Consequently its log evidence changes by `log(s)`, and the global hyperstate
weights must be rebuilt using that updated member evidence. Local-noise and
independent-member alternatives marginalize exactly as before.

For L1=0 the first-only filter has zero mass at z. The implementation validates
this before avoiding 0/0. If a hyperstate was already unsupported, no conditional
update can restore its evidence support. If s=0, the new pair makes that
hyperstate unsupported. The positive-noise alternatives keep the full declared
model supported even when a zero-noise pair disagrees. Only after replacement
is the anchored filter propagated through `count-a` transitions for the next
unissued outcome. Arrival itself is not an additional issue-clock transition.

If later revealed evidence exists, the target ordinal is not a and a local ratio
on the stored latest filter would be wrong: the pair affects a smoothed older
rate and the later filter through transitions. That case retains journal replay.
These identities motivate the implementation but do not replace its independent
reference, numerical, branch and lifecycle tests.
