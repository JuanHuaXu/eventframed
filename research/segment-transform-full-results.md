# Full-posterior transform validation

Follow-up to `segment-transform-go-results.md`. Research/test-only copies leave
the frozen fitter, production serving and all previous experiment artifacts
unchanged. No learning-quality claim is revised.

## Tests

The candidate uses the original full posterior recurrence with only the
likelihood-builder call replaced. Four existing independent/lifecycle tests
were replicated against it: partition enumeration using factorial beta
integrals, as-of and invalid-input checks, empty/maximum history, and concurrent
fits. All passed with the Go race detector. In particular, excluded outcomes
cannot change the model, repeated fitting is idempotent, and a subsequent fit
does not mutate an earlier snapshot.

A separate paired test covered 54 combinations: history lengths 16/64/272,
label caps 1/16/64, generic masses 0/.95/1, and hazards .01/.5. Histories have
missing labels and variable delays. All log evidence, last-segment weights and
eligible-origin lists were exactly equal. Across 27,648 forecast comparisons,
maximum absolute error was 1.5543122344752192e-15 (tolerance 1e-12).

## Full-fit component timings

Apple M4, darwin/arm64, sequential 300ms benchmark targets. History size is NOT
label count: delayed/missing outcomes are excluded and eligible labels capped
at 64. Existing prior/index caches are warm, not cold-start measurements.

| History frames | Original ms/op | Transform ms/op | Original B/op | Transform B/op |
|---|---:|---:|---:|---:|
| 32 | 11.122177 | 9.008364 | 308752 | 472592 |
| 80 | 60.232917 | 53.525438 | 310016 | 473856 |
| 272 | 65.835608 | 57.852028 | 313344 | 477184 |

Six allocations per original fit versus seven per transformed fit. The extra
163,840 bytes are fit-local scratch. These are short exploratory measurements,
not confidence intervals, tail latency, queue deadlines, or loaded service
benchmarks. The 272-frame fixture saves about 12% time; it does not turn a
roughly 58ms fit into a cheap hot-path operation.

```sh
go test ./internal/observationlearners -run '^TestTransformSegment' -race -count=1 -v
go test ./internal/observationlearners -run '^TestTransformFullPair$' -v -bench '^BenchmarkTransformFullPair$' -benchtime=300ms -count=1
```

Next: evaluate deadline completion and foreground interference in an isolated
bounded background workload before considering integration. Production shadow
direction 6 remains open, as do the learner's separate quality shortcomings.
