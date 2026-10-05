# Group-journal dwell v23: no-wait worsens loaded queueing

Date: 2026-10-01. The [frozen protocol](mmm-recall-dwell-v23-protocol.md)
compared the existing 8 ms research group worker with the same worker at a
1 ns dwell, plus the single guarded SQLite reference. This is a **failed
Goal 6 candidate**, not a production change. The no-wait worker used the
unchanged journal implementation; only its constructor argument differed.

## Full Recall load

Each arm/case completed three rotated trials: 576 exact-as-of Recall offers,
200 distinct nominations per offer, no future packed event, and 768
future-only writes in the writer case. Every group trial closed/reopened with
192 durable acknowledged journal rows; all group batches accounted for 576
rows per arm/case, with zero as-of guard rejections. The focused journal
contract and race test, and `go vet ./internal/service`, passed. The full
load run was **not** race-instrumented or a power-loss recovery test.

| Arm | Quiet offer p99 | Writer offer p99 | Writer call p99 | Writer queue p99 | Writer journal p99 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Single guarded SQLite | 9.25 ms | 202.07 ms | 52.18 ms | 155.41 ms | 24.13 ms |
| Group, 8 ms dwell | 23.77 ms | 570.61 ms | 63.73 ms | 514.82 ms | 29.13 ms |
| Group, 1 ns dwell | 11.69 ms | **1175.76 ms** | 75.11 ms | **1101.55 ms** | 46.46 ms |

The per-trial writer offer p99 values were 212.88, 137.81, and 151.90 ms
for single SQLite; 581.22, 517.29, and 506.52 ms for 8 ms group; and
1212.15, 1152.74, and 1134.17 ms for 1 ns group. The no-wait group made
538 one-entry and 19 two-entry writer batches, versus 117 one-entry,
107 two-entry, 75 three-entry, and 5 four-entry batches with the 8 ms dwell.
Both group arms fail the fixed <100 ms writer offer p99 gate; the 1 ns arm
also fails the no-worse-than-single-reference gate. The opt-in Go test exits
nonzero because it enforces this frozen candidate decision.

## Interpretation

Removing dwell helps quiet request latency, but under matched writes it
approximately doubles the already-bad grouped queue delay. Thus dwell
**alone is not the loaded-latency root cause**. The no-wait schedule also
largely removes batching, increasing transactions and guard acquisitions;
the present data do not separate that increased commit frequency from other
shared-writer contention or Recall worker saturation. The single-journal
reference remains above 100 ms too. A next design must change the
publication/commit ordering or admission-capacity relationship while
retaining every as-of and durable-acknowledgement invariant; it cannot be a
retuning of dwell on this consumed fixture.

Reproduce from the repository root:

```sh
go test -race ./internal/service -run '^TestResearchSQLiteGroupJournalContractV22$' -count=1
go vet ./internal/service
EVENTFRAME_RUN_RECALL_DWELL_V23=1 go test ./internal/service -run '^TestResearchRecallGroupDwellV23$' -count=1 -v
```

The last command intentionally reports `FAIL` on the latency gate after all
trial and durability assertions have passed. At run time, SHA-256 source
hashes were: protocol `37681447ecd70350084753029641f55daed101457f4eef6e4cd16c698b276106`,
v23 harness `4bf91f3301ec462d9e41f4ad17f91d06493915c7a4cba3127182e2429f0ebd61`,
unchanged v22 group worker `e77612716407dbc8f476315f3c5d66e6d951e8813de7c20f9f9aeb04dacc545a`,
unchanged full Recall helper `f295522f00333b6a5f408e100a38d4e1a271415d9a60e0dc99e0e98236111e81`,
and batch as-of guard `b4ad200efcafba1c97bce12a2c27e4ea49e32989f0608f5f9f0ddfbe120bb168`.
All seven whole goals remain open; production and the whitepaper are untouched.
