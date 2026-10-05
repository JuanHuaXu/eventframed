# Finite beta-log table: exact component reduction

The cap64 model uses only predictive terms
log((y+.5)/(n+1)) and log(1-(y+.5)/(n+1)),0<=y<=n<=63.
A per-fit64x64x2 table evaluates those terms once rather than repeatedly for
each mask and interval. Only initialized triangular entries are used. Rejected
inputs cannot exceed the cap. No global persistent cache or empirical shortcut
is introduced, and both complement and positive expressions retain the original
floating-point evaluation order.

This research-only candidate layers on batch normalization and tail transforms.
Its fit-local table is64KiB; stack space is not included in heap B/op metrics.
The benchmark includes table initialization, unlike a prewarmed shared-table
measurement. Asymptotic interval complexity remains quadratic in label count.

## Verification

Independent partition enumeration, as-of/bounds, empty/maximum history and
concurrent fitting checks passed with race instrumentation.25 paired full fits
(history0/1/16/64/272, mass0/1e-12/.5/.95/1) match the batch candidate bit-for-bit,
including forecast vectors, log evidence and boundary weights. These fixtures
are evidence of implementation equivalence, not a universal floating-point proof.
Prior frozen-fit tolerance and quality failures remain unchanged.

## Component timings

Apple M4, three sequential300ms benchmark targets.272-frame delayed/missing
fixture,cap64,context checks retained; includes per-fit table construction.

| Variant | ms/op, three runs | Heap B/op | Allocations |
|---|---|---:|---:|
|Batch normalization|42.402,42.198,42.183|477184|7|
|Batch plus table|29.125,29.915,29.068|477184|7|

Median reduction about31%. No separate reduction in heap allocation is claimed.
This does not establish loaded deadline completion, a tail bound or a rescue of
the learner's accuracy. The prior service experiments show why component gains
must not be promoted directly to operational claims.

```sh
go test ./internal/observationlearners -run '^TestTable' -race -count=1 -v
go test ./internal/observationlearners -run '^$' -bench '^BenchmarkTablePair$' -benchtime=300ms -count=3
```

Next: loaded changed-work validation with explicit offered-job accounting.
Keep any deadline allowance comparison separate: changing it simultaneously
would prevent attributing the result to this arithmetic optimization. Production
code, prior frozen artifacts and the seven direction statuses are unchanged.
