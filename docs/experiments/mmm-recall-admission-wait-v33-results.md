# Ordered admission-wait attribution v33: results

Date: 2026-10-01. Research-only Goal 6 diagnostic against the
[frozen protocol](mmm-recall-admission-wait-v33-protocol.md).
The learner behavior and workload were unchanged. Fresh enabled-only
6 ms and 4 ms fixtures each completed 192 full 200-event Recalls,
64 durable labels, 256 future-only writes, zero drops and all
exact-nomination, no-future, as-of, mutation, journal-reopen,
durable-replay and per-label phase-conservation checks.

| Offer gap | Offer p99 | Frontier age p99 | Journal lookup p50 / p99 | Selected handoff p50 / p99 | Receipt-to-admission p50 / p99 | Guarded admission p50 / p99 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 6 ms | 39.872 ms | 47.159 ms | 1.074 / 6.347 ms | 0.808 / 17.684 ms | 0.002 / 0.012 ms | 15.942 / 21.512 ms |
| 4 ms | 42.331 ms | **412.096 ms** | 1.045 / 2.223 ms | **266.767 / 381.887 ms** | 0.001 / 0.006 ms | 17.977 / 25.235 ms |

The five oldest 4 ms labels were IDs 62, 64, 63, 60, and 59.
Each spent 373–382 ms between successful journal lookup and
coordinator receipt; journal lookup was 0.8–1.5 ms,
receipt-to-admission about 1 microsecond, guarded admission
14–15 ms, and admission-end-to-feedback-offer about 1 ms.
The v33 hypothesis that the **receipt-to-admission** stage
dominates is falsified. The delay is in the bounded selected
handoff channel before receipt. That channel's occupancy can be
explained by serial guarded admissions around 18 ms median
against selected arrivals around every 12 ms at the 4 ms offer
cadence, but this comparison is still a throughput explanation,
not a direct sub-operation profile of the admission guard.

The 6 ms fixture has selected arrivals roughly every 18 ms and
guarded admission median 15.94 ms; the handoff backlog is much
smaller. Enlarging the channel alone would increase tolerated
pending work without reducing frontier age. The next diagnostic
should split guarded admission into source validation, durable
append/read, and post-record validation before a rescue is tried.

Command:

```sh
EVENTFRAME_RUN_RECALL_ADMISSION_WAIT_V33=1 go test ./internal/service -run '^TestResearchRecallAdmissionWaitV33$' -count=1 -v -timeout 5m
```

At-run SHA-256:

```text
77278c8d378e0f5af328b1bf59cdfa6425f38cdd3da6398dfef93121d183cd87  docs/experiments/mmm-recall-admission-wait-v33-protocol.md
fa0765bde3b7d3024de243924c1f7c0b62a72969ca89cd6af83c23f029aadc04  internal/service/research_recall_admission_wait_v33_test.go
979b9aa6d69ee13c7a1ef1ca17d218f1e8ff575c092f375cdf0ab53ddd6af0eb  internal/service/research_recall_decoupled_drain_v29_test.go
d035444aab9ad350008224e71dc12d30baf48d990c7fd3b9bba1024eb387a4e0  internal/service/research_recall_live_learning_v26_test.go
```

No production code changed. This is attribution, not a
performance rescue or whole Goal 6 pass.
