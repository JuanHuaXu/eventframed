# Evidence-responsive observation budget, v1

**Frozen joint screen: FAIL on both fresh splits.** A fixed-mixture
likelihood-ratio monitor reallocates labels after arriving evidence and
improves several scores, but detects off-boundary late shifts too late to
meet the predeclared recovery and miss gates. No candidate is promoted.
This is a bounded Goal1/7 component, not an Anti-Pigeon certificate.

## Design and audit

The [frozen protocol](mmm-observation-responsive-v1-protocol.md) compared
old and responsive random/uncertainty/learned policies on eight cases,
16 independently fitted baselines x 16 streams per case/split, 512 clocks
and exactly128 requests per arm. The monitor mixed eight fixed restart
origins and ten declared alternatives against the frozen fitted baseline.
It saw only selected, nonmissing, **arrived** labels; its alert changed only
later acquisition rates, never the already-scored forecast. The
[independent verifier](../../research/observation-responsive-v1-verify.mjs)
reconstructed all 2,048 trajectories and 12,288 arm trajectories per split:
generator conditional probabilities, per-tick scores, label chronology,
budgets, and every monitor crossing from recorded fitted baseline
probabilities. Repeated context keys had identical baseline probabilities.
Source hashes match; a full confirmation replay under `-race` was
byte-identical. Package tests and `go vet` pass.

Archive SHA-256: design
`43e8f8ab8c1c9482186c428234e32b656b5cb3eae0a87f57b3ae4ead62b218a0`;
confirmation
`27082bf31450a35b869f4f297e6261b339cbf1dbd2ffb7ee189b30b945e19712`.

## Confirmation result

The table compares old and responsive **learned-disagreement** arms on the
same potential-outcome tape. Lower Brier and recovery clocks are better.
Fit-cluster intervals use `mean +/- 3.5 SE` over16 fits.

| Case | Post expected Brier old -> responsive | Recovery clocks old -> responsive | Frozen result |
| --- | ---: | ---: | --- |
| Early160 bit2 | .1306 -> **.1248** | 128.5 -> 121.4 | Protection passes |
| Late360 bit2 | .2392 -> **.2261** | 126.5 -> 121.1 | **Late recovery fails** |
| Late400 bit2 | .3072 -> **.2822** | 109.8 -> 107.5 | **Late recovery fails** |
| 20%-noise bit2 | .2556 -> .2604 | Not applicable | Protection interval fails |
| 32-clock delayed bit2 | .2223 -> .2284 | 174.0 -> 178.2 | Protection and arrival parity fail |
| 10%-bit2 context | .1628 -> **.1452** | 129.4 -> 115.8 | Protection passes |
| Stable 20%-noise | Essentially unchanged | Not applicable | False-alert and harm gates pass |
| Majority out-of-family | .2923 -> **.2854** | Not applicable | Relative non-harm passes |

On confirmation, `late360` post expected-Brier gain is .01311, but its
fit-cluster interval `[-.00036,.02657]` crosses zero. Recovery improves only
5.4 clocks, and miss fraction falls .176->.137, below the required .10
absolute reduction. `late400` gains .02491 with positive lower endpoint
.00734, but recovery improves only 2.3 clocks and misses fall .781->.715.
Design reproduces the same failure: late360 recovery lead 6.1 clocks,
late400 2.9. The responsive learned arm receives about45.0 instead of30.4
post-change labels on late360 and38.3 instead of22.5 on late400, but the
median monitor alert is 61-66 clocks **after** those changes. Once the alert
arrives, too much of the evaluation horizon is gone. Neither late change
aligns to a monitor start or fixed budget boundary.

The monitor alerts on 2/256 stable design trajectories and1/256 stable
confirmation trajectories, under the frozen threshold100. That is an
**empirical** false-alert observation, not a 1% error guarantee: the fitted
baseline is approximate, so the exact-conditional-null martingale premise
does not hold automatically. Early160 and skew256 post scores and recovery
improve. The majority OOD score improves relatively but remains about .285;
no valid unknown-rule recovery is established. Responsive learned beats
responsive random and uncertainty at the frozen score magnitude in5/6
known cases on design and6/6 on confirmation. This does not cure the
late-timing, noisy or delayed protection failures. On delayed256 the
arrived-label mean gap also exceeds the frozen two-label ceiling.

## Cost and next inference

Working learner plus monitor state is888 bytes. Four isolated 20,000-sample
cost runs measured monitor-plus-update p99 2.50-8.17us and
forecast-plus-select p99 84-167ns, within their component gates. These
numbers exclude fitting, retrieval, storage, queues and loaded request
latency. The verifier reconstructs monitor arithmetic from recorded
baseline probabilities; it does not independently refit the baseline.

The result rules out this threshold100, eight-start, 64-clock burst as a
joint rescue on the declared family. A next candidate needs earlier
evidence or a better horizon-aware decision objective, plus explicit
control for fitted-baseline uncertainty and delayed-label availability.
Tuning this monitor on the consumed archives would overfit. Production and
the whitepaper are unchanged; all seven whole research goals remain open.
