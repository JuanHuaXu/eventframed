# Context-tree averaging v70 results

Decision: **FAIL the frozen learner screen; do not integrate or replace MMM.**
The exact tree mixture passes mathematical tests but is not a general rescue.
All seven research directions remain open.

## Evidence

512 paired fits, 64 identical labels per arm, eight seeds per cell, two phases,
four families, two noise levels and four old/new-label compositions. Each fitted
model is scored over the exact 512-input uniform distribution, integrating the
declared Bernoulli outcome noise. These are synthetic expected Brier risks, not
observed agent accuracy, partial-view MMM performance, or a continuous-learning
trial. Only fully labeled historical training inputs enter either learner.

All eight primary improvement checks failed. Twenty-six of 64 protection cells
exceeded the frozen +0.005 Brier harm tolerance. Selected confirmation means
(lower Brier is better; all new-label counts are within a fixed 64-label window):

| Family | Noise | New labels | Subset control | Tree | Tree minus control |
| --- | ---: | ---: | ---: | ---: | ---: |
| Multiplexer3 | 0 | 64 | 0.004123 | 0.001480 | -0.002642 |
| Multiplexer3 | 0.1 | 64 | 0.107805 | 0.119047 | +0.011242 |
| Multiplexer3 | 0 | 32 | 0.206115 | 0.210325 | +0.004211 |
| Majority3 | 0.1 | 64 | 0.121099 | 0.141490 | +0.020391 |
| Parity3 | 0.1 | 64 | 0.101646 | 0.141872 | +0.040226 |
| Parity4 negative control | 0 | 64 | 0.021616 | 0.255255 | +0.233639 |

The clean multiplexer benefit has descriptive paired-seed 95% t interval
[-0.003010,-0.002274]. The noisy multiplexer regression interval
[-0.008975,+0.031458] includes zero; the noisy majority and parity3 intervals
[+0.012344,+0.028438] and [+0.021910,+0.058542] do not. Eight-seed t intervals
are descriptive approximations, not simultaneous confidence sequences or
distribution-free certification. The frozen screen uses means, not these intervals.

## Interpretation and limitations

Branch-specific pooling helps the noiseless multiplexer but does not resolve
noisy interactions or mixed-regime labels. The parity4 control exposes the known
depth-three capacity limit. Failures remain after excluding that control, so the
decision is not solely a deliberately out-of-family challenge.

Protocol limitation: requiring an absolute 0.005 Brier improvement is unattainable
when the control risk is already below 0.005. The clean 64-new-label multiplexer
is such a cell, so its primary failure alone is not evidence of a poor learner.
The substantial independent protection failures still reject general adoption.
Preserve this protocol/result unchanged; a future protocol should predeclare
excess-risk-relative or floor-aware improvement thresholds before new data.

This implements our bounded tree-prior averaging adaptation of
[Chipman, George & McCulloch (1998)](https://www.rob-mcculloch.org/some_papers_and_talks/papers/published/cartfinal.pdf),
not their stochastic CART search. Exact inference within a declared hypothesis
family does not establish that the family matches reality. Label permutation
tests are predictive symmetries, not SCM-backed causal identification.

Recommendation 1 advances through a fresh generator-level paired screen and a
documented protocol limitation. Recommendation 4 now has another explicitly
rejected challenger. Recommendation 2 remains open: fitting a stationary tree to
a mixed-regime window does not model the change itself. Do not tune tree depth or
mixing weights against these confirmation cells and rename it validation.

## Reproduction and integrity

- Protocol: `mmm-context-tree-v70-protocol.md`, frozen before the run.
- Raw artifact: `mmm-context-tree-v70.jsonl`, 512 rows plus metadata.
- SHA256: `9e156ef09bbfef04be922e092d03fd1b01c25e08c0f6c63ce3e6496c1a7d6e86`.
- All 48 recorded source/protocol/dependency hashes verified by
  `node research/context-tree-v70-summary.mjs` from the runtime repo root.
- Experiment command: `EVENTFRAME_CONTEXT_TREE_ARTIFACT=<fresh-absolute-path> go test ./internal/observationlearners -run '^TestContextTreeV70Experiment$' -count=1 -v`.
- 512 paired fits completed in 6.70 seconds. Per-fit timings are diagnostic only;
  a separately launched race suite may overlap, so do not use them as a latency gate.
- Independent ordinary-probability evidence enumeration agrees with log-space
  context inference to 1e-12. Label complement, coordinate rotation, replay,
  input bounds, probability bounds and training-slice immutability checks passed.
- `go vet ./internal/observationlearners` passed.
- Focused race checks passed in 1.419 seconds: new tree evidence/symmetry,
  subset evidence/conditional inference, observer parity and stopping-gate tests.
  The package-wide race command was deliberately stopped after about five minutes
  because it also replays archived experiments. It did not complete and is not
  reported as a pass. No unrelated archived replay was changed or deleted.

## Isolated cost measurements

After stopping the broad replay and finishing focused race checks, on Apple M4,
Go1.27.1 darwin/arm64, GOMAXPROCS10:

```text
go test ./internal/observationlearners -run '^$' -bench '^BenchmarkContextTree' -benchmem -benchtime=500ms -count=3
BenchmarkContextTreeFit64-10   93  6096930 ns/op  116088 B/op  839 allocs/op
BenchmarkContextTreeFit64-10   98  6073966 ns/op  116088 B/op  839 allocs/op
BenchmarkContextTreeFit64-10   99  6080207 ns/op  116088 B/op  839 allocs/op
BenchmarkContextTreeLookup-10 1000000000 0.2551 ns/op 0 B/op 0 allocs/op
BenchmarkContextTreeLookup-10 1000000000 0.2545 ns/op 0 B/op 0 allocs/op
BenchmarkContextTreeLookup-10 1000000000 0.2544 ns/op 0 B/op 0 allocs/op
```

A 64-label fit takes about6.1ms and allocates about113KiB in this isolated
microbenchmark. The retained 512-entry float64 table is4KiB. Lookup numbers are
tight-loop table-load throughput only, not per-request prediction latency, partial
observation marginalization, queue cost, persistence cost or service tail latency.
These costs are conditional on nine Boolean inputs and depth3, not arbitrary
EventFrame dimensions. A fast compiled table does not rescue the learning failure.

No served forecast, production configuration, shared database, whitepaper,
commit or remote was changed. New source and experiment are research-only.
