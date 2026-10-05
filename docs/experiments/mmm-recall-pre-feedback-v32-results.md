# Pre-feedback phase attribution v32: results

Date: 2026-10-01. Research-only Goal 6 diagnostic against the
[frozen protocol](mmm-recall-pre-feedback-v32-protocol.md). The
unchanged v29 learner was instrumented with timestamps only.
One fresh enabled-only 6 ms fixture and one 4 ms fixture each
completed 192 full 200-event Recalls, 64 durable labels,
256 future-only concurrent writes, zero tap drops, exact
nomination/no-future/as-of, mutation rejection, journal reopen,
durable replay, and exact per-label phase conservation.

| Offer gap | Offer p99 | Frontier age p99 | Before admission p50 / p99 | Guarded admission p50 / p99 | Feedback queue p50 / p99 |
| --- | ---: | ---: | ---: | ---: | ---: |
| 6 ms | 37.280 ms | 57.799 ms | 1.859 / 16.954 ms | 16.457 / 23.270 ms | 1.091 / 5.239 ms |
| 4 ms | 41.008 ms | **423.195 ms** | **296.239 / 394.265 ms** | 18.174 / 29.317 ms | 0.967 / 2.240 ms |

The five oldest 4 ms labels were IDs 64, 61, 62, 60, and 59.
Their frontier ages were 413–423 ms. In those same labels,
tap-take-to-admission-start was 383–394 ms, guarded admission
13–15 ms, and admission-end-to-feedback-offer under 1 ms.
Thus the frozen hypothesis that the feedback queue dominates
the 4 ms failure is **falsified**. The backlog is before guarded
admission, while admission itself is a substantial serial service
cost. At 6 ms the median admission time is16.46 ms against an
~18 ms selected-arrival interval; at 4 ms the median is18.17 ms
against an ~12 ms selected-arrival interval. That capacity
comparison is suggestive, not a causal proof by itself.

The before-admission bucket still combines committed-journal
lookup, selected-channel handoff, out-of-order waiting, and
waiting behind earlier admissions. More timestamps are required
before changing journal, ordering or admission behavior.

Command:

```sh
EVENTFRAME_RUN_RECALL_PREFEEDBACK_V32=1 go test ./internal/service -run '^TestResearchRecallPreFeedbackV32$' -count=1 -v -timeout 5m
```

At-run SHA-256:

```text
113b69a6f0795261ded31e3bb317d26befc8b359774399239713aceaa95fe2b8  docs/experiments/mmm-recall-pre-feedback-v32-protocol.md
849bb87bd19e88aee125f9526130de60473736dc204be1794c479f1a84a8c802  internal/service/research_recall_pre_feedback_v32_test.go
23702b44dcc8677a9586b3826cdb319be90f9a0308b0a64b0686de9b93b5a733  internal/service/research_recall_decoupled_drain_v29_test.go
d035444aab9ad350008224e71dc12d30baf48d990c7fd3b9bba1024eb387a4e0  internal/service/research_recall_live_learning_v26_test.go
```

This is phase attribution, not a performance rescue or whole
Goal 6 pass. No production code was changed.
