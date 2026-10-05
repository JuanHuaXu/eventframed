# Eight bounded Recall workers pass the v25 loaded-latency component

Date: 2026-10-01. The [frozen protocol](mmm-recall-workers-v25-protocol.md)
compared 4, 8, and 16 Recall workers on the unchanged single guarded SQLite
WAL/FULL journal. The research-only full Recall fixture had 200 live events,
`RecallK=200`, `PackK=10`, 192 offers and 256 concurrent future-only writes
per loaded trial. Arms rotated over three trials, with quiet/loaded order
reversed in the middle rotation. Measured loaded offer rate was
124.93-125.00/s, so improved latency is **not** explained by slowing offers.

Every trial completed 192 offers with exactly 200 distinct live nominations
per offer and no future packed event; each loaded trial completed 256
overlapping future-only writes; every trial reopened with 192 acknowledged
durable journal rows. There were no service errors. The full matrix ran
without race instrumentation; a separate focused eight-worker loaded trial
and the single-journal contract both passed under `-race`, and `go vet`
passed. The focused race timing is not used below.

| Workers | Quiet offer p99 | Loaded offer p99 | Loaded trial p99s | Loaded call p99 | Loaded queue p99 | Guard-wait p99 |
| ---: | ---: | ---: | --- | ---: | ---: | ---: |
| 4 | 9.67 ms | 130.81 ms | 144.58 / 91.01 / 118.85 ms | 49.23 ms | 93.45 ms | 16.87 ms |
| **8** | **9.35 ms** | **48.13 ms** | **48.70 / 45.18 / 57.45 ms** | 48.13 ms | **0.041 ms** | 16.07 ms |
| 16 | 12.39 ms | 50.87 ms | 48.02 / 52.04 / 52.24 ms | 50.86 ms | 0.038 ms | 17.14 ms |

The 8- and 16-worker arms meet the predeclared <100 ms loaded offer p99
criterion both pooled and in each trial, and their pooled quiet p99 is
within 10 ms of the four-worker control. Eight workers are the more
economical **research candidate**: sixteen offered no clear loaded gain and
had one quiet trial at 20.98 ms p99, although its pooled quiet gate passed.
Loaded mean occupied call time remained about 31.8 ms with eight workers,
close to the four-worker 32.3 ms. The much smaller queue, not a dramatic
per-call speedup, accounts for most of the observed offer-latency gain.
That interpretation is consistent with extra workers overlapping blocked
calls; it is not proof that the publication guard scales without limit.

This is a **Goal 6 latency component pass only** on one finite, local,
synthetic workload. It does not establish label-completion freshness,
cross-epoch state transfer, crash/power-loss recovery, a rate/burst matrix,
real-agent usefulness, or production p99. The 8-worker setting is not
promoted to production. Next test burstier and slower offer regimes plus
live-label freshness and visible mutations under the same as-of/durable
contract, with a fresh frozen protocol.

Reproduce:

```sh
EVENTFRAME_RUN_RECALL_WORKERS_V25=1 go test ./internal/service -run '^TestResearchRecallWorkersV25$' -count=1 -v
EVENTFRAME_RUN_RECALL_WORKERS_V25_RACE=1 go test -race ./internal/service -run '^TestResearchRecallWorkersV25FocusedRace$' -count=1 -timeout 5m
go test -race ./internal/service -run '^TestResearchSQLiteJournalContractV20$' -count=1
go vet ./internal/service
```

SHA-256 at run: protocol `501e8e6b9c07fff65e2eb16d7b685a3aa915a1af05d6a05d0f4c3c3545e98fa8`;
matrix harness `6e9e8417cdc63e0128d9a1e79b9da0f50ef29f8fa9ca11237c09b5a1b24d070a`;
focused race harness `27c675756a1506e7ae7a89335ed1c1bda04d871464e633355b70e296d24ddf22`;
single-journal contract `52154dbf940ac30e1bac0768eedcd67d317321f65af25a9b22296c755950cd46`.
All seven whole goals remain open; production and whitepaper are untouched.
