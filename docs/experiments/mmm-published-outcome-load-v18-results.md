# Loaded published-outcome v18: scored-learning gate fails

Date: 2026-10-02. The frozen [protocol](mmm-published-outcome-load-v18-protocol.md)
**fails** in both normal feedback trials. The test-only, single-owner
pipeline durably published 16/16 distinct-event full-stream outcomes
while completing 128/128 event writes and 128/128 full Recalls, with
zero stale, future, top-150-oracle, journal/pin or marker violations.
It did **not** deliver an updated posterior into any loaded scored
Recall. This is a Goal 6 negative result, not a latency or learning pass.

| Trial/arm | Write p99 | Recall call p99 | Recall offer p99 | Outcome publish p99 | Certified Recalls | Recalls offered after first outcome, certified | Scored learned candidates |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1, no feedback | 65.72 ms | 46.03 ms | 54.07 ms | n/a | 5/128 | n/a | 0 |
| 1, feedback | 109.55 ms | 48.32 ms | 48.84 ms | 92.04 ms | 7/128 | 0/121 | **0** |
| 2, no feedback | 65.07 ms | 47.63 ms | 51.77 ms | n/a | 4/128 | n/a | 0 |
| 2, feedback | 84.73 ms | 49.43 ms | 49.43 ms | 85.07 ms | 6/128 | 0/122 | **0** |

All observed normal-arm write p99 values were below the 250 ms gate,
Recall call and offer p99 below 100 ms, outcome p99/max below
100/250 ms, and published-view maximum age below 250 ms. Those
component timings do not rescue the failed scored-learning criterion.
The feedback arm delivered its 16 labels from one committed journal
on distinct nominated events; labels are synthetic and do not measure
predictive quality or independent evidence validity.

The mechanism is explicit in the current store contract:
`putResearchEventBatch` increments `EvidenceEpoch` for every new event.
Selection and omitted-influence certificates bind to one exact epoch.
The initial synthetic certificates were valid at epoch 217, but the
128 loaded writes advanced the final epoch to 345. The early 4-7
Recalls still carried certification; none offered after the first
outcome did. `Service.Recall` therefore correctly withheld the
posterior from its scored law. Making the old certificates appear
valid would invalidate their coverage claim. A future candidate must
justify and cost fresh, epoch-bound evidence or a sound incremental
certificate, not merely bypass this gate.

The initial harness run was discarded: certificates were issued before
seeding and were stale before load. The next run exposed another
harness error: pre-seeded vectors had norm two because the normalizing
append helper was bypassed, corrupting the independent top-150 oracle.
After moving certificate issuance after seeding and normalizing all
vectors, the no-feedback oracle and both feedback arms had zero oracle
violations. These are fixture corrections, not post-hoc model rescues.

Reproduce the frozen final-state test with:

```sh
EVENTFRAME_RUN_PUBLISHED_OUTCOME_LOAD_V18=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedOutcomeLoadV18$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_OUTCOME_LOAD_V18=1 EVENTFRAME_RACE_CORRECTNESS_ONLY=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedOutcomeLoadV18$' -count=1 -v -timeout 5m
```

The normal command exits nonzero only for the scored-learning gate;
its control arms pass. The correctness-only race command now passes
after exempting the normal-time scored-learning assertion when no
post-feedback Recall is offered. This changes no normal-run v18
criterion. Under race instrumentation, outcomes completed after the
Recall offers and timing exceeded the normal gates; those timings
are not used as performance estimates.
No production code or active serving path changed. All seven whole
research goals remain open.
