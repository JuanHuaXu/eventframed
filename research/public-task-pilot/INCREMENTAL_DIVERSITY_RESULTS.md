# CPU diagnosis and exact incremental diversity candidate

## Evidence before optimization

The CPU diagnostic completed256 actual service calls,128 per packing arm, with
matching journals and200 nominated records throughout. All recorded source and
profile hashes were verified. Profiles are retained beside
task-lexical-profile-results.json. The protocol precedes dispatch.

Experimental CPU samples total2.66s: diversify accounts for2.13s cumulative
(80.08%), with tokenJaccardSets1.85s (69.55%). Control samples total2.50s:
diversify1.46s (58.40%), tokenJaccardSets1.27s (50.80%). These percentages overlap
and must not be added. They describe sampled CPU, not fractions of wall latency.

```sh
go tool pprof -top -cum research/public-task-pilot/task-lexical-profile-results.json.true.cpu.pprof
go tool pprof -top -cum research/public-task-pilot/task-lexical-profile-results.json.false.cpu.pprof
```

## Exact candidate, not weaker selection

Original diversity repeatedly computes each survivor's maximum similarity to
every prior selection. The new research-only helper retains that maximum and
updates it against the newly selected item. Inductively the penalty at every
selection round is identical. Score calculation, priority-dependent penalties,
certified-bucket exemptions, strict greater-than tie breaking, input order and
the appended unselected tail remain unchanged. No candidate or comparison
needed by the maximum is discarded. Production Select is unchanged.

For n candidates, k selected items and token-comparison cost t, repeated
comparisons cost O(n*k^2*t); the incremental version costs O(n*k*t), plus the
existing candidate-copy/removal costs. It adds O(n) penalty storage. With k
treated as a fixed cap both are linear in n; the useful improvement is removing
the repeated work across k, not claiming sublinear corpus retrieval.

The initial2000-case race-enabled exact-output test passed across varying sizes,
limits, ties, overlapping/empty5w1h fields, priorities and AP keys. A separate
200-candidate20-seed test covers the benchmark size and duplicate IDs. No raw
full-text replacement is used: tokenization remains through Event.FrameText.

## Isolated benchmark

Apple M4, Go darwin/arm64, CPU4, three200ms runs, pack limit20. Nanoseconds/op:

| n | Original runs | Incremental runs | Median ratio |
| --- | --- | --- | --- |
| 50 | 1439423,1313555,1318256 | 194110,195060,194947 | 6.76x |
| 200 | 6879140,6901979,6889703 | 974708,976418,973607 | 7.07x |

At200, cumulative allocation per operation rises from approximately703008bytes
and1974 allocations to704807bytes and1975 allocations. This is the bounded
penalty slice cost, not a memory-saving claim. Fixture content is generated
5w1h fields, not the service's full public corpus workload.

```sh
go test -race ./internal/packing -count=1 -timeout=120s
go test ./internal/packing -run '^$' -bench '^BenchmarkResearchIncrementalDiversity$' -benchmem -benchtime=200ms -count=3 -cpu=4
```

This candidate is not wired into Select or the service overlay. The speedup is
for diversity only; it does not establish an end-to-end throughput rescue.
Next integrate through a fresh overlay, verify all saved public packing results,
then repeat the unprofiled fixed-arrival experiment including failed conditions.
All seven full goals remain open. No production or whitepaper changes.
