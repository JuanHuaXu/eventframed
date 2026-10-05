# Batch intent prototype v1: real-store recovery component

The [frozen protocol](mmm-batch-intent-prototype-v1-protocol.md) passes its
finite test-only storage controls against real LibraVDB and a private SQLite
WAL/FULL sidecar. This is a Goal 6 **component** result, not a production
integration or a loaded serving/freshness pass.

## Authority controls

`internal/researchbatch/batch_intent_prototype_test.go` implements a private
batch intent, ordered per-event motion/accepted rows, and reopen reconciliation.
It does not modify `Service.Observe`, the production backend constructor,
`researchpublication.Publisher`, or `researchlineage.Ledger`.

The focused test suite covers crashes after intent persistence, raw backend
commit, and sidecar finalization. Reopen discards an uncommitted intent only
when the backend is still at the before snapshot and proposed new IDs are
absent; it finalizes an already committed batch only after exact event and
vector readback. Exact retry preserves version counts, while changed digest
or payload, malformed/altered intent, unexpected backend advance, missing
sidecar, and an attempted new sidecar over existing backend history fail
closed. Mixed duplicate/new writes retain ordered touch versions. Future-only
motion preserves historical as-of reads; visible backfill invalidates them.
Two additional failing-then-fixed regressions cover an ordinary Go monotonic
clock on exact retry and forged duplicate authority from a tampered accepted
sidecar row. Duplicate-only acknowledgement now checks the backend digest
through its raw batch operation; persisted event/vector comparisons use the
serialized value rather than `time.Time` internals.

`go test -race ./internal/researchbatch -count=1 -timeout 3m` and
`go vet ./internal/researchbatch` pass. This establishes behavior under
injected process-phase interruptions, not atomicity under hardware power loss,
concurrent owners, arbitrary availability times, or malicious tampering with
both databases.

## Isolated timing

Apple M4, Go 1.27.1, darwin/arm64. Each benchmark operation writes 128
same-tenant events to a fresh store. Setup and close are outside the timer.
Each arm used three operations per sample and three independent samples;
the table reports the median `ns/op` sample converted to milliseconds.

| 128-event operation | Raw backend | Batch intent + sidecar |
| --- | ---: | ---: |
| batch size 1 | 845.65 ms | 872.31 ms |
| batch size 4 | 213.53 ms | 230.10 ms |
| batch size 16 | 63.64 ms | 70.69 ms |

At batch size 16, the test-only authority layer adds about 11.1% to raw
batch time. Sidecar-protected batch-16 is about 12.34x faster than
sidecar-protected singles in this isolated benchmark. This measures no queue
dwell, concurrent Recall, embedding/conversion, per-event acknowledgement,
learner freshness, or full-request p95/p99. Allocation at batch-16 rises from
about 8.84 MB/13.1k allocations raw to 11.35 MB/29.3k with the sidecar.
The benchmark is a throughput lead, not a pass of the 4 ms Goal 6 screen.

## Next boundary

Before a shared-runtime candidate, establish one owner across processes,
durable uncertain-commit recovery, per-event acknowledgement semantics,
`Service.Observe` conversion parity, and loaded writer/learner plus Recall
latency under the unchanged Goal 6 gate. The separate sidecar is a prototype
of authority accounting, not a recommendation to add another production DB.
All seven whole research goals remain open; production is untouched.

Reproduce:

```sh
go test -race ./internal/researchbatch -count=1 -timeout 3m
go vet ./internal/researchbatch
go test ./internal/researchbatch -run '^$' -bench '^BenchmarkBatchIntentPrototype$' -benchtime=3x -count=3 -benchmem -timeout 5m
```

At-run SHA-256:

```text
c339ff6c985202fc9dcbde7a9280677ce290285125148f9e6ab35b8c172f66fe  internal/researchbatch/batch_intent_prototype_test.go
ec7c0a67a245fda9740106b37dc1dc1d6924e2d3179abebdaeeaf419e6fe54cd  internal/researchbatch/batch_intent_prototype_run_test.go
983af3bb6e9f4b3e92ad15567e7e3ecc8d8b0c476dcc19f25bd987f7685ed978  docs/experiments/mmm-batch-intent-prototype-v1-protocol.md
ac4d79fda3b0f95f78b917676520f2d9085773766af6fdef7ccb0313c72892ec  internal/store/libravdbstore/research_event_batch.go
```
