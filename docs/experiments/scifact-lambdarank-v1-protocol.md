# Top-Ten-Aware Rescue

2026-10-04. New frozen rescue after pairrank-v1 FAILED cross-fit recall10 and
NDCG versus BM25. Preserve that negative and evaluator schema-repair evidence.
Use EXACT existing pairrank inputs/source snapshot/labels/folds/feature map,
same200candidate universe, .25tanh residual,200steps/.2rate/.001ridge/4cap.
No new features, optimizer tuning or calibration/confirmation consumption.

Only change the pair weighting. Each iteration ranks each training frontier
with current scores and stable-ID ties. d(rank)=1/log2(rank+1) for rank<=10,
else0. Binary positive/uncited swap magnitude is |d(rank_p)-d(rank_n)|/IDCG10.
Normalize these magnitudes to sum1 within each query, then apply unchanged
equal-family/equal-query weights. The common positive IDCG factor cancels;
therefore the implementation uses raw discount differences and their sum.
Zero-sum/zero-pair cases give no gradient, remain in quality denominators.
Ranks and swap magnitudes are stop-gradient weights, recomputed each step.
This is a LambdaRank-inspired pseudo-gradient, NOT differentiation through
sorting, a convex objective, monotonic NDCG guarantee or faithful LambdaMART.

[Burges2010](https://www.microsoft.com/en-us/research/publication/from-ranknet-to-lambdarank-to-lambdamart-an-overview/)
motivates weighting ranking gradients by metric swap impact. The cause of the
previous harm is not proven: raw overlap features, native nomination bias and
loss mismatch all remain credible. This separate rescue tests loss alignment.
Falsifier remains cross-fit harm versus unchanged common-frontier BM25.

New isolated package/command; original pairrank files/models remain immutable.
Race original+lambda roots x3, discount/brute-force binary swap/fit/cancellation
controls. Independently refit all six lambda models and reconstruct351rankings.
Count dynamic200sorting/training costs and allocations; original failed fit
costs retained. Snapshot/replay, not a live agent/background-learning trial.
All seven WHOLE goals OPEN; production/private corpora/whitepaper unchanged.
