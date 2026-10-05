# Batch reader-factor v1: concurrent writes dominate static growth

The [frozen four-cell diagnostic](mmm-batch-reader-factor-v1-protocol.md)
completed three rotated fresh blocks against real LibraVDB and the private
batch-intent sidecar. Every cell returned exactly 200 past-eligible tenant-A
events on each of 192 searches at nominal 4 ms offers, with zero future or
other-tenant leakage. Where writes were present, all 256 future events and
per-version motion records were confirmed after the run.

| Block | Base static p99 | Expanded A static p99 | Concurrent B writer p99 | Concurrent A writer p99 |
| --- | ---: | ---: | ---: | ---: |
| 0 | 2.903 ms | 2.993 ms | 10.057 ms | 24.938 ms |
| 1 | 2.718 ms | 2.919 ms | 8.675 ms | 20.374 ms |
| 2 | 2.824 ms | 2.986 ms | 9.878 ms | 17.992 ms |
| Pooled | **2.817 ms** | **2.983 ms** | **9.786 ms** | **20.374 ms** |

Relative to pooled base-static p99, preinserting 256 future events into the
searched tenant without concurrent writes costs only **1.059x**. Concurrent
batch writes to a *different* tenant cost **3.474x**, even though tenant-A
search corpus stays at 200. Concurrent same-tenant batches cost **7.233x**.
Median search calls remain about 2.0-2.6 ms; the harm is mainly in the tail.
The writer cells completed in 252-279 ms per arm. Reader offer gaps stayed
near 4 ms median, with p99 around 4.8-5.0 ms.

The source-level mechanism for cross-tenant interference is concrete:
`libravdbstore.Store.Search` holds `eventMu.RLock` over the search, while
`PutResearchEventBatch` holds `eventMu.Lock` over the entire batch including
its transaction and verification. This global lock is independent of tenant.
The active same-tenant write cell adds more delay than the different-tenant
cell, but this diagnostic does not separate collection/index contention,
changing index state, and CPU effects. A static future-heavy corpus alone
does not explain the earlier batch-16 reader regression. This is a
mechanism diagnostic, not a proof that changing the lock would be correct or
faster; read/write snapshot invariants would need their own audit.

The focused test passed, and ordinary `go test -race ./internal/researchbatch`
and `go vet ./internal/researchbatch` passed with this opt-in diagnostic
disabled. Do not promote batch-16 as-is. A next frozen candidate may cap
batch residency or explicitly yield to readers, but it must preserve
durable per-event authority and be tested against the same reader-tail gate
on fresh timing trials. No full `Service.Recall`, background learner or
4 ms freshness claim follows. All seven whole goals remain open; production
is untouched.

Reproduce:

```sh
EVENTFRAME_RUN_BATCH_READER_FACTOR_V1=1 go test ./internal/researchbatch -run '^TestBatchReaderFactorsV1$' -count=1 -v -timeout 5m
go test -race ./internal/researchbatch -count=1 -timeout 3m
go vet ./internal/researchbatch
```

At-run SHA-256:

```text
df8c21fe36e9a0d81af6f36ded8c96bf5c375e69dde5d9fb49d1e7c5f9fc324f  internal/researchbatch/batch_reader_factor_test.go
72507e8d50af52f1682d86c63a94bdc56a56731f7888f2f6b6bbcc82091e0714  docs/experiments/mmm-batch-reader-factor-v1-protocol.md
ac4d79fda3b0f95f78b917676520f2d9085773766af6fdef7ccb0313c72892ec  internal/store/libravdbstore/research_event_batch.go
198a859b3039cabd199d93d975791ae9fe0bfd9fa6dafa6c4dc230ad2fc21974  internal/store/libravdbstore/store.go
```
