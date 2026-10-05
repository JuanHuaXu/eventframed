# Full private graph-deletion component cost

`BenchmarkPrivateDelete` runs actual incoming discovery, cosine calculations,
lazy reconnection, intermediate immutable edits, target retirement and entry
summary updates. Source parsing/construction and capture verification are outside
the timer. The benchmark retains the latest result to prevent elimination;
it does not simulate a population of stalled readers or a persistent writer.

Command: `RESEARCH_DELETE_CAPTURE=.../serial-work-control.json go test
./internal/researchindex -run '^$' -bench '^BenchmarkPrivateDelete$'
-benchtime=1s -count=3 -cpu=4`. Apple M4, darwin/arm64, no race instrumentation.
The command passed, with package elapsed21.647s.

| Corpus / target | ns/op, three repeats | B/op | allocs/op | Pair calls | Changed records |
| --- | --- | ---: | ---: | ---: | ---: |
| 800 / seed-0 | 67141,67556,68369 | 180816 | 1732 | 0 | 36 |
| 800 / seed-24 | 194212,195652,195383 | 462584 | 4430 | 29 | 62 |
| 6400 / seed-0 | 73781,73845,73986 | 193080 | 1853 | 0 | 39 |
| 6400 / seed-2709 | 290921,285378,287532 | 721096 | 6850 | 15 | 87 |

The selected cases are the ordinary seed-0 deletion and the forced entry-point
deletion at each corpus size. They were selected in code before running this
benchmark, but are not an independent confirmation distribution or worst-case
coverage. No p99 inference is possible from these per-benchmark means.

## Interpretation

The combined component is67-291us in these fixtures, no longer just a supplied-edit
copying measurement. This supports pursuing integration rather than another
micro-optimization screen. However,181-721KB allocation per operation is material:
full record ownership copies and multiple per-level immutable roots add GC work.
Allocation rate and live memory must be measured under sustained offered load.

No ID-map/revision updates, persistent transaction, synchronization with readers,
recovery, external vector ownership or insertion work is included. These figures
do not establish sub-100ms durable serving, and are not a paired speedup comparison
with earlier differently scoped benchmarks. Next integrate coherent metadata and
publication, then test long load; retain allocation reduction as a measured lead.
All seven whole research goals remain open; production code is unchanged.
