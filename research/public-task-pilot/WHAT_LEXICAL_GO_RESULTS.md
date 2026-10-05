# Frozen lexical feature: Go port

The Go comparator matches all108 reference orderings (54 saved queries, lexical
alone and task-class-combined). Maximum absolute score discrepancy from the
JavaScript comparator is2.7755575615628914e-17; no order changes. Golden tests
verify reference source/input hashes and input immutability before comparing.
This validates a port, not a new outcome experiment or live packing integration.

Race-enabled tests also pass for empty/over-cap frontiers, duplicate identities,
NaN scores, malformed/duplicate what fields, oversize queries, plan-identity
mismatch, pre-cancelled context and zero-token stable ties. Cancellation checks
occur before work, while reading candidates and after sorting; this is not
preemption of each regular-expression call. Byte limits in Go are stricter than
JavaScript UTF16 length for some non-ASCII inputs; the public traces are ASCII.

The port preserves ordered term reductions rather than iterating over Go maps,
so randomized map iteration cannot alter the scoring sum order. It uses no
model, network, outcome labels, relation dictionary or corpus-wide scan.

Standalone benchmark, Apple M4 darwin/arm64, cpu4, three200ms samples:

| Frontier | Median average time | Approximate allocations |
| ---: | ---: | ---: |
| 50 | 76.125 microseconds | 25KB,549 allocations |
| 200 | 293.073 microseconds | 97.5KB,2128 allocations |

Repeated public Curiosity facts are cardinality fixtures, not independent data.
The benchmark includes what-field parsing, normalization, term scoring and
sorting, but excludes task planning, diversity/token packing, journal writes,
embeddings and serving load. These are averages, not p95/p99. Total bounded work
scales with input text/token volume plus candidate sorting; feature vocabulary
is frontier-local, not a corpus-size guarantee for upstream retrieval.

```sh
go test -race ./internal/researchcalendar -run TestLexical -v
go test ./internal/researchcalendar -run '^$' -bench BenchmarkWhatLexical -benchmem -benchtime=200ms -count=3 -cpu=4
```

Raw benchmark: what-lexical-go-benchmark.txt. Implementation/tests:
internal/researchcalendar/lexical.go, lexical_golden_test.go and
lexical_boundary_test.go. Original JavaScript and prior artifacts are unchanged.

Next integrate into selection without letting diversity resort away the lexical
order, while restoring original candidate scores/laws and journaling the new
method. Then run actual service comparisons and fresh outcome tests. All seven
whole directions remain open; production and whitepaper are unchanged.
