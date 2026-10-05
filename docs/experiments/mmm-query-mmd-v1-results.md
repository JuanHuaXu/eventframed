# MMD representativeness: no acquisition rescue

Follow the [frozen protocol](mmm-query-mmd-protocol.md). This is an R-I-inspired
Bernoulli adaptation of the representativeness factor in Tang/Sloman/Kaski2026,
not R-IDeA and not a theorem-backed de-amplification certificate. A fixed
bandwidth1 RBF kernel compares the selector's63 observed input multiset with
all161 visible input samples. Append each candidate's input and compute the
signed factor1-.5*(post-MMD/current-MMD). Publication remains unchanged.

## Outcome

Phase1 delayed,672 histories; every active selector buys672 queries:

| Selector | Actual-answer sampled Brier | Teacher-weighted population Brier |
| --- | ---: | ---: |
| Random | 0.167334 | 0.167203 |
| Entropy | 0.166630 | 0.166794 |
| Joint8 | 0.167366 | 0.166848 |
| MMD-weighted joint8 | 0.167362 | 0.166820 |
| Representativeness alone | 0.166908 | 0.167000 |

MMD-weighted joint changes34/672 choices, representativeness alone544/672.
Neither beats entropy on these aggregate metrics. Both methods fail both
frozen primary/supplementary all-cell nonharm/transition screens. The small
combined-method difference is not a validated gain. Representative-only's mean
improvement over random does not pass the full rule or establish generalization.

All2688 records and84cells are retained. The actual dataset triggers neither
negative factors nor zero-denominator fallback; both are covered by unit tests.
These are consumed-data comparisons, not untouched or sequential confirmation.

## Verification

-200 independent append-MMD unit comparisons, baseline direct checks, duplicate
 weighting, permutation/bit symmetry, ownership, zero denominator, negative
 factor and invalid-input checks pass. Input-view tests trap hidden outcomes,
 teacher probabilities and future161+ inputs.
-The independent audit rebuilds one signed empirical measure, rather than the
 implementation's three kernel sums. It checks10621 baseline/append MMDs,
 9277 information scores,32256 selected branch losses,1008 means and672 paired
 bounds. All pass, including unchanged original-policy controls.
-Full replay is byte-identical; source and implementation hashes are recorded.

Benchmark includes the as-of input view, full MMD setup and both selectors with
precomputed forecast arrays:2016 calls, median0.116042ms,p99 0.195542ms,
maximum0.231541ms, Nodev26.8.1. It excludes posterior fits, database retrieval,
persistence and serving. This low local cost does not establish total acquisition
cost parity or compensate for failed quality criteria.

Artifacts: `mmm-query-mmd-v1.json`, `-replay.json`, `-audit.json`,
`-benchmark.json`, `-replay-benchmark.json`.

## Decision

Do not tune bandwidth, target horizon or factor weight on the observed failures.
The frozen representativeness-only component is not a rescue. A genuinely
different hypothesis/observation design or a separately trained signed-usefulness
model remains a research lead, with all as-of and external-evidence constraints
intact. A disagreement proxy may be tested as a heuristic, but cannot inherit
the invalid subset implication identified in the source review. None of this
completes the other six goals, prospective agent validation or loaded serving.
All seven goals remain open; no production or whitepaper promotion.
