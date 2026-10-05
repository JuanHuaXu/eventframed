# Dependence diagnosis and two-order rescue

## Original-likelihood diagnosis

Frozen `mmm-spike-two-feature-contract.md`: enumerate four inclusion states
and integrate the original logistic likelihood for intercept plus up to two
active coefficients. Same pi=1/255 and unit Gaussian priors as the screened
model. Midpoint128/192 resolutions agree within2e-7; symmetry and independent-
feature negative controls pass. This is numerical reference integration,
not a formal uniform quadrature certificate or a255-feature posterior audit.

| n | Observed design | Reference discordant-query P | Forward VI P | Reverse VI P | Equal average P |
| ---: | --- | ---: | ---: | ---: | ---: |
| 16 | identical features | .500000 | .786317 | .213683 | .500000 |
| 64 | identical features | .500000 | .951553 | .048447 | .500000 |
| 16 | balanced independent | .790825 | .789374 | .789374 | .789374 |
| 64 | balanced independent | .950336 | .951601 | .951601 | .951601 |

For the identical design, the query asks for feature1=true and feature2=false,
which was absent from training. Exchange symmetry makes the reference answer
.5. Each coordinate ordering selects a different plausible explanation.
At n64 reference posterior state masses are approximately
[null .000000, only1 .484943, only2 .484943, both .030114]. The result confirms
a factorization/mode-selection failure, not a mathematical identification of
which feature is causal. More repeated identical evidence does not resolve it.

Equal-order averaging repairs this toy symmetry, not the complete posterior.
Maximum predictive errors over all four queries drop from.286317 to.037382
at n16, and from.451553 to.000677 at n64. The independent-design errors remain
.001451/.001265. In particular, n16 diagonal-query posterior mass is not
correctly recovered by averaging. Do not call this exact Bayesian averaging.

## Frozen84-fit rescue pilot

`mmm-spike-order-ensemble-contract.md`: reverse all255 masks, retain the same
prior/data/initialization/cap/integration, and average ascending and descending
forecasts with fixed equal weights. Reuse audited ascending records; no
evaluator-dependent weight or order selection. All84 reverse fits converge
in5-44cycles. Collection12.65s/replay12.53s, full output byte-identical.

Expected Brier:

| Phase / schedule | Ascending | Descending | Ensemble | Markov |
| --- | ---: | ---: | ---: | ---: |
| 0 / immediate | .187476343 | .187476341 | .187476342 | .178898229 |
| 0 / delayed | .188412738 | .188412741 | .188412740 | .179315442 |
| 1 / immediate | .182601775 | .182601772 | .182601773 | .176164100 |
| 1 / delayed | .185928413 | .185667129 | .185740370 | .181535407 |

Only phase1/majority3/delayed changes in mean expected Brier by more than1e-6
between orders: .0530191 -> .0475321, ensemble .0490702; Markov .0477122.
Largest individual probability difference across orders is.0944914. The
ensemble remains worse than Markov in every pooled expected-score cell.
This is a narrow toy rescue, NOT a successful general rescue, and does not
justify another broad unchanged-model run or production adoption.

## Verification and cost

Reverse-mask source/moment audit passes: maxmeanerror8.89e-16,
maxvarianceerror2.23e-16, max2,337 integration evaluations, no range clamps.
Both-order as-of/evaluator poisoning, admitted-label positive controls and
wrapper preservation tests pass. Full spike race suite passes37.779s.

Apple M4, three repetitions of three operations, pilot-prior fixture:

| Operation | ns/op | B/op | allocations/op |
| --- | --- | ---: | ---: |
| Two predictions and averaging | 7269736 / 7216111 / 7231972 | 16384 | 8 |
| Two complete fits | 19454264 / 19628597 / 19564028 | ~9565525 | 1108 |

Command: `go test ./internal/observationlearners -run '^$' -bench '^BenchmarkSpikeTwoOrder' -benchtime=3x -count=3 -benchmem`.
Prediction excludes fitting; allocation totals are not peak RSS. No loaded
serving-latency claim. Roughly doubled computation buys almost no pilot-wide
improvement, so do not make this the default.

## Artifacts and next step

Reverse/replay SHA256:
`9ac36e85ee8f23cf1cccc1b472440ba36ae97f601f96fb689524530793dbd46f`.
Audit: `mmm-spike-reverse-v1-pilot-audit.json`, SHA256
`d41c9e16e967c1f6180e35053b2ea23b839a0c505a3fafc012ca784647fd99f5`.
Summary: `mmm-spike-order-ensemble-v1-summary.json`, SHA256
`60cf1efe7e7ca7748f621e7af621e410962a75ce3e28032305ff0ecfab895542`.
The source/moment auditor recognizes an explicit descending order; it must
not reconstruct these factors with ascending masks.

Next isolate publication staleness: single original frozen model, refit only
when newly admitted labels change the64-label window, compare against its
32-frame publication schedule on identical queries. Charge every refit and
prediction, and respect delayed arrivals and missing labels. This is an
ideal immediate-publication diagnostic, not a working asynchronous serving
claim. If it fails, stop blaming order or cadence and evaluate a genuinely
different structured/adaptive prior with a new predeclared contract.
All seven goals remain OPEN; production and whitepaper untouched.
