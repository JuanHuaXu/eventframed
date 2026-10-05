# v96: interval-local weighting over learned experts

## Verdict: FAIL, 98/106 gates

All48 whole-stream and48 late-half non-harm gates pass. Both majority-to-parity
recovery gates pass, but both parity-to-majority recovery gates and all6
stationary interaction-gain gates fail. This is a protection/gain tradeoff,
not a completed rescue or a whole-direction success.

The [frozen protocol](mmm-online-interval-v96-protocol.md) retains every v95
control and adds the interval arm on768 new paired streams with new seeds,
two phases and12 cases. Training windows, fit times, input views and model
forecasts are identical across the online controls; only weighting changes.
Neither the simulator's change time nor a future outcome enters the selector.
Both views receive complete full-frame training evidence, not selective labels.

See [raw data](mmm-online-interval-v96.json),
[summary with all106 gates](mmm-online-interval-v96-summary.json), and
[independent evaluator](../../research/online-interval-v96-summary.mjs).
Paired32-trajectory intervals use the predeclared z=3.5 normal approximation;
they are not exact coverage guarantees or anytime confidence sequences.

## Same-stream confirmation comparison

| Case and segment | Generic Brier | Fixed share | Interval |
|---|---:|---:|---:|
| Parity4, all256 | 0.096451 | 0.073742 | 0.092368 |
| Parity4, late128 | 0.071568 | 0.051301 | 0.069121 |
| Majority3, all256 | 0.075463 | 0.075653 | 0.076902 |
| Majority to parity, late128 | 0.205803 | 0.202548 | 0.198301 |
| Parity to majority, late128 | 0.174070 | 0.183986 | 0.175470 |

These compare controls on the SAME v96 streams, not unmatched v95 means.
Interval majority-to-parity gain is0.007502 with interval[0.005082,0.009922],
passing recovery. Parity-to-majority gain is-0.001400 with
interval[-0.001748,-0.001052]: still harmful, although within the1% non-harm
margin and much less harmful than fixed sharing. We do not replace the failed
positive recovery requirement with the weaker non-harm result.

Stationary parity4 all-stream gain shrinks from fixed-share0.022709 to
interval0.004082, below the required0.005. All6 interaction gains fail in this
way despite positive lower bounds. The interval arm's parity4 late expected
accuracy is94.34%, versus94.96% for the fixed-share control and93.20% for
generic. These are simulator expected rates with5% noise, not real-agent or
original MMM94.7% replications.

## What was verified

The finite-horizon implementation uses geometric interval learners and a
surrogate weighted-loss meta-update. A separate literal interval enumeration,
with direct probability base updates instead of log weights, matches its
predictions. Convexity, positive weights, invalid input non-mutation, ordered
feedback, duplicate rejection and horizon exhaustion are tested.

- Reference/lifecycle race tests PASS:1.254s package time.
- Learned-stream race smoke PASS:2.817s, including exact paired replays.
- Generation PASS:93.786s; exact768-record replay PASS:88.599s.
- Vet, independent source/summary reproduction and whitespace checks PASS.
- Artifact SHA256: a6ac4766c413a26c87d81129dd23e44d3e2b3c29d420321ee0c249931f7d7cfd.

Recorded zero prefix-bound defects apply ONLY to the retained v95 online
controls. The interval arm does not inherit their constant lifetime bound.
The SAOL-inspired interval comparison and convex output do not establish a
finite1% guarantee; the declared empirical gates determine this experiment's
verdict. There is no same-instance concurrent or missing-feedback guarantee.

## Performance

Apple M4,Go1.27.1 darwin/arm64,GOMAXPROCS10. The [full fixture](mmm-online-interval-v96-benchmarks.txt)
takes114.06-134.31ms per256-step stream across three single-run repetitions.
This includes eight refit rounds, three constructed models, all six arms,
two views, scoring and hashing. It is not per-query latency or a measured
candidate-only regression against v95's different benchmark sample.

The [interval kernel alone](mmm-online-interval-v96-kernel-benchmarks.txt),
three500ms repetitions, takes92.22-92.77us per256-update lifetime, about360-362ns
per update on average. It allocates49,056 bytes and511 objects per lifetime.
Single cold repetitions in the full timing file take282-321us and are retained.
Only ten slots are retained at the maximum supported horizon512, but interval
creation incurs allocation churn. Bounded live state is not zero allocation.
No serving path or production configuration was changed.

## Next action

The selector-only comparison exposes a real tradeoff; it does not prove stale
model windows are the only cause. Test [shorter model-window experts](../../research/short-window-experts-proposal.md)
alongside retained64-frame experts, using fresh streams and charging the extra
fits. Preserve all106 gates, the failed v95/v96 results and the distinction
between useful comparative evidence and a validated upgrade.
