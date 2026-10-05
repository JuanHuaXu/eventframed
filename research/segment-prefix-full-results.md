# Prepared prefix through full posterior fitting

Research-only full fitter reuses the checked ordered-label prefix while
recomputing current eligible origins, elapsed-frame gaps, boundary posterior and
hazard mixture. Cached state does not replace those chronology calculations.

36 complete posterior pairs pass bit-for-bit: history32/64/272,cap16/64,
mass0/.95/1,hazard.01/.5. The fixtures contain delayed/missing outcomes.
Four concurrent fits reading the same prepared state match the reference.
Mid-fit cancellation returns no model; saved prepared state remains unchanged
and refitting recovers exactly. Changing an earlier eligible label or removing
one through delayed availability rejects prefix reuse. Both tests pass under
race instrumentation. These are scoped tests, not a proof for arbitrary
future mutations, eviction algorithms or cross-tenant persistence.

## Full-fit timings

Apple M4, sequential300ms benchmark targets, three repetitions.272-frame
fixture,cap64. Preparation happens outside these timed fits and was separately
measured in `segment-prefix-results.md` at about28ms with>5.5MB retained state.

| Operation | ms/op, three runs | Heap B/op | Allocations |
|---|---|---:|---:|
|Full table-based rebuild|28.897,29.173,29.069|477184-477185|7|
|Prepared full fit|14.817,15.072,15.091|477184|7|

Median reduction about48%. This is not cold-start performance: preparation
plus one prepared fit remains slower than rebuilding once. A real service
experiment must include cold preparation and distinguish compatible from
invalidated prefixes. It must not populate the candidate from verification
oracles or omit state construction from the deadline budget.

```sh
go test ./internal/observationlearners -run '^TestPrefix(FullChronology|OwnershipAndCancellation)$' -race -count=1 -v
go test ./internal/observationlearners -run '^$' -bench '^BenchmarkPrefixFull$' -benchtime=300ms -count=3
```

No change to production, prior archived fitters, or claims of learning quality.
Next: cold-start service workload with explicit prefix invalidation and memory
accounting. All seven research directions remain open.
