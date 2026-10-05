# Combined Scheduling And Publication V26: Results

2026-10-03. **Not adopted: all eight original screens FAIL.** Combination
improves outcome publication relative to the paired joined-only controls,
but still misses background freshness and regresses offered Recall beyond100ms.
All seven original whole goals remain OPEN; production and whitepaper untouched.

[Protocol](mmm-combined-witness-v26-protocol.md),
[audit](../../research/combined-witness-v26/audit.json),
[run](../../research/combined-witness-v26/run.json),
[raw](../../research/combined-witness-v26/raw.ndjson),
[technical checks](../../research/combined-witness-v26/technical-prefreeze/checks.json).
1132 source/protocol/runner/checker files frozen before eight NEW normal trials.
1024writes,1024fullRecalls,128distinct mixed outcomes,153600 candidate laws.
Both arms use joined publication; only the fixed scheduler's enablement differs.

| Rep | Motion | Scheduler | Call p99 | Offered Recall p99 | Write p99 | Outcome p99/max | View max | Screen |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | Future | Off | 51.765 | 59.740 | 132.483 | 464.092 | 40.028 | FAIL |
| 1 | Future | On | 55.244 | 151.151 | 81.633 | 360.096 | 30.060 | FAIL |
| 1 | Visible | Off | 49.481 | 61.249 | 99.391 | 460.142 | 29.112 | FAIL |
| 1 | Visible | On | 53.916 | 165.349 | 86.705 | 372.997 | 37.337 | FAIL |
| 2 | Future | Off | 51.649 | 74.599 | 139.981 | 471.241 | 39.084 | FAIL |
| 2 | Future | On | 58.906 | 184.223 | 90.077 | 389.844 | 37.179 | FAIL |
| 2 | Visible | Off | 55.367 | 69.351 | 108.793 | 480.429 | 29.975 | FAIL |
| 2 | Visible | On | 60.204 | 174.636 | 94.100 | 387.284 | 31.071 | FAIL |

Values ms. Original gates unchanged: call AND offered Recall p99<100;
write offered p99<250; outcome offered p99<100/max<250; view max<250.
Outcome p99=max because n=16. Fixed trial order/two repetitions are descriptive,
not random-order population tail guarantees. Collector Go PASS means complete
collection; the independent adoption verdict remains FAIL.

Scheduled future runs serve1022/1038 learned decisions,884/897 cross-epoch,
versus346/348 learned and293/294 cross-epoch control. All16label sources are
eventually used in each scheduled future run versus6/16 controls. Observed
maximum first-use lags371.608/403.544ms versus439.198/453.153ms for control's
uncensored subset. More laws encounter labels partly because the scheduled
backlog drains later. This is NOT improved accuracy or fixed-window freshness.
Visible arms transport zero, ordinary Beta arithmetic error zero, source
epochs unchanged. All144runtime mutations per trial conserved.

Independent frozen auditor passes exact nomination/as-of/packed-law/source/
wire/commit/timing checks, bounded batch-to-chain/ack chronology, grant/count
conservation and shared/exclusive intervals. All129journal bindings per trial
committed in39-43batches; no missing/dropped work. Fourteen corrupted-tape
controls rejected, including grants/overlap, batch IDs/early ack, future/pin/
source/wire/hash errors and fabricated timing/gates. Queues drain, peak9,
maximum8readers. Combined lifecycle/cancellation and existing component race
tests pass; core race3repetitions and vet pass. Prior sealed sources unchanged.

Neither separate mechanism nor their combination meets the target. Next
investigate the forecast-computation handoff boundary and serialized native
metadata/journal cost. Do not retune rotation on this consumed cohort, weaken
gates, hide offered backlog or release leases without a complete dependency
and historical-journal admissibility proof. No generated-agent, broad certificate,
corpus-scaling or whole-goal success claim from this synthetic workload.
