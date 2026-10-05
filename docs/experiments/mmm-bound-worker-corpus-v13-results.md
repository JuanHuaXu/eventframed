# Bound worker v13: 50-event corpus breadth

Date: 2026-10-01. [Frozen contract](mmm-bound-worker-corpus-v13-contract.md).
Research-only, production unchanged. Apple M4, Go 1.27.1 darwin/arm64.
The definitive ordinary-build run used the strengthened harness, which checks
the exact set of 50 learner-frontier IDs and excludes future events from each
packed probe result. Three paired trials per cell completed all 192 Recall
offers and 64 durable witnessed labels; writer cells each completed 256
future-only writes. The final race-instrumented run also passed all twelve
cell trials without a detected race or source/frontier failure.

| Live events | Arm | Recall offer p99 | Recall call p99 | Queue p99 | Label-age p99 | Frozen writer gate |
| ---: | --- | ---: | ---: | ---: | ---: | --- |
| 1 | quiet | 10.12 ms | 10.11 ms | 0.034 ms | 10.27 ms | control |
| 1 | writer | 33.45 ms | 33.43 ms | 2.32 ms | 19.34 ms | PASS |
| 50 | quiet | 14.15 ms | 14.14 ms | 0.025 ms | 5.79 ms | control |
| 50 | writer | 81.93 ms | 47.01 ms | 45.81 ms | 21.87 ms | **PASS**, little headroom |

Pooled counts per cell are 576 Recall calls and 192 labels. Writer cells
contain 768 completed future writes. At 50 live events, every learner tap
contained exactly the 50 predeclared IDs, including `seed`, throughout all
64 feedback cycles per trial. The backend runtime version moved 51 to 307
in each writer trial; future writes did not enter the as-of packets. The
50-event writer arm's maximum offer-to-return was 86.46 ms in the definitive
run, close to the 100 ms bound. Call and queue p99 are marginal percentiles
over the same calls and must not be added to reconstruct offer p99.

Two earlier ordinary-build repetitions on the same algorithm had 50-event
writer offer p99 of 84.15 and 92.35 ms, with label-age p99 of 21.86 and
20.57 ms. Those runs lacked the exact 49 non-seed ID assertion and are
supportive timing observations, **not** the definitive as-of identity proof.
That audit gap was found before final reporting, repaired in the harness,
and followed by the fresh passing run above. The final `-race` run was for
correctness only: instrumentation pushed the 50-event writer offer p99 to
3.03 s, mostly queued time, and is not a serving-latency estimate.

Reproduce from the project root:

```sh
EVENTFRAME_RUN_MOTION_CORPUS_V13=1 EVENTFRAME_RESEARCH_PERF_GATE=1 go test ./internal/service -run '^TestResearchMotionBoundWorkerCorpusBreadthV13$' -count=1 -v
EVENTFRAME_RUN_MOTION_CORPUS_V13=1 go test -race ./internal/service -run '^TestResearchMotionBoundWorkerCorpusBreadthV13$' -count=1 -v
go vet ./internal/service
```

Artifacts: [ordinary final log](mmm-bound-worker-corpus-v13-verified.log),
[race final log](mmm-bound-worker-corpus-v13-verified-race.log), and the
[pre-assertion first](mmm-bound-worker-corpus-v13-raw.log),
[repeat](mmm-bound-worker-corpus-v13-repeat.log), and
[race](mmm-bound-worker-corpus-v13-race.log) logs. SHA-256 of the frozen
contract is `5ab36d6ee5783a8c908e38b575b4b6588e8af5dafaed4eb0c3aaa5088dd5d137`;
the final test file is `f62c679e51c972b629e69b84b0c7422035068647d89e79d95ea6ded8b254cc8c`;
the definitive ordinary log is `f8862133a3628c4415e94ea9d829b5172fd599cd2f72c138c8d5c560ddd27dcc`.

This is a synthetic same-topic 50-event frontier, local persistent LibraVDB,
hash embedding, one selected source-bound label per learner query, and one
finite M4 offer rate (~125/s). It is not a 200-candidate or million-event
corpus result, real-text/remote-contract result, population p99 guarantee,
mixed visible-mutation recovery, power-loss proof, or agent-task learning
benefit. Goal 6 remains open. The next discriminating screen is the 200-
candidate ceiling with realistic text/contract costs and matched quiet and
writer controls; the near-100 ms earlier maximum argues against assuming the
50-event pass extrapolates.
