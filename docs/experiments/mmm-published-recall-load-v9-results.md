# Published-LSN full Recall load v9: frozen screen fails

Date: 2026-10-02. The [frozen protocol](mmm-published-recall-load-v9-protocol.md)
ran twice normally and twice in its predeclared race-correctness mode.
Each trial offered 128 visible event writes and 128 full `Service.Recall`
calls at nominal 4 ms cadence, with four Recall workers and one
journaled batch writer. The private 256D collection began with 200
eligible and 17 future events. Every successful Recall was checked
against a runtime-version top-150 oracle and its durable Bayesian
frontier journal.

| Metric | Normal trial 1 | Normal trial 2 | Frozen normal gate |
| --- | ---: | ---: | ---: |
| Writes acknowledged | 128/128 | 128/128 | 128/128 |
| Recalls successful | 118/128 | 115/128 | 128/128 |
| Stale-snapshot rejections | 91 | 96 | Report |
| Search calls, including retries | 209 | 211 | Report |
| Write offer-to-ack p99 | 43.819 ms | 43.432 ms | <250 ms |
| Recall call p99 | 103.022 ms | 99.248 ms | <100 ms |
| Recall offer-to-done p99 | 534.563 ms | 515.847 ms | <100 ms |
| Published-view max age | 28.082 ms | 35.672 ms | <250 ms |

The 10 and 13 failed Recalls exhausted the service's five-attempt
stale-snapshot retry limit after as-of-visible writes landed between
Search and journal commit. Completed Recalls had no observed
version-oracle, durable-journal, future-leak, or acknowledged-before-offer
violation. The much larger offer-to-done tail than call tail is a real
queueing effect: four workers cannot drain the fixed offers while
serial journal publication and retries occupy them. The two runs
formed 26 and 27 event batches, respectively.

In race-correctness mode, instrumentation was **not** held to the
normal timing gates, as predeclared. Both trials still failed semantic
completion: 124/128 Recalls succeeded, four exhausted retries in each,
and all 128 writes were acknowledged. No data race, oracle mismatch,
future leak, or journal mismatch was reported. Instrumented write-age
p99 was 0.978/1.001 s and Recall offer-to-done p99 3.962/4.102 s;
these are not production timings. Ordinary package tests and vet passed.

This is a negative Goal 6 result, not a validation of full serving.
The backend-only loaded v7 and deterministic service v8 components
remain valid at their narrower scopes. Existing
[guarded group-journal v22](mmm-recall-group-journal-v22-results.md)
also failed loaded latency, so repeating its dwell-only batch schedule
is not a supported rescue. A distinct candidate is bounded admission
that protects an individual Recall's read-to-journal interval from
visible event writes while preserving writer freshness and durable
ordering; it needs a separately frozen test. All seven whole goals
remain open. Production was untouched.

Reproduce:

```sh
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V9=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV9$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V9=1 EVENTFRAME_RACE_CORRECTNESS_ONLY=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV9$' -count=1 -v -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

SHA-256: test `7fa5659e1880f2095d4984c93ce30c7e1cc24bddf79e5507a1b998511ab7cab7`;
protocol `0dc7b76b8f1223a5fe461d0d77d2728f1dcfa80ba64e2e8f5cd964ce432d9488`.
