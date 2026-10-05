# Conditional integration v10: diagnostic screen FAILED

Replayed192 consumed v9 streams and scored73,736 frames after supported tree
fits. This is a fixed-mask tree diagnostic, not a rerun of the online mixture or
an untouched confirmation. All uniform forecasts match v9's journal exactly.

Clustered-input post-change tree Brier, using the same acquired fields:

| Scenario/split | Uniform | Past-audit joint | Oracle joint |
| --- | --- | --- | --- |
| shift128 design | 0.186430 | 0.179845 | 0.179654 |
| shift128 confirmation | 0.200313 | 0.196286 | 0.196041 |
| recurring design | 0.233752 | 0.224881 | 0.224509 |
| recurring confirmation | 0.236978 | 0.230780 | 0.230275 |

Empirical integration misses the >=.005 shift gain in confirmation (.004027).
No group/window exceeds the .01 mean-harm allowance. Preserve the failure: this
does not qualify for advancing under the predeclared screen. Oracle integration
also gains less than .005 in that cell (.004272). Thus the tested estimator is
already close to this oracle reference on these masks; estimating the same input
law more accurately is not a supported standalone rescue. The oracle is not an
upper bound for every predictor or every realized sample.

The isolated integration change does improve several correlated cases, so the
assumption has measurable consequences. It does not explain all v9 failures.
Unchanged observation choices can still hide useful coordinates; weak fitted
trees and mixture weighting remain alternative limitations. A separate experiment
would need to test distribution-aware acquisition itself with a frozen protocol,
not relabel these fixed-mask forecasts as an online policy improvement.

## Audit and Scope

The joint estimate uses at most256 already-admitted audit inputs, with one total
uniform pseudo-observation. Trees still use64 audit labels. All fits occur after
forecasting and only at the original cadence. Labels are not used to weight the
input distribution. The true distribution is confined to the named oracle arm.

Exact weighted enumeration tests cover all3^9 partial assignments; tests also
cover uniform parity, full-observation invariance, input validation, immutable
snapshots, oracle normalization and final-outcome non-leakage. Targeted race tests
and command vet pass. Summary recomputes per-stream scores and applies all frozen
criteria. Full sources of the new calculation and the v9 input hash are retained
in the diagnostic artifact; original source snapshots remain in v9.

The frozen lookup has no dependency on the future corpus. Building its table is
exponential in the fixed nine-bit domain (3^9 states), so this is not evidence of
scalability to unrestricted real-world variables. No production code path uses it.

[Microbenchmarks](mmm-conditional-v10-benchmark.txt) measured7.25-7.36ns per
cached lookup with the returned value retained, and about56us/~312KiB per table
build. These are bounded numerical operations, not end-to-end latency results.

Artifacts: [protocol](mmm-conditional-v10-protocol.md),
[raw diagnostics](mmm-conditional-v10.json.gz),
[summary](mmm-conditional-v10-summary.json).

```sh
go run ./cmd/eventframe-observation-conditional docs/experiments/mmm-dependent-v9.json.gz NEW.json.gz
python3 research/conditional_summary.py NEW.json.gz NEW-summary.json
go test -race ./internal/observationlearners -run '^TestConditional' -count=1
```

The broad research objective remains open. This closes one mechanism diagnostic,
not the real-task, adaptive-window, Anti-Pigeon or shadow-learner requirements.
