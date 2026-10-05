# V59 conditional anchoring: correct property, failed rescue

## Verdict

The baseline-preservation invariant and isolated model/replay audits PASS.
The scientific rescue FAILS: broad risk, stationary protection, recovery and
downstream utility regress. The prediction-value observer also FAILS its
unchanged 400 ms complete-loop limit in 76/120 cells. No adoption, fresh
confirmation dispatch, or completed whole research goal follows.

[Protocol](mmm-tree-v59-protocol.md),
[all 960 arms](../../research/tree-v59-diagnostic/diagnostic.jsonl),
[independent readback](../../research/tree-v59-diagnostic/readback.json),
[V57 comparison](../../research/tree-v59-diagnostic/comparison.json),
[V58 comparison](../../research/tree-v59-diagnostic/comparison-v58.json),
[source freeze](../../research/tree-v59-diagnostic/freeze.json).

## What Was Tested

The same 40 CONSUMED V54 worlds, 20 regimes, two public-score geometries,
three evidence-delay schedules and eight arms produce 960 arms and 96,000
distinct latent outcomes. Paired measurements share ONE underlying outcome;
48,000 requested second measurements per paired policy are not new independent
outcomes. Five paired policies each request 2,800 measurements per arm;
Full/Adaptive/no-pair have 2,400 and are resource ablations, not equal-cost
observation controls. Diagnostic n1 per cell, no population confidence intervals.

The new conditional tree retains each member's public baseline in the initial
forecast and an identity hypothesis with prior mass .8. Twenty centered logit
offset hypotheses each have mass .01. Three shared noise hypotheses, depth 7,
255 nodes and a suffix of 600 ISSUED positions are frozen before dispatch.
Member-specific factors replace absolute-rate pooling; no hidden clean rates,
future outcomes, source availability or selected second outcomes enter learning
or acquisition. Original-position replacement and expiry semantics are retained.

