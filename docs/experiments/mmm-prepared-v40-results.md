# Transaction-scoped write statements v40 results

Status: PASSED the finite isolated-ledger screen. This warrants loaded wrapper
integration, not a serving-latency, learned-model or direction-level success.

## Mechanism and contract

`AppendBatchPrepared` prepares the existing lookup, admission-existence and
insert SQL once per transaction. The original `AppendBatch` remains the default
query-per-record control. Both share input bounds, identity/JSON validation,
ordering, exact-byte retry checks, all-or-nothing commit and acknowledgment
semantics. Statements are closed on return/panic and cannot survive a transaction.
No FULL durability, readback, terminal or authority checks were removed.

The pinned modernc.org/sqlite v1.57.0 `stmt.go` caches single-statement native
handles in `newStmt`, then reuses them in execution. Thus this experiment tests
actual statement reuse rather than merely a renamed query wrapper. Earlier
profiles did not establish parsing as dominant; this paired experiment supplies
direct evidence of benefit at the tested storage boundary only.

Prepared/control acknowledgment and replay parity, late-conflict rollback,
missing-admission rejection, cancellation, panic, preparation failure and actual
process exits before/after COMMIT pass. Existing shared count/byte caps and retry
tests remain. Ledger race tests passed three repetitions; full ledger, learner
and service race suites and vet also passed. Process exits are not hardware
power-loss tests. New APIs are not installed in a production consumer.

## Experiment

[Frozen protocol](mmm-prepared-v40-protocol.md),
[raw artifact](mmm-prepared-v40.jsonl). Twelve rotated cells, three trials at
each of sizes 50 and 200, 32 admission/terminal pairs per cell. Each cell uses a
fresh WAL ledger with synchronous=FULL verified. Admission payloads are fixed
1024-byte valid JSON storage fixtures, not learned predictions. Timing includes
each full commit, excludes subsequent full byte/identity/sequence readback, and
excludes fixture preparation and open/close equally for both paths.

| Trial | Size | Control mean pair ms | Prepared mean pair ms | Reduction | Control pair p95 ms | Prepared pair p95 ms |
| --- | --- | --- | --- | --- | --- | --- |
| 0 | 50 | 1.9855 | 0.9436 | 52.48% | 2.8828 | 1.0788 |
| 1 | 50 | 1.6250 | 0.9631 | 40.73% | 1.7260 | 1.3927 |
| 2 | 50 | 1.6555 | 0.9673 | 41.57% | 1.9464 | 1.1913 |
| 0 | 200 | 5.8412 | 3.1174 | 46.63% | 8.9745 | 6.2311 |
| 1 | 200 | 5.8561 | 3.3853 | 42.19% | 9.1035 | 6.8385 |
| 2 | 200 | 5.9403 | 3.2045 | 46.05% | 9.2958 | 6.8817 |

All three size-200 trials exceed the predeclared 10% mean-pair improvement
screen. Pair p95 is nearest rank over 32 within-cell pairs, not a confidence
bound. No trial or warmup was excluded. This finite result does not establish
population tail behavior or predict the loaded end-to-end percentage gain.

The full run passed in 1.58s and verified 96,000 stored records. Independent
parsing checked all twelve unique trial/size/mode cells, embedded source hashes,
32 timings per operation per cell, verified-record counts and the paired screen.
Artifact SHA-256:
`5e2c625b707b737e6ca36cf5f78a4437e079ceb8d9a5b2e1983ebf1cc982898c`.

## Next step

Wire the prepared ledger method through an opt-in durable-wrapper constructor,
preserving the existing constructor and query-per-record control. Verify warm
originals, reopen/retry and uncertain-commit handling before a same-run loaded
comparison against v39's post-guard readback/discard path. Keep the 250ms
completion-age, serving, queue/pending and authority contracts unchanged.
No default switch, production configuration, paper publication or push occurred.
All seven direction-level requirements remain open.
