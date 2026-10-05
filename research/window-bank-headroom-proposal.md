# Diagnose v97 forecast headroom before another rescue

Status: proposed consumed-data diagnostic, not a new confirmation experiment.
v97 fails4/106 gates, all on late parity-to-majority. The short generic expert
has better late expected Brier (confirmation0.153458 versus generic64's0.170529),
but the bank is worse (0.177947). At step160 the short expert uses only
origins128..159 and receives mean weight0.005532. Its mean weight remains only
0.036797 at step192. This makes selector inertia a credible explanation;
segment averages alone do not prove when a different weight would have helped.

Do not simply boost the short expert: on stationary parity4 its late Brier is
0.149902 versus generic64's0.073149. That intervention has an observed risk.
The failed non-harm intervals also cannot be erased by the improved overall
102/106 count or by the passing lifetime realized-loss certificate.

## Read-only next experiment

Reproduce the two switch directions in both phases from the frozen v97 seed
generator, validate their original records exactly, and additionally record
per32-step block forecasts, bank weights and expected losses. Do not feed the
diagnostic back into fitting, selection, activation or scores. Bind the parent
artifact hash and original source hashes. Label these already consumed streams.

For each issued forecast vector p and simulator probability q, define the
best attainable point in its scalar convex hull:

    p_oracle = clip(q, min_k p_k, max_k p_k).

The expected Brier is (p_oracle-q)^2+q(1-q). This is an optimistic oracle lower
bound on loss for mixing the available forecasts, not an implementable policy.
Compare it with the actual bank and generic64. If this bound offers too little
gain, reweighting alone cannot rescue that block with these experts.

Also compute the best CONSTANT generic64/generic32 mixture in each block:
with d_t=p_short,t-p_long,t,

    alpha = clip(sum_t d_t*(q_t-p_long,t) / sum_t d_t^2, 0, 1),

using alpha=0 if the denominator vanishes. Its utility separates lack of
forecast headroom from the need for per-event oracle selection. It still uses
held-out simulator truth and is not a deployable or validated selection rule.
Unit-test endpoint, equal-forecast and interior cases, plus the ordering
oracle-hull loss <= best fixed-pair loss <= either fixed endpoint loss.

Only after this diagnosis choose the next intervention: recent prequential
selection evidence, explicit regime-aware weight validity, or improved models.
Any rescue must receive fresh design/confirmation data and retain all106
quality gates. Do not infer independence from near-identical expert outputs or
recycle post-change labels to manufacture a retrospective admission certificate.
