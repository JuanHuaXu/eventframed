# Event-batch throughput and authority boundary v1

This is a Goal 6 feasibility audit, not a serving change or a passing loaded
experiment. The existing raw LibraVDB research batch operation can commit
1..16 same-tenant events in one transaction, but it is not wired through
the research publication/lineage adapter or `Service.Observe`.

## Performance evidence

On Apple M4 with Go 1.27.1/darwin/arm64, Go benchmark
`BenchmarkResearchEventBatch` with three fresh
128-event operations per arm reported 841.8 ms/op for 128 individual writes,
429.9 ms/op for batches of two, 214.5 ms/op for four, 109.1 ms/op for eight,
and 62.9 ms/op for sixteen. The batch-16/single ratio is 0.0747 (about
13.4x throughput in this *isolated* benchmark). It excludes store setup and
close, guard acquisition, durable lineage, `Service.Observe` conversion,
queue dwell, concurrent Recalls and per-event acknowledgement. This is not
an event-loop latency or freshness result.

An earlier [index-scaling study](../../research/public-task-pilot/INDEX_SCALING_RESULTS.md)
at an existing 200-event HNSW corpus found 144.4 ms for sixteen separate
writes versus 10.0 ms for one sixteen-event batch. Thus the newer isolated
number confirms a known backend throughput opportunity rather than a novel
fix. Neither study measured the 4 ms full-Recall learning screen.

## Why direct wiring is invalid

The raw `PutResearchEventBatch` advances runtime version and evidence epoch
by the count of newly inserted events, but returns the same *final* snapshot
for every result. It records each event's available time in the backend
motion table. In contrast:

- `researchpublication.Publisher.Commit` currently accepts an ingestion only
  when runtime version and evidence epoch each advance by exactly one, and
  publishes one motion entry. A batch jump would quarantine the publisher.
- `researchlineage.Ledger.RecordWithMotion` likewise requires one touched
  event and one version step for an ingestion. A batch would fail durable
  lineage recording after a possibly committed backend transaction.
- `store.ResearchSnapshotCompatible` requires complete motion history for
  every intervening version. Publishing only the final batch version would
  make otherwise valid historical as-of reads fail closed.

The current 256-write Goal 6 fixture gives every future-only event the same
`AvailableAt`, so one time value could be recorded at each of the batch's
versions in that fixture. A general batch contract must either require
equal availability times or carry a per-event ordered time vector. Version
assignment must exclude duplicates and be bound to the backend's actual
accepted order; it cannot rely on the final snapshot alone.

## Required next experiment

Before any loaded service comparison, add a research-only batch-aware
publication/lineage transaction that checks a bounded version jump,
atomically publishes **every** accepted event's motion and touch version,
and preserves exact retry versus conflict after uncertain commit. Prove
same-tenant/same-time input validation, duplicate handling, cancellation,
reopen, interleavings with visible backfill, and unknown-outcome quarantine.
Only then connect a bounded `Observe` batch that preserves the existing
post-contract frame conversion, embedding, digest, and per-event identity.
The loaded test must count all 256 writes, individual offer-to-ack p99,
writer total completion, 192 Recalls, 64 durable labels, zero drops,
as-of validation, and frontier age under the unchanged 4 ms gate. A
benchmark speedup alone cannot justify changing authority semantics.

No upstream PR or issue covering this local research batch boundary was
listed when checked; local `HEAD` matched remote `origin/main` at
`1a7edb62b6be4031fd01ebeab8071b17303a7815`. Production remains
untouched. All seven whole research goals remain open.

Reproduce the isolated benchmark:

```sh
go test ./internal/store/libravdbstore -run '^$' -bench '^BenchmarkResearchEventBatch$' -benchtime=3x -count=1 -benchmem -timeout 5m
```

At-run SHA-256:

```text
ac4d79fda3b0f95f78b917676520f2d9085773766af6fdef7ccb0313c72892ec  internal/store/libravdbstore/research_event_batch.go
6d69338578344475391685dc317c0f21da91254a5911da80fc1cf82af9fcbe11  internal/store/libravdbstore/research_event_batch_test.go
96d95be6492ffbfa1080a9657292e0b512c96d61ce8136693dad363bf82db62e  internal/researchpublication/publication.go
036aad4921f0f4cd3ba83073c5da7b132f2b8c7889220d90d0b793cd4b6fe34f  internal/researchlineage/ledger.go
```
