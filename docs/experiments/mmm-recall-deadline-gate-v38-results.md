# Deadline-aware writer/admission scheduling v38: results

Date: 2026-10-01. Research-only Goal 6 candidate against the
[frozen protocol](mmm-recall-deadline-gate-v38-protocol.md).
Two new matched 4 ms control/deadline pairs kept the v37
256 future-only writes, 192 full 200-event Recalls, eight
workers, 64 selected bound durable labels and unchanged
internal as-of/durability guard. Unit checks passed both
writer-first and old-frontier-first deadline arbitration.
Every full-load arm completed 256 writes and 64 labels,
consumed 192 unique frontiers, dropped zero tap observations,
and passed exact nomination, no-future, visible-mutation
rejection, journal reopen, durable replay and phase conservation.

| Trial | Arm | Recall offer p99 | Frontier age p99 | Feedback age p99 | Writer offer p99 | Writer call p99 | Writer total | Writer starts during offers |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | Deadline | 37.285 ms | 223.289 ms | 22.943 ms | 31.037 ms | 17.988 ms | 3.674 s | 35 |
| 0 | Control | 52.563 ms | 436.014 ms | 26.538 ms | 25.017 ms | 25.017 ms | 3.752 s | 40 |
| 1 | Control | 46.997 ms | 433.665 ms | 27.210 ms | 24.987 ms | 24.987 ms | 3.743 s | 40 |
| 1 | Deadline | 36.083 ms | 220.896 ms | 28.843 ms | 35.001 ms | 17.999 ms | 3.696 s | 35 |

The pooled deadline/control Recall-p99 ratio was0.792;
candidate frontier age p99 **220.896 ms** and feedback age
p99 22.943 ms. Those pass the frozen serving and freshness
gates. Writer total completion ratio was0.983 and candidate
writer starts during offers were70/80=87.5% of control,
also passing. However pooled writer offer p99 rose
24.987->31.947 ms, a **1.279 ratio**, above the frozen
<=1.25 limit. The candidate therefore **fails** the full
two-sided component screen; the test command exits nonzero.

This is a near miss, not validated Goal 6 throughput.
Deadline arbitration is materially better balanced than v37's
strict admission priority, but moving even 2.9 percentage
points beyond the allowed writer-tail increase matters under
the predeclared contract. Retuning the 30/200 ms deadlines
on this cohort would overfit a timing sample. A fresh
architecture-level lead is needed, such as a guarded-write
throughput improvement or a proved time-partitioned guard,
with writer and learner costs both measured.

Commands:

```sh
go test ./internal/service -run '^TestResearchDeadlineGateV38OrdersByDeadline$' -count=1 -v
EVENTFRAME_RUN_RECALL_DEADLINE_V38=1 go test ./internal/service -run '^TestResearchRecallDeadlineGateV38$' -count=1 -v -timeout 5m
go vet ./internal/service ./internal/researchmemory ./internal/researchpublicationstore
go test ./internal/service ./internal/researchmemory ./internal/researchpublicationstore -count=1 -timeout 5m
go test -race ./internal/service -run '^(TestResearchDeadlineGateV38OrdersByDeadline|TestResearchPriorityGateV37OrdersAdmissionFirst)$' -count=1 -v -timeout 2m
EVENTFRAME_RUN_RECALL_DECOUPLED_V29_RACE=1 go test -race ./internal/service -run '^TestResearchRecallDecoupledDrainV29FocusedRace$' -count=1 -v -timeout 5m
```

Vet, package tests and the deadline/priority unit race checks
passed after the interface change. The separate focused live
learning `-race` run also passed with 192 frontiers, 64 labels,
zero drops and no reported Go data race. Race-build timings are
excluded from ordinary performance comparisons.

At-run SHA-256:

```text
16145e069fbaabf06754535e5b64d300d13a18d248bc91ac26258e4fe4a592ef  docs/experiments/mmm-recall-deadline-gate-v38-protocol.md
5b13d993b641676ca485cdf687bac153d27170b60a858e6629b4955893a09d18  internal/service/research_recall_deadline_gate_v38_test.go
b1271c99fe480ab491c68a6cf4f578cb47ddc40717f8b42efcac725c69ced66f  internal/service/research_recall_priority_gate_v37_test.go
847c4a34a997f3fecc88e13585bd8c4b5e80e6adfe9fa131e6b630eff849103d  internal/service/research_recall_decoupled_drain_v29_test.go
fda5c431a8257a23a58f2ae9fe93de32794123be9a84d75e2eb1dd69b590bf32  internal/service/research_recall_live_learning_v26_test.go
```

Production code was untouched. All seven whole goals remain open.
