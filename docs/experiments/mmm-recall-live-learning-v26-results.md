# Full Recall plus durable live learning v26: freshness fails

Date: 2026-10-01. The [frozen v26 protocol](mmm-recall-live-learning-v26-protocol.md)
joined the v25 eight-worker, 200-event full Recall workload with a bounded
queue-64 research consumer and durable feedback. Every third committed
frontier supplied one **synthetic lifecycle label**, 64 per enabled trial.
The fixture does not validate prediction quality or authenticate real-world
truth. Both arms used guarded SQLite WAL/FULL journals and 256 overlapping
future-only event writes. Three rotated isolated trials completed in an
ordinary build.

| Trial | Off offer p99 | On offer p99 | On frontier-to-published p99 | On feedback-offer-to-published p99 |
| ---: | ---: | ---: | ---: | ---: |
| 0 | 50.70 ms | 34.82 ms | 410.82 ms | 21.04 ms |
| 1 | 53.10 ms | 34.39 ms | 407.44 ms | 18.93 ms |
| 2 | 47.68 ms | 35.19 ms | 411.65 ms | 22.09 ms |
| Pooled | 49.10 ms | 34.98 ms | **410.82 ms** | 21.04 ms |

The enabled arm satisfies the frozen <100 ms serving gate and the
off-relative 1.10 ceiling, but **fails** the <250 ms frontier-to-publication
freshness gate. The shorter feedback-offer clock does not rescue it: most
age accrues before feedback is offered, in the tap-wait/admission path.
The apparent enabled-arm serving speedup is a cross-run observation, not
evidence that learning makes Recall faster. Actual offer gaps remain near
8 ms in both arms.

Every ordinary trial completed 192 offers, exactly 200 distinct live
nominations per offer, no future packed event, 256 overlapping future-only
writes, 192 durable Recall journals after close/reopen, and zero tap drops.
Enabled trials published all 64 labels with no failed/pending work;
128 ordered admit/feedback rows per trial replayed successfully. A final
visible mutation made an older bound admission fail without executing its
callback. The ordinary test exits nonzero solely because of the predeclared
freshness gate. `go vet ./internal/service ./internal/researchmemory` passes.

## Ordering failure under race instrumentation

A separate one-trial `-race` run did **not** complete the protocol: after
seven consumed frontiers and two labels it rejected admission 3 with
`invalid or stale frozen prediction`; the stopped consumer then dropped
121 frontiers. No Go data-race report appeared, but this is **not a passed
race test**. The existing [adapter timing test](../../internal/researchmemory/adapter_test.go)
confirms that a prediction earlier than previously admitted feedback must
be rejected. `Recall` taps a frontier after journal commit, so eight
concurrent Recalls can enqueue in **completion order**, not necessarily
ascending `AsOf` order. That is the leading explanation for this specific
backdating rejection, not yet a separately instrumented tap-order trace.
Do not bypass the check or backdate the learner. A successor must declare a
bounded event-time ordering/watermark policy and measure its added age and
drops under both ordinary and race-instrumented load.

The v26 result does not establish Goal 6: freshness fails even in the
ordinary build, and the race-stressed lifecycle is not robust. It does not
cover visible-mutation **state transfer**, power-loss recovery, real labels,
or production/OpenClaw latency. Production and the whitepaper are untouched.
The later [v27 phase diagnostic](mmm-recall-live-phases-v27-results.md)
instrumented this fixture without changing its update policy; its source
hash differs from the original v26 at-run hash below.

Reproduce from the repository root:

```sh
EVENTFRAME_RUN_RECALL_LIVE_V26=1 go test ./internal/service -run '^TestResearchRecallLiveLearningV26$' -count=1 -v
EVENTFRAME_RUN_RECALL_LIVE_V26_RACE=1 go test -race ./internal/service -run '^TestResearchRecallLiveLearningV26FocusedRace$' -count=1 -timeout 5m
go test ./internal/researchmemory -run '^TestJournalTimingAndDuplicate$' -count=1
go vet ./internal/service ./internal/researchmemory
```

The ordinary command is expected to fail the frozen age gate on the recorded
workload. The race-stressed backdating failure is schedule-dependent and may
not recur in every replay; a successful replay would not erase the observed
failure. SHA-256 at run: protocol
`79e57325a7189418f3afc47c07790e6ff5eca7d2aef23a3cec4408a39ca5f28f`;
v26 harness `729b059faa927a35488e9c25ee081331fb0152d9a42e2cdadc28cb3033234781`;
focused race harness `cffabb4ea78d9ef38d31f0ed526d7221ce80956caf2d9e12ba7c71a384d09e1b`;
frozen prediction rule `28437acf691f0e4a3b9346da66dacb57800d2991b7b641191fad8ad3d45eb34e`.
