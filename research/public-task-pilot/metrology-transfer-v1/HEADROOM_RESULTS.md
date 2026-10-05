# Post-hoc gate feasibility

2026-10-03. Consumed-data diagnostic, NOT a new model or a replacement result.
The original lexical transfer failure and its observed confirmation loss remain
unchanged. The original source/model/raw/protocol are not edited or resampled.

Both DESIGN baselines already score34/36. Their two remaining mistakes occur
in two distinct six-question families. With zero literal losses and at least
two net gains, the ONLY admissible improvement vector is
`(0,1/6,0,0,0,1/6)`: fix both mistakes and introduce no new error. Even an
omniscient36/36 ranking has mean gain0.05555556, SE0.03513642 and the original
one-sided95% lower bound **-0.01524603**. Thus that frozen DESIGN acceptance
gate is impossible on this cohort, irrespective of the reranking technique.
This is a measurement-headroom limitation, not evidence that all retrieval
corrections fail to generalize.

| Cell | Baseline /36 | Maximum net gains | Best admissible lower bound | Oracle gate possible |
| --- | ---: | ---: | ---: | --- |
| Structured design |34|2|-0.01524603|No|
| Text design |34|2|-0.01524603|No|
| Structured confirmation |31|5|0.00823691|Yes|
| Text confirmation |32|4|0.00823691|Yes|

The confirmation best bound uses three evenly distributed gains, not necessarily
every possible correction: a t lower bound is not monotone in each individual
gain. We therefore enumerate EVERY4096 feasible family-count vector, including
allowed paraphrase losses, rather than assuming that a perfect-ranking point
estimate supplies an upper bound on the confidence statistic. A separate
per-question implementation enumerates262144 binary paraphrase assignments
in EACH cell (1048576total) and matches all four maxima. Positive, impossible
and invalid-input controls test the search. Necessary optimistic feasibility
does not construct an implementable forecaster or prove a probability guarantee.

Artifacts: `headroom.json`, `headroom-verified.json`. Input result SHA is retained
and rechecked. Source scripts are `../metrology-headroom.mjs`,
`../metrology-headroom-test.mjs`, `../metrology-headroom-verify.mjs`.

Next: keep the original gate and negative result, and obtain NEW outcome-labeled
tasks with more independent evidence families and adequate baseline headroom.
Do not cherry-pick only known baseline misses, tune on these confirmation labels,
or lower the interval threshold to manufacture a pass. A larger public benchmark
is a promising independent lead; source licensing, corpus/frame conversion,
nomination/packing, no-label serving and full computational cost still require
their own frozen protocol and audit. All seven WHOLE goals remain OPEN/ACTIVE.
