# Resolved admission isolated cost v56 results

**PASS the frozen isolated diagnostic screen.** At200 events, every cold/warm
trial improves new-admission mean by at least10%; exact-retry means improve in
every50/200 trial. This does not rescue v55's failed loaded latency screen or
establish writer non-harm. No runtime code or default changed in this experiment.

[Protocol](mmm-resolved-cost-v56-protocol.md),
[artifact](mmm-resolved-cost-v56.jsonl). SHA-256:
`bf67f7646c64a2036413c32f9010f0e3c579d0bf6630dd6d83a84477b9655f32`.

## Execution and integrity

```sh
go test -race ./internal/researchmemory -run '^TestResolvedAdmissionCostAccounting$' -count=3
go test -race ./internal/researchmemory -count=1
go vet ./internal/researchmemory
EVENTFRAME_RESOLVED_COST_ARTIFACT=<LOCAL_ROOT>/docs/experiments/mmm-resolved-cost-v56.jsonl go test ./internal/researchmemory -run '^TestResolvedAdmissionCostExperiment$' -count=1 -v
```

Targeted race checks PASS3.544s; full learner race suite PASS12.265s; vet clean.
Experiment Go PASS8.369s. Go1.27.1,darwin/arm64,10 CPUs,GOMAXPROCS10. Tests and
measurement did not overlap. The24 cells use fresh isolated SQLite files and
public deterministic fixtures. No production, private data, real-agent outcomes,
or service traffic. All captured source/hash pairs match measured local files;
the runtime files shared with v55 have unchanged hashes.

Each cell runs32 cycles. Both owners use the same combined verified cleanup.
Warm cells train on64 synthetic labels with one completed-publication barrier
per label before timing. Cold cells have no labels. There are no labels or fits
during measured operations. Within every comparison, the full training-original
and measured-original hashes match exactly. Every fresh original matches its
exact retry, expected ID/readiness and stored original read during cleanup.

Across24 cells:96000 measured originals,96000 fresh verified terminals, and
3000 additional terminal retries after reopening. Warm setup has768 labels in
total, separate from the measured lifecycle counts. Replay preserves label,
pending, queue and ID counts; first/last stored originals match. No mismatches
or unexpected errors. Raw sample lengths, positivity, counts, paired hashes and
the frozen screens were independently recomputed after execution.

## Timings

Means per batch, milliseconds. Each range spans the three trials; percentages
are paired within trial, not ratios formed from range endpoints.

| Events | State | Control / resolved fresh admission | Paired fresh reduction | Control / resolved exact retry | Paired retry reduction |
| --- | --- | --- | --- | --- | --- |
| 50 | Cold | 1.108-1.280 / 1.022-1.050 | 6.38-20.16% | 1.159-1.294 / 0.803-0.839 | 30.75-35.92% |
| 50 | Trained | 1.090-1.097 / 0.982-1.001 | 8.76-9.92% | 1.208-1.229 / 0.824-0.843 | 31.22-31.99% |
| 200 | Cold | 4.016-4.063 / 3.489-3.573 | 11.73-13.30% | 4.555-4.613 / 3.157-3.192 | 30.64-30.80% |
| 200 | Trained | 3.899-3.913 / 3.415-3.451 | 11.60-12.71% | 4.572-4.608 / 3.117-3.132 | 31.78-32.35% |

All six200-event fresh comparisons meet10%. All twelve retry comparisons meet
the no-more-than5%-regression condition. These are descriptive engineering
screens over fixed fixtures, not confidence intervals or population guarantees.

Combined verified cleanup remains roughly4.49-4.65ms for resolved200-event
batches. It is not changed by the admission optimization and still costs more
than fresh admission in this isolated fixture. Do not attribute every variation
in this unchanged phase to removal of admission reads.

After close, database sizes are exactly equal between arms:1966080 and2248704
bytes for50-event cold/trained cells;7872512 and8810496 bytes for200-event
cold/trained cells. Replay takes about24-28ms at50 and96-100ms at200. These
files contain1600/6400 measured originals plus their terminals and optional
training history; startup still scans retained history. No bounded-startup,
bounded-disk or storage-format improvement is claimed.

## What follows

Removing duplicate preflight reads has a measurable isolated benefit while
preserving actual originals. It is therefore not merely an apparent speedup
created by dropping work in v55. But v55 remains a failed loaded rescue:
235.352/250.052/276.898ms age p95, elevated write tails, seven dropped observations.

Next use a separately frozen matched-arrival load diagnostic to distinguish
queue pressure and serving contention from isolated per-batch cost. Report
arrival lateness and any producer backlog, not only time inside Recall, so an
overloaded generator cannot hide latency by delaying requests before measurement.
Keep the old closed-loop results; a different offered-load test cannot overwrite
them. Full history/feedback authority, concurrent fitting and real outcome
validation remain open. All seven research directions remain IN PROGRESS.

No whitepaper accuracy claim, production behavior, push or deployment changed.
