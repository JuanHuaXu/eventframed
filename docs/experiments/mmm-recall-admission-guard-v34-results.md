# Guarded admission sub-operation profile v34: results

Date: 2026-10-01. Research-only Goal 6 diagnostic against the
[frozen protocol](mmm-recall-admission-guard-v34-protocol.md).
Fresh enabled-only 6 ms and 4 ms fixtures each completed
192 full 200-event Recalls, 64 durable labels, 256 future-only
writes, zero drops and all exact nomination/no-future/as-of,
mutation, journal-reopen, replay and phase-conservation checks.

| Offer gap | Frontier age p99 | Guard-entry + source validation p50 / p99 | Durable admit p50 / p99 | Durable readback p50 / p99 | Post-record validation p50 / p99 |
| --- | ---: | ---: | ---: | ---: | ---: |
| 6 ms | 62.344 ms | **15.713 / 22.517 ms** | 0.152 / 0.531 ms | 0.026 / 0.193 ms | 0.917 / 1.641 ms |
| 4 ms | **396.108 ms** | **16.809 / 24.785 ms** | 0.136 / 0.489 ms | 0.023 / 0.184 ms | 0.816 / 1.780 ms |

Guard exit was about 1 microsecond median in both arms. The five
oldest 4 ms labels each spent 12.9–13.6 ms before callback entry,
under 0.11 ms in durable admit, under 0.03 ms in readback and
0.76–0.79 ms in post-record validation. Each label's five
sub-operations summed exactly to its guarded-admission duration.

The lead that SQLite append or duplicate post-record validation
is the principal serial admission cost is **falsified**. The
dominant pre-callback bucket includes waiting to enter the
mutation guard, its snapshot/lineage check, and the first
candidate/source validation. Those must be split before a
rescue. Removing the second validation alone cannot plausibly
recover the >6 ms per-selected-label throughput gap at 4 ms.

Command:

```sh
EVENTFRAME_RUN_RECALL_ADMISSION_GUARD_V34=1 go test ./internal/service -run '^TestResearchRecallAdmissionGuardV34$' -count=1 -v -timeout 5m
```

At-run SHA-256:

```text
47701b6312fe8a7a7d1fc54e00ca88e0ff9fcb2836747c2a403e90898a73924e  docs/experiments/mmm-recall-admission-guard-v34-protocol.md
20eec380718c6c28e70a32db110ff35a67dbc84e4a96a0e49ccc024a17fbf570  internal/service/research_recall_admission_guard_v34_test.go
49fc2616aaf7d0b031ab1a0e15d4051e9b1bf9865c5aac6887d3af65173a488b  internal/service/research_recall_decoupled_drain_v29_test.go
cb6328f538ac6894d7089edf059c37cb182b8069fa8e4b9c6b6c3f4ca0f52e94  internal/service/research_recall_live_learning_v26_test.go
```

No production code changed. This is attribution, not a
performance rescue or whole Goal 6 pass.
