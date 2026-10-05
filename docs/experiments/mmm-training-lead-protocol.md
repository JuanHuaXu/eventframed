# Effective training lead diagnostic

Consumed full acquisition-to-training artifacts; no new efficacy test. Match all
2688 records to original v120 data. For each paid query and window64/32, locate
the first retained fit whose Origins contains the query origin. Classify exactly
one: no fit in scored horizon, missing-natural label added, timely advance over
natural delivery, or natural label already available by first inclusion.

Natural arrival is origin+Delay, or infinity if Missing. Comparison is inclusive:
arrival<=fit.Clock means no training-time advantage. This counts first inclusion,
not independent samples or effect size. A no-training-lead query may still change
the mixer earlier. Last publication is224; do not invent a256 fit.

Report query counts, first-fit delay from paid reveal, and classes by policy,
window and query-clock modulo32, plus phase/case/schedule cells. No case selection,
no use as a prospective predictor. Independently validate classes sum to query
counts and every included label was revealed by its fit. Full scoring replay.
