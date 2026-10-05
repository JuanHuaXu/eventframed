# Stable batch normalization: fresh-fit component improvement

Research-only replacement of pairwise log-add normalization over each512-member
family with max-shifted log-sum-exp. No model/prior/likelihood factors, label
selection, cap, horizon or hazard semantics change. Interval complexity stays
quadratic in label count; this reduces transcendental normalization calls, not
the asymptotic bound. Existing transform and cancellation checkpoints remain.
Frozen code and earlier source-hashed artifacts are unchanged.

For log weights v, evaluate max(v)+log(sum(exp(v-max(v)))). The finite family
weights here ensure at least one shifted term1. Generic mass0/1 is handled by
the existing family mixture, not by omitting or replacing evidence. Both family
normalizers remain finite independently of their mixture weights.

## Checks

Independent partition enumeration using factorial beta integrals, as-of/bounds,
empty/maximum histories and concurrent fits passed with race instrumentation.
54 full-fit pairs span history16/64/272,cap1/16/64 and generic masses
0,1e-12,.5,.95,1-1e-12,1. Maximum discrepancies versus frozen direct fitting:

- Forecast6.3282712403633923e-15 (limit1e-12).
- Log evidence1.4210854715202004e-14 (limit1e-10).
- Segment-boundary weight2.1094237467877974e-15 (limit1e-12).

Origins match exactly. This is numerical, not bitwise, equivalence; do not
assert exact output hashes for reordered floating-point arithmetic.

## Component timing

Apple M4, darwin/arm64. Three sequential300ms benchmark targets,272-frame
delayed/missing fixture,cap64. Warm caches; no concurrent task-started benchmark.

| Normalization | ms/op, three runs | Bytes/op | Allocations |
|---|---|---:|---:|
|Pairwise|58.346,58.674,58.474|477184-477186|7|
|Batch|42.733,42.378,42.756|477184-477186|7|

Median improvement approximately26.9%. Both arms use transformed tail evaluation
and context-aware fitting; this isolates normalization, unlike comparison to
the initial frozen implementation. This is not population performance evidence
or a loaded request benchmark. It does not establish that64 distinct jobs meet
100ms deadlines, nor rescue the learner's separate quality failures.

```sh
go test ./internal/observationlearners -run '^TestBatch' -race -count=1 -v
go test ./internal/observationlearners -run '^$' -bench '^BenchmarkBatchFullPair$' -benchtime=300ms -count=3
```

Next: test the revised cost on changed-work scheduling, retaining shadow-off,
offered-work accounting and freshness rules. Finite beta-predictive log terms
are another potential repeated calculation, but require a separate measured
comparison rather than silently stacking optimizations. All directions open.
