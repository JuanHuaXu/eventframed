# Bound worker v14: 200-candidate ceiling fails under writes

Date: 2026-10-01. [Frozen protocol](mmm-bound-worker-corpus-v14-contract.md).
**FAIL** the predeclared 200-candidate writer-arm Recall p99 <100 ms gate.
The matched 50-candidate control passed. Research-only; production unchanged.
Ordinary Go build on Apple M4, local persistent LibraVDB, hash embedding,
short same-topic synthetic EventFrames, four probe workers, 8 ms paced offers.

| Live events | Arm | Offer p99 | Call p99 | Queue p99 | Label-age p99 | Gate |
| ---: | --- | ---: | ---: | ---: | ---: | --- |
| 50 | quiet | 13.38 ms | 13.37 ms | 0.020 ms | 5.17 ms | control |
| 50 | writer | 89.22 ms | 47.38 ms | 51.48 ms | 22.93 ms | PASS |
| 200 | quiet | 16.66 ms | 16.66 ms | 0.054 ms | 6.79 ms | control |
| 200 | writer | **580.80 ms** | 69.07 ms | 528.07 ms | 24.79 ms | **FAIL Recall** |

Each cell pooled three trials, 576 completed Recall offers and 192 completed
durable source-bound labels; writer cells each completed 768 future-only
writes. All expected 50 or 200 live IDs remained in each learner frontier;
packed probe results excluded future records. The 200 writer trials moved
backend runtime version 201 to 457 and individually reported offer p99
608.63, 573.58, and 569.98 ms. The large queue percentile is a downstream
symptom of insufficient service capacity at this offered rate, not proof of
which internal lock or phase caused it. Marginal call and queue percentiles
must not be added as if they were one request's timeline.

The learner kept feedback fresh while Recall fell behind. A safety-preserving
implementation may need a different scheduling or persistence design, but
that is a hypothesis. Next isolate a 200-candidate probe plus future writer
with **no bound learner** before attributing the contention to the learner's
guard. Do not adopt a narrower frontier as a substitute for the research
objective or relax the latency gate after this failure.

Reproduce from the project root:

```sh
EVENTFRAME_RUN_MOTION_CORPUS_V14=1 EVENTFRAME_RESEARCH_PERF_GATE=1 go test ./internal/service -run '^TestResearchMotionBoundWorkerFrontierCeilingV14$' -count=1 -v
```

The command intentionally exits nonzero because of the frozen 200-cell gate.
[Raw log](mmm-bound-worker-corpus-v14-raw.log) SHA-256:
`81ac2c0af5346a5564eb99988ce79fc267a628dbe3526e773f05e54053130469`.
Frozen protocol SHA-256:
`07add373153b65d7d91f08b8830be012dd41a97c6d4a28e08fefb9bb7f5ff340`.
Test-file SHA-256 at the run:
`73d3a5a75b693da2feb0e1338eb9cf595928697b6d85ed5e478a8b8e82abff8f`.
This screen is not a remote contract, real-text, agent-answer, population p99,
or million-event corpus result. Goal 6 remains open.