[Kull et al. (2017)](https://proceedings.mlr.press/v54/kull17a/kull17a.pdf)
motivates retaining an identity calibration possibility. The centered finite
prior and paired-noise conditional tree are OUR construction, not that paper's
MLE procedure, exact beta family, or empirical guarantee. A preserved initial
baseline is not a post-update non-harm theorem.

## Complete Results

| Policy | Issued Brier | Terminal Brier | Top-10 utility | Total loop (s) | Maximum loop (ms) |
| --- | ---: | ---: | ---: | ---: | ---: |
| Full control | .224699096 | .208151654 | .727150389 | 7.708 | 80.322 |
| Adaptive control | .214747855 | .181598705 | .815037852 | 31.864 | 319.358 |
| Anchored no-pair | .251824449 | .241864167 | .632853380 | 5.892 | 50.752 |
| Random | .249004603 | .237745273 | .649973677 | 7.308 | 63.161 |
| Uncertainty | .248513679 | .237136475 | .658473677 | 13.417 | 114.469 |
| Noise-class information | .248501434 | .237199825 | .653973677 | 13.439 | 114.939 |
| Noise-class concentration | .248484899 | .237274430 | .657973677 | 13.471 | 131.929 |
| Prediction value | .248659310 | .237445664 | .655973677 | 46.364 | 436.544 |

Every new learner policy wins only 3 and loses 117 cells versus V58. Prediction
value increases issued risk by .008547880 and terminal risk by .004312005,
while utility falls .101103185. Against V57, its issued/terminal risk increases
.030876508/.049013317 and utility falls .154835986 (15 wins, 105 losses).
Anchoring fails the declared empirical falsifier; the correct mathematical
property does not establish a useful model.

All five paired policies have 105/120 cells with >.01 Adaptive harm, and their
Full-relative mean risk gains are NEGATIVE (.0238-.0243 loss), not the required
.01 gain. Recovery is 2.383333 rounds slower than Adaptive over 60 shift cells,
with no improvements over new no-pair. On 60 stationary cells prediction-value
risk is .235630121 versus Adaptive .199637762; maximum harm is .078075815.
Its mean risks by delay are .243257787 immediate, .252928424 fixed150 and
.249791717 uniform299. No subsets or failed outcomes were dropped.

Concentration improves internally over random by .000519704 and uncertainty
by .000028780, but costs 13.471 s versus 7.308/13.417 s. Prediction value is
worse than uncertainty by .000145631 and costs 46.364 s. Equal requested
measurements do not mean equal TOTAL cost; Goal 7 has not passed.

## Two Post-Collection Capacity Findings

[Range audit](../../research/tree-v59-diagnostic/capacity.json) checks 1,728,000
issued forecasts against their declared family. Every clean forecast is a
convex mixture of fixed member rates, hence lies in that member's rate interval.
26,874/96,000 underlying probabilities (27.99375%) lie outside these intervals.
Bernoulli expected Brier decomposes into variance plus squared prediction error;
the family must incur at least .010307287 excess risk on average. Irreducible
variance .160094456 plus this relaxed floor is .170401743. This is not an
attainable oracle: shared-parameter and tree restrictions are relaxed.

[Order audit](../../research/tree-v59-diagnostic/order-capacity.json) proves a
different restriction. All conditional atoms increase with the public baseline.
Members in the same terminal leaf share their mixture coefficients, so their
SIMULTANEOUS predictions are ordered by baseline; identical baselines give
identical forecasts. All 12,000 snapshots pass 276,000 within-leaf pair checks,
including 12,000 equality checks. Hidden rates are used ONLY by this offline
auditor. An isotonic oracle gives .006054597 mean excess and .006021784 terminal
excess; the tight independent case alone has .0132 excess. These are relaxed
lower bounds, not empirical rescue results or new adoption gates.

The two floors overlap and CANNOT be summed. Snapshot order does not constrain
sequential issued predictions with different histories. Neither floor uniquely
explains all observed loss; short retained history, noise/source mismatch and
model weighting remain separate competing causes.

## Cost, Accounting And Audit

Constructor allocation is 1,818,256 bytes, below 8 MiB, not RSS. Cached Predict
takes 208.8-210.0 ns with zero allocation; 150-origin concentration takes
3.244-3.253 ms with zero allocation; all-target prediction value takes
15.997-16.474 ms and 12,040 bytes/12 allocations. The predictive complete-loop
maximum is 436.544 ms, with 76 failures of 400 ms. These are not loaded serving
latency, freshness, persistence or agent benchmarks. Host variation and changes
to the law prevent causal speed claims from cross-run timing ratios.

All paired policies request 48,000 second measurements. Expired requested /
available replies: random 2,154/2,114; uncertainty 2,004/1,967; information
1,986/1,950; concentration 1,966/1,929; prediction value 2,016/1,980. They remain
charged and acknowledged without learning. Missing replies are separate:
1,053/978/980/992/981 respectively. No denominator silently excludes expiry.

Seven model roots and three fixture roots race-pass; separate vet, allocation,
closure and microbench jobs pass. Independent small-tree enumeration covers
all five depth-2 prunings. Delayed suffix reconstruction compares 4,656 forecasts
and 120 values across depths 0/2/3/7 and windows 2/9/64. The fixture includes
18 nonvacuous future forks, 52 corruption rejections and 3,840 distinct seeds /
23,040 domain-separated channels. All 960 arms replay independently, including
288,000 prediction values, laws, noise weights, receipts, choices and snapshots.
The reference recounts retained member ledgers, uses a separate moment solver
and explicitly sums the single latent outcome; no candidate inference is reused.

All 240 Full/Adaptive controls are bitwise unchanged excluding cost versus
BOTH V57 and V58. All populations are identical. 77 actual compiler inputs plus
12 support files are prospectively hashed/copied; all eight commands terminate
0. The 14 pre-existing dirty tracked files and parent archived copies remain
unchanged. Production, private data, sealed task labels, whitepaper and publishing
are untouched. No commit, push, install or deployment.

First receipt Forecast is the CLEAN-outcome forecast; its Value is noisy W1.
ObservedIssued separately records the W1 law. The study scores analytic clean
expected risk, not the noisy receipt as if it were ground truth. Second-receipt
forecasts target W2. Future adapters must preserve this distinction explicitly.

## Next Lead

[V60 member-local escape](../../research/tree-v60-direction.md) adds a positive-
mass independent member branch beneath each terminal group, with a baseline-
preserving prior on a full [0,1] grid. Widening V59 offsets alone would leave
its mandatory within-leaf ordering intact. The proposed branch removes both
structural obstructions while retaining shared-evidence alternatives, but does
not prove practical learning, valid Anti-Pigeon authority, recovery, or useful
equal-total-cost observation. All seven WHOLE goals remain OPEN/ACTIVE.

Local workflow correction: large diagnostic JSONs must be read through explicit
selected fields or streaming summaries. A broad read here was truncated; it was
replaced with bounded readback, without changing evidence or global instructions.
