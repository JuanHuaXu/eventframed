# Bounded ordered learning pipeline v28: results

Date: 2026-10-01. Research-only Goal 6 component; no production change.
The [frozen protocol](mmm-recall-ordered-pipeline-v28-protocol.md) selects
offered indices 0,3,...,189, orders selected admissions by `AsOf`, and
pipelines feedback behind admission. It retains the v26 full-Recall fixture:
200 live events, 192 offers at a nominal 8 ms cadence, eight workers,
256 concurrent future-only writes, a queue-64 frontier tap, one guarded
SQLite WAL/FULL journal, and 64 bound durable lifecycle labels per trial.

## Ordinary matched trials

The opt-in test completed three rotated learning-off/on pairs. Each
enabled trial consumed 192 committed frontiers, published 64 labels,
and dropped zero tap observations. Durable journal reopen, 128-row
ledger/replay, exact nomination, as-of exclusion, visible-mutation
rejection, and phase checks passed.

| Trial | Off offer p99 | On offer p99 | On frontier-to-published p99 | On feedback-to-published p99 |
| --- | ---: | ---: | ---: | ---: |
| 0 | 48.971 ms | 36.871 ms | 42.932 ms | 20.509 ms |
| 1 | 48.323 ms | 35.708 ms | 38.416 ms | 20.415 ms |
| 2 | 51.587 ms | 38.051 ms | 43.315 ms | 22.363 ms |
| Pooled | 48.492 ms | 36.871 ms | 42.932 ms | 20.509 ms |

The pooled on/off offer-p99 ratio was 0.760. Thus this finite ordinary
component passed the frozen <100 ms per-trial/pooled serving gate,
the <=1.10 on/off ratio, and both <250 ms pooled freshness gates.
These timings do not include Go race instrumentation.

## Focused race result

The separately frozen `-race` lifecycle check **failed**:

```text
live consumer seen=137 labels=35 dropped=55 err=await selected frontier 105: research frontier tap closed
```

No Go data race was reported before the failure, but the run did not
complete the required 192 frontiers/64 labels/zero drops. The ordered
consumer blocks tap drainage while it validates and durably admits each
selected frontier. Under race-build slowdown the queue overflows; a
selected frontier is then missing. This is the leading mechanism from
the observed consumer structure and drop count, not a proven general
production root cause. Unlike v26/v27, this run did not reach a
backdated-admission error. It cannot establish race-stressed lifecycle
correctness or a Goal 6 pass.

Commands:

```sh
EVENTFRAME_RUN_RECALL_ORDERED_V28=1 go test ./internal/service -run '^TestResearchRecallOrderedPipelineV28$' -count=1 -v
EVENTFRAME_RUN_RECALL_ORDERED_V28_RACE=1 go test -race ./internal/service -run '^TestResearchRecallOrderedPipelineV28FocusedRace$' -count=1 -timeout 5m
go vet ./internal/service ./internal/researchmemory
```

`go vet` passed. At-run SHA-256:

```text
220a2255ba4e6655a84452e0cbecf861ee972ee5041b3350df12a48edda4bd1c  docs/experiments/mmm-recall-ordered-pipeline-v28-protocol.md
5fd0470b957e5b3ed251f34f999157168a60ec2cbfd0c72cca17d7b9989c923f  internal/service/research_recall_ordered_pipeline_v28_test.go
44101187924a1d70a8f6b3d266b2cdd1012b32a9bbdbe683dbb907d5737846ea  internal/service/research_recall_live_learning_v26_test.go
```

Next test a separately bounded tap drainer feeding ordered admission,
without changing the queue-64 boundary or hiding over-capacity loss in
an unbounded buffer. The ordinary component pass is not full Goal 6:
arbitrary lateness, missing frontiers, cross-epoch transfer, power-loss
recovery, real outcomes, and OpenClaw serving remain unverified.
