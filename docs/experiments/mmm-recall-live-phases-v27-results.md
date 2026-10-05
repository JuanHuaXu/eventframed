# Live-learning phases v27: tap backlog and confirmed backdating

Date: 2026-10-01. The [frozen phase protocol](mmm-recall-live-phases-v27-protocol.md)
instrumented the same research-only, eight-worker, 200-event, queue-64,
64-label durable fixture as failed v26. Three ordinary enabled trials
completed all 192 Recalls, 256 future-only writes, 64 labels, journal
reopen/replay and visible-mutation rejection each. The diagnostic test
passes because it checks phase integrity, **not** the still-failed v26
250 ms freshness gate. `go vet ./internal/service ./internal/researchmemory`
passes.

| Phase, pooled 192 labels | p50 | p99 |
| --- | ---: | ---: |
| Tap enqueue to consumer take | **139.13 ms** | **418.23 ms** |
| Consumer take to guarded feedback offer | 15.60 ms | 22.78 ms |
| Feedback offer through guarded return | 14.13 ms | 20.23 ms |
| Guarded return to observed publication | 0.0047 ms | 1.44 ms |
| Full committed-frontier to publication |  | **456.67 ms** |

The three frontier-age p99 values were 396.23, 457.14, and 404.55 ms;
serving offer p99 was 34.68, 38.62, and 34.59 ms. For **each label**, the
test checks exact conservation of the four measured intervals against
frontier age. Percentiles across different phases are not additive. This
locates the ordinary freshness failure primarily at the tap queue, not at
worker publication. The consumer performs a selected label's validated
admission and guarded feedback serially; their pooled medians sum to about
29.7 ms while one label is selected every third ~8 ms offer, a plausible
capacity mismatch. This comparison of medians is an inference, not a
measured steady-state service-rate theorem.

A focused `-race` replay failed after two labels and seven consumed
frontiers. The third selected frontier had `AsOf=2026-09-12T00:00:12Z`,
but the previous accepted label was available at
`2026-09-12T00:00:15Z`; admission rejected the backdated prediction and
the stopped consumer then dropped 121 frontiers. This **confirms** that
completion-order delivery can regress in event time under this workload.
There was no Go data-race report, but the race-stressed lifecycle test did
not pass. A later clean run would not erase the observed violation.

## Decision

Do not relax the learner's no-backdating check, reinterpret feedback age as
frontier age, or merely move the queue. A rescue needs a bounded event-time
ordering/watermark policy **and** enough validated admission/feedback
throughput to keep frontier-to-publication p99 below 250 ms at the same
64-label/192-request workload. It must account for skipped, late, and
never-arriving frontiers without using feedback from the future to forecast
the past. The existing v25 serving-latency component still stands; Goal 6
is open, and production/whitepaper remain untouched.

Reproduce:

```sh
EVENTFRAME_RUN_RECALL_LIVE_PHASES_V27=1 go test ./internal/service -run '^TestResearchRecallLivePhasesV27$' -count=1 -v
EVENTFRAME_RUN_RECALL_LIVE_PHASES_V27_RACE=1 go test -race ./internal/service -run '^TestResearchRecallLivePhasesV27FocusedRace$' -count=1 -timeout 5m -v
go vet ./internal/service ./internal/researchmemory
```

The second command's backdating failure is schedule-dependent. SHA-256 at
run: protocol `dacd4fe6f68d4eace2c6bb163f20112ac598c464706b772894b7e627ddb89314`;
instrumented v26 fixture `c0203de7deaafa88e3ccefd98f0cd7816a0fd4167ae402ebe9923f6cdb3e5c85`;
v27 assertions `ccae58f7565f2641dcd2858980a9a09dd688fcae4549a971468c49599d417fb3`.
