# Recall phase-removal v18: journal contention dominates, durability remains open

Date: 2026-10-01. [Frozen contract](mmm-recall-phase-removal-v18-contract.md).
Run host: Apple M4 `arm64`, Go 1.27.1, repository HEAD `1a7edb6`
with a dirty research worktree.
The 200-event factorial completed all 576 Recall offers per arm/cell and 768
future-only writes per writer cell. Every nomination was the exact 200-event
as-of set and no packed candidate was from the future. The diagnostic
in-memory journal sink encoded and retained all 192 distinct journals in
each trial. Graph version remained unchanged in every cache trial.

| Arm | Quiet offer p99 | Writer offer p99 | Writer call p99 | Writer queue p99 | Writer Search p99 | Writer graph p99 | Writer journal p99 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Native durable | 14.60 ms | 854.21 ms | 67.08 ms | 799.17 ms | 16.06 ms | 21.24 ms | 28.12 ms |
| Cached graph only | 15.42 ms | 884.88 ms | 57.87 ms | 831.93 ms | 16.20 ms | <0.001 ms | 25.20 ms |
| In-memory journal only | 4.39 ms | 107.62 ms | 45.78 ms | 71.42 ms | 15.54 ms | 14.26 ms | 0.805 ms |
| Both diagnostic removals | 4.42 ms | 39.73 ms | 33.78 ms | 7.60 ms | 17.04 ms | <0.001 ms | 0.661 ms |

Graph caching alone does not prevent queue growth. Removing the persistent
journal write has a much larger effect, though its writer arm still narrowly
misses the original <100 ms offer gate. Both removals together pass that
numerical gate **only as a non-durable diagnostic**, not as a valid EventFrame
serving design. The quiet in-memory-journal arm also speeds Search and graph
reads, consistent with shared backend pressure, but the experiment does not
identify LibraVDB's internal lock or I/O operation responsible. Marginal
p99s cannot be added together or used as a per-request decomposition.

The next legitimate rescue is a durable journal path whose acknowledgement,
snapshot/as-of validation, conflict semantics, and replay survive future-only
writes and restart without coupling every Recall to the same backend write
path. A research sidecar with a bounded writer gate is a candidate, **not**
yet demonstrated: a two-database design needs explicit uncertain-commit and
reopen rules. Re-run the original bound-worker 200-event gate before claiming
Goal 6 progress from any such implementation. Production remains unchanged.

Reproduce:

```sh
EVENTFRAME_RUN_RECALL_REMOVAL_V18=1 go test ./internal/service -run '^TestResearchRecallPhaseRemovalV18$' -count=1 -v
```

[Raw log](mmm-recall-phase-removal-v18-raw.log) SHA-256
`dd0ad6a3585b045fba4a94b5719f17ee44714e32e79c609064f6e3eec7823fa8`;
contract `ac69bcab587c7bdcf19085d78ba5ab9e6a7c8f0da74f15adcb2be643de276933`;
test source at run `a0857090efd1e7a774166c5123b11c6bd9e5158c7f55b3c553dd04b89dac95c6`.
The full factorial under Go's race detector timed out after 10 minutes in
instrumented LibraVDB HNSW writer work; it is not a race pass or a race
finding. The ordinary-build factorial completed. A separate narrow
`TestResearchRecallProfileInstrumentationConcurrent` race check passed for
the research trace and journal sink; it does not cover LibraVDB's loaded
behavior. [Timeout log](mmm-recall-phase-removal-v18-race.log) SHA-256
`1dfb8cae5e265f33881a9d7ba9887a6e3c9945b5331ceea96d805d79d879dbbf`.
Goal 6 remains open.
