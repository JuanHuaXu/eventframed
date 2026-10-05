# Recall store-phase profile v17: multiple writer-sensitive phases

Date: 2026-10-01. [Frozen contract](mmm-recall-store-profile-v17-contract.md).
Run host: Apple M4 `arm64`, Go 1.27.1, repository HEAD `1a7edb6`
with a dirty research worktree.
The paired, no-learner full-service diagnostic completed every offer, future
write, and exact as-of nomination. At 200 live events, loaded offer p99 is
836.55 ms, mostly queue wait; the same fixture is 15.15 ms without writers.
This reproduces v15's failure shape while naming the store phases.

| Live events | Arm | Offer p99 | Call p99 | Queue p99 | Search p99 | Graph p99 | Journal p99 | Residual batch span p99 |
| ---: | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 50 | quiet | 14.18 ms | 14.18 ms | 0.024 ms | 5.54 ms | 3.24 ms | 8.50 ms | 4.12 ms |
| 50 | writer | 284.63 ms | 45.92 ms | 248.10 ms | 13.66 ms | 14.15 ms | 22.87 ms | 0.35 ms |
| 200 | quiet | 15.15 ms | 15.15 ms | 0.018 ms | 6.13 ms | 4.26 ms | 9.97 ms | 4.50 ms |
| 200 | writer | 836.55 ms | 72.09 ms | 783.32 ms | 15.90 ms | 20.27 ms | 25.63 ms | 0.67 ms |

Each cell has 576 Recall offers; writer cells complete 768 future-only
writes. Search, graph read, and journal put each occur once per Recall, with
no observed Recall retry. Residual reads occur 50 or 200 times per Recall,
but run concurrently. Their summed durations are **not** an exclusive phase
time and cannot be added to the other columns. Optional Anti-Pigeon and
posterior reads are absent in this fixture. The 200-event writer Snapshot
p99 also reaches 15.83 ms, but is near zero in the other cells. Marginal
p99s are not per-request additive or a causal attribution.

The writer slows several store-boundary operations and reduces full-service
throughput below the 8 ms offer rate. The evidence does **not** identify one
sole cause or authorize dropping durable journals. Next use a matched
research-only phase-removal diagnostic (including the native arm) to test
whether graph reads, journal persistence, or their combination is necessary
for the queue growth. Any such removal changes semantics and cannot be
reported as a passing production design.

Reproduce:

```sh
EVENTFRAME_RUN_RECALL_PROFILE_V17=1 go test ./internal/service -run '^TestResearchRecallStorePhaseProfileV17$' -count=1 -v
```

[Paired log](mmm-recall-store-profile-v17-paired.log) SHA-256
`70f40a831f238b53cfdfd294f1210dc8b50ed3f2517c4ec2343ccbfe92ec8761`;
contract `a4f01e3dc082d7a6499221514e6b1b62970029ff1a9d8464d171daafdfe58212`;
test source at run `739ec472c6bafae3bf9a9aab7cfe93e26ca158262d1993b1126e1246821dabe2`.
An earlier unpaired run and two harness panics are retained as diagnostic
history, not counted as result evidence. Production is unchanged; Goal 6 open.
