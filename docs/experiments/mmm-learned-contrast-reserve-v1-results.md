# Finite-horizon reserve for known-hypothesis acquisition

**Frozen joint screen: FAIL on design and untouched confirmation.** The
80/48 request reserve greatly improves a change at clock384 but harms
early, noisy, delayed, skewed and out-of-family cases. The candidate is not
promoted. This is an isolated Goal1/7 component experiment, not serving or
agent-task evidence.

## Evidence and controls

The [frozen protocol](mmm-learned-contrast-reserve-v1-protocol.md) used seven
cases, 16 independently fitted baselines x 16 streams per case/split, and
six arms: original random/uncertainty/learned versus the same policies with
the reserve. All arms made exactly128 requests; reserve arms made exactly80
before clock384 and48 after. Forecasts preceded selected, nonmissing, due
labels. The [independent verifier](../../research/learned-contrast-reserve-v1-verify.mjs)
reconstructed the generator conditional probability and every forecast's
proper expected/realized Brier, request/delivery timing, recovery and frozen
decision rule for all 1,792 trajectories and 10,752 arm trajectories per
split. Source hashes match. A full confirmation recollection under `-race`
was byte-identical to the original archive; package tests and `go vet` pass.

Archive SHA-256: design
`4f6ab06caf0b0ac6a33795720c5eb52b5ac96738af3e20e559952c3038deaddd`;
confirmation
`0f5ee993a9fd058d85a06cec5632c9880e76d919f4d3d2833757140239724aed`.

## Confirmation result

Old and reserve columns use the **learned-disagreement** policy on the same
potential-outcome tape. Lower post expected Brier and recovery clocks are
better. All figures below are means over 256 trajectories per case.

| Case | Post expected Brier old -> reserve | Recovery clocks old -> reserve | Frozen result |
| --- | ---: | ---: | --- |
| Early bit2 | .1236 -> .1373 | 128.8 -> 144.8 | Protection fails |
| Late bit2 | .2798 -> **.2025** | 119.9 -> **100.0** | Late rescue passes |
| 20%-noise bit2 | .2538 -> .2621 | Not applicable | Protection fails on interval |
| 32-clock delayed bit2 | .2222 -> .2442 | 173.1 -> 189.6 | Protection fails |
| 10%-bit2 context | .1631 -> .1847 | 128.5 -> 144.8 | Protection fails |
| Stable 20%-noise | .1777 -> .1777 (full stream) | Not applicable | Non-harm passes |
| Majority out-of-family | .2932 -> .3046 | Not applicable | Non-harm fails |

The late case gets about38.3 **arrived** post-clock384 labels instead of25.4,
with exactly48 versus about32 post-clock384 requests. Its recovery misses
fall from .543 to .020; post expected-Brier gain is .07725 with 16-fit
cluster interval `[.06738,.08713]`. Design independently agrees: recovery
120.0->99.6 clocks, misses .535->.016, and post expected Brier .2798->.2020.
The late gate therefore passes on both splits.

The joint rule still fails decisively. Confirmation post expected-Brier harm
is .01372 early, .00826 noisy (fit-cluster harm upper .01431), .02196
delayed, .02158 skewed, and .01141 on majority OOD (upper .01660). Early,
delayed and skewed recovery also slow by about16-17 clocks. Design shows
the same pattern. Reserve learned still beats its *reserve* random and
uncertainty controls at the frozen magnitude in all five known cases on
design and four of five on confirmation. That relative advantage does not
repair its harm versus the old learned arm. Stable forecasts are effectively
unchanged.

## Interpretation and limits

The stage boundary at clock384 is a declared horizon policy, not a runtime
changepoint detector. It also coincides with the `late_bit2` change by
construction, based on the prior consumed transfer diagnostic. Independent
stream seeds do not remove that structural alignment; changes at other late
clocks or under unknown horizons remain untested. The result supports a
**local budget-allocation trade-off**, not generally adaptive observation.
More late labels cannot fix the out-of-family majority model, and spending
them late removes labels needed for earlier shifts.

The learner state remains232 bytes. A fresh four-run 20,000-sample component
cost check of the unchanged forecast/update plus the same nomination
primitive measured update p99 1.58-5.88us and forecast-plus-select p99
84-167ns. It excludes fitting, retrieval, storage, queues and request
latency. The reserve adds a constant-time stage-budget calculation, not a
new fitted state; loaded cost is unmeasured.

The next credible path is evidence-responsive allocation with independently
audited change signals and a false-trigger budget, tested over *off-boundary*
change times. Simply adjusting the 80/48 ratio on these consumed cohorts
would overfit. No production or whitepaper change was made. All seven whole
research goals remain open.
