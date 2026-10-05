# Full-Recall cadence envelope v30: results

Date: 2026-10-01. Research-only Goal 6 screen against the
[frozen protocol](mmm-recall-cadence-v30-protocol.md). Each rate had
one fresh matched learning-off/on pair, 192 full 200-event Recalls
per arm, eight workers, 256 future-only concurrent writes, and
64 durable labels in the enabled arm. The unchanged v29 learner
uses bounded tap drainage and ordered admission/feedback.

| Nominal interval | Actual off/on median gap | Off offer p99 | On offer p99 | On/off ratio | On frontier age p99 | On feedback age p99 | Result |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| 8 ms | 8.000 / 8.000 ms | 55.946 ms | 36.722 ms | 0.656 | 41.633 ms | 21.047 ms | Pass |
| 6 ms | 5.997 / 5.988 ms | 44.348 ms | 37.404 ms | 0.843 | 64.851 ms | 24.681 ms | Pass |
| 4 ms | 3.992 / 3.998 ms | 38.517 ms | 42.094 ms | 1.093 | **401.485 ms** | 26.049 ms | **Fail: freshness** |

All pairs completed their finite exact-nomination, no-future,
as-of, visible-mutation, guarded journal, durable-replay and
phase-conservation checks. Every enabled arm consumed 192 unique
frontiers, published 64 labels and dropped zero tap observations.
The 4 ms arm meets the serving p99 and <=1.10 ratio gates but fails
the frozen <250 ms frontier-to-publication freshness gate. The
test command therefore exits nonzero by design. A separate closed-tap
negative control passed: it returned an error with zero admissions.

This is evidence of a **learning freshness capacity boundary**, not
of serving collapse or data loss, at this fixed workload. Feedback
offer-to-publication stays near 26 ms at 4 ms, so the extra age
accumulates before feedback is offered; phase attribution is a
follow-up, not proven by this table alone. The 6 ms pass is one
screening pair, not a robust throughput guarantee.

Commands:

```sh
go test ./internal/service -run '^TestResearchRecallDecoupledDrainV29ClosedTap$' -count=1 -v
EVENTFRAME_RUN_RECALL_CADENCE_V30=1 go test ./internal/service -run '^TestResearchRecallCadenceV30$' -count=1 -v -timeout 5m
```

At-run SHA-256:

```text
a5edd81dbd856e4cf646eb10372bcc968afd3df0fcaa0cbc8181f5f4a7d48293  docs/experiments/mmm-recall-cadence-v30-protocol.md
9d7246c422bdd6fd07b7185e2decdde301f4bc31c1c91da80f2ea8eb7f4515e2  internal/service/research_recall_cadence_v30_test.go
f93928442b993f7f91d33661668a434f62f51772ecc2c639f7dcc9887b5b0d96  internal/service/research_recall_live_learning_v26_test.go
a2c26f9b4d5e1032359a8987e9e5f7b1a52df273a6110167d73f169290dc269a  internal/service/research_recall_decoupled_drain_v29_test.go
```

Next repeat the 6 ms arm on independent fixtures and profile the
4 ms pre-feedback queue/validation phase. No production or whole
Goal 6 claim follows from this one-pair rate screen.
