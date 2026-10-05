# Adaptive held-out forest: partial benefit, broad gain gate not met

320 fresh trajectories,10cases, two cohorts of16independent4096-label bases
per case. The frozen input forest replaces only uniform input weights inside
the retained subset learner. Count models, selector, Anti-Pigeon gate, audit
budget and64-label fit schedule remain unchanged. Fixed-observation and own-
observer candidates share training evidence with the uniform-input control.

All20 own-observer full/post non-harm screens pass. Only one of14changed-case
mean-gain screens passes (design dependent XOR2). The broader improvement goal
is NOT achieved. Preserve positive subthreshold benefits without promoting them
to complete validation.

| Case / cohort | Control post Brier | Forest post Brier | Gain [mean +/-3.5SE] |
|---|---:|---:|---:|
| Dependent bit/design |.195153|.193324|.001829 [.000786,.002872]|
| Dependent bit/second |.205645|.203591|.002054 [.000669,.003439]|
| Dependent XOR2/design |.206696|.200582|.006114 [.001931,.010297]|
| Dependent XOR2/second |.214862|.210461|.004401 [.001128,.007674]|
| Uniform parity4/design |.245900|.245899|.000001 [-.000022,.000023]|
| Uniform parity4/second |.241903|.241983|-.000080 [-.000408,.000247]|

Intervals are exploratory paired trajectory screens, not simultaneous or
anytime guarantees. Uniform changed-case effects are tiny/mixed. Stable and
null effects are tiny. The earlier histogram's large parity regression is not
observed here, but that is a cross-experiment comparison, not a matched
histogram ablation on this cohort. Fixed-arm dependent gains are very close
to coupled gains, consistent with most benefit occurring in prediction rather
than substantially better acquisition. This is not a formal mediation proof.

The higher-order XOR INPUT limitation from the fixed-mask screen remains;
dependent XOR2 OUTCOME here is a different case and does not resolve it.

## Cost

Fixed-arm acquisition costs equal control exactly. Second-cohort dependent-bit
own-observer cost falls4.72461->4.67908 coordinates/frame; dependent-XOR2
increases4.62329->4.62512. No blanket claim of cheaper acquisition.

Separate fit64 microbenchmark (Apple M4,3x300ms):

    uniform: 6.058672 / 6.185518 / 6.049626 ms; ~651264B,3allocs
    forest:  6.076692 / 6.066016 / 6.056705 ms; ~661448B,12allocs

Both are dominated by the same subset outcome fit. Differences lie within
the observed run variation; this is not proof of zero overhead or speedup.
Forest adds about10KiB and9allocations to this fitting operation. Foreground
lookups retain the existing compiled representation, but no loaded serving,
queue freshness, persistence or tail-latency benchmark is established here.

## Verification

- New isolated `internal/researchinput` adapter is bit-identical to the frozen
  test component on20random datasets; race1.431s.
- Original-control metric/tape parity on uniform and dependent cases passed
  under race3.528s; vet of all touched research packages passed.
- Collector completed320streams in56.41s;829pre-collection sources captured
  and verified. Shared fit/split counts, fixed costs, seed/cell/metric checks
  passed; summary replay byte-identical. No full collection replay claimed.
- Existing tracked diff remains fourfiles,+63/-1, untouched by this work.

Raw:`mmm-forest-integration-v1.jsonl`.
SHA256:`dc5f09d9938e2448b49f8849bc8690906df59c3abdeb45090aabbe53cd28d5b1`.
Frozen contract:`mmm-forest-integration-v1-contract.md`.
Summary:`mmm-forest-integration-v1-summary.json`.

## Next Boundary

Retain this as a partial research candidate, not a production default. No
split-ratio or threshold tuning. Useful next tests expand beyond perfect copied
inputs into imperfect/shifted dependence and delayed evidence, with the same
candidate frozen. Higher-order structure still needs a separate justified
technique. All seven original goals remain open; paper/remotes/production untouched.
