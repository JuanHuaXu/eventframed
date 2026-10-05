# Bounded admission-priority writer scheduling v37: results

Date: 2026-10-01. Research-only Goal 6 intervention against the
[frozen protocol](mmm-recall-priority-gate-v37-protocol.md).
Two fresh matched control/candidate pairs used 4 ms nominal offers,
192 full 200-event Recalls, 256 future-only writes, eight workers,
64 selected durable labels, and the unchanged internal as-of/
durability guard. The test-only scheduler gives waiting admission
or feedback the next external permit after an active writer ends.
Every arm consumed 192 unique frontiers, published 64 labels,
completed 256 writes, dropped zero tap observations and passed
exact nomination, no-future, visible-mutation rejection,
journal reopen, durable replay and phase conservation.

| Trial | Arm | Recall offer p99 | Frontier age p99 | Feedback age p99 | Writer offer p99 | Writer call p99 | Writer total | Writer starts during offers |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | Control | 50.324 ms | 430.434 ms | 26.691 ms | 25.285 ms | 25.285 ms | 3.789 s | 40 |
| 0 | Priority | 38.962 ms | **324.834 ms** | 47.506 ms | **36.102 ms** | 18.137 ms | 3.722 s | 35 |
| 1 | Priority | 38.224 ms | 238.666 ms | 39.363 ms | 38.136 ms | 18.418 ms | 3.914 s | 34 |
| 1 | Control | 46.227 ms | 429.258 ms | 26.187 ms | 25.936 ms | 25.936 ms | 3.762 s | 40 |

Pooled candidate/control Recall-offer p99 ratio was0.783;
candidate frontier age p99 **324.336 ms**, feedback age p99
39.573 ms. Pooled writer offer p99 rose25.379->36.233 ms,
a **1.428 ratio** against the frozen <=1.25 limit. Total
writer completion ratio was1.011 and candidate writer starts
during offers were69/80=86.25% of control, both within their
limits. The frozen component **fails** because pooled freshness
remains above250 ms and writer p99 exceeds its allowance.
The test command exits nonzero by design.

The result is not proof that prioritizing learning is useless:
one candidate trial met freshness and writer total work did not
explode. But strict next-admission priority moves tail wait onto
writer offers without reliably clearing the learner's target.
Enlarging buffers or omitting writes/labels would not rescue
this same contract. A new scheduler must meet both sides on
fresh fixtures under the unchanged 64-label/256-write load.

Commands:

```sh
go test ./internal/service -run '^TestResearchPriorityGateV37OrdersAdmissionFirst$' -count=1 -v
EVENTFRAME_RUN_RECALL_PRIORITY_V37=1 go test ./internal/service -run '^TestResearchRecallPriorityGateV37$' -count=1 -v -timeout 5m
```

At-run SHA-256:

```text
b8598f80a6f709f2ffd36ea8058f87bd867153e0e1eb0b2c4a56c8ed664491ab  docs/experiments/mmm-recall-priority-gate-v37-protocol.md
1de193003c58299005c1126295464b7cb06fa07782f50441ba48e2b765589ce7  internal/service/research_recall_priority_gate_v37_test.go
eeb392921e9d245127727aa04f980efe0f594a3cb33090d435ed295e90b7e45c  internal/service/research_recall_decoupled_drain_v29_test.go
44e4a9c13bccc08cd58227ef2d5dafcc7cf28a933fa3d74032f34133bccc8b62  internal/service/research_recall_live_learning_v26_test.go
```

Production code was untouched; all seven whole goals remain open.
