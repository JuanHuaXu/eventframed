# Unchanged deadline-scheduler replication v39: results

Date: 2026-10-01. Research-only Goal 6 replication against the
[frozen protocol](mmm-recall-deadline-replication-v39-protocol.md).
The v38 scheduler and its 200 ms frontier/30 ms writer soft
deadlines were unchanged. Four new rotated 4 ms matched pairs
used fresh stores but the same deterministic synthetic corpus.
Every arm completed 192 full 200-event Recalls, 256 future-only
writes, 64 durable labels, zero tap drops, exact nomination,
no-future/as-of, visible-mutation rejection, journal reopen,
durable replay and phase conservation.

| Pair | Control writer p99 | Deadline writer p99 | Matched ratio | Deadline frontier age p99 | Deadline Recall p99 |
| --- | ---: | ---: | ---: | ---: | ---: |
| 0 | 24.739 ms | 31.978 ms | **1.293** | 221.314 ms | 37.688 ms |
| 1 | 26.678 ms | 34.972 ms | **1.311** | 220.237 ms | 39.219 ms |
| 2 | 25.936 ms | 38.017 ms | **1.466** | 224.607 ms | 42.304 ms |
| 3 | 26.944 ms | 37.275 ms | **1.383** | 223.648 ms | 43.517 ms |

Pooled control/deadline Recall p99 was45.828/39.563 ms
(ratio0.863). Candidate frontier age p99 was223.648 ms and
feedback age p99 22.086 ms. Every candidate trial individually
met the <250 ms frontier criterion. Writer starts during the
offer window were139/160=86.9% of control; summed writer
completion time ratio was1.020. Those pass their frozen gates.
But pooled writer offer p99 rose26.291->35.027 ms, a
**1.332 ratio** above the <=1.25 limit. All four matched
writer-tail ratios also exceed1.25. The unchanged candidate
therefore **fails replication**; the test command exits nonzero
by design.

This is evidence against further deadline tweaking on this
same load shape. It does not prove that every scheduler is
impossible, but v37 priority and v38/v39 deadline scheduling
move enough contention onto writer tails to violate the
declared two-sided contract. A next Goal 6 lead should reduce
guarded-write contention or establish a formally safe
time-partitioned guard, while measuring real writer and learner
costs. The no-writer v36 arm is not an acceptable substitute.

Command:

```sh
EVENTFRAME_RUN_RECALL_DEADLINE_V39=1 go test ./internal/service -run '^TestResearchRecallDeadlineReplicationV39$' -count=1 -v -timeout 5m
go vet ./internal/service ./internal/researchmemory ./internal/researchpublicationstore
go test ./internal/service ./internal/researchmemory ./internal/researchpublicationstore -count=1 -timeout 5m
```

Vet and package tests passed after the replication test was added.

At-run SHA-256:

```text
80a3455edcaea730f17e347b4c8cd9d53151fe0b85e7e136f3ac9d7a2c1e0635  docs/experiments/mmm-recall-deadline-replication-v39-protocol.md
59d1bed15d9a10280f3ffe4e98f808b0dbd866f72caab52555e29eb042aab529  internal/service/research_recall_deadline_replication_v39_test.go
5b13d993b641676ca485cdf687bac153d27170b60a858e6629b4955893a09d18  internal/service/research_recall_deadline_gate_v38_test.go
b1271c99fe480ab491c68a6cf4f578cb47ddc40717f8b42efcac725c69ced66f  internal/service/research_recall_priority_gate_v37_test.go
847c4a34a997f3fecc88e13585bd8c4b5e80e6adfe9fa131e6b630eff849103d  internal/service/research_recall_decoupled_drain_v29_test.go
fda5c431a8257a23a58f2ae9fe93de32794123be9a84d75e2eb1dd69b590bf32  internal/service/research_recall_live_learning_v26_test.go
```

Production code was untouched. All seven whole goals remain open.
