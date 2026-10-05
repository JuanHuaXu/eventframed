# Already-pending batching does not yet rescue durable service

## Scope and controls

Added an unwired research queue and a generated copy of the atomic durable
harness. CaptureTurn still performs validation,5w1h extraction and embedding.
Only timed Put enters batching. Capacity64, maximum batch16, no accumulation
timer; shared transactions settle synchronously. Callers wait for settlement,
and late success remains a failed latency gate. Any callback error conservatively
halts subsequent commits until reconciliation. This is not production code.

Queue tests passed ten race-enabled repetitions. The combined queue, admission,
and libravdbstore race suites passed once. Covers bounded pending capacity,
manual cancellation after dispatch, canceled pending work, draining/joining,
deadline propagation, malformed result cardinality and stop-after-error.
No durable fault injection or crash proof is claimed.

The first run overlapped our race tests and is diagnostic-only. See
BATCH_QUEUE_CLEAN_RUN.md. Both processes finished before the unchanged rerun.
Use batch-queue-clean-results.json for the results below. Both runs and their
retained databases remain available; neither was overwritten. Other host
activity was not controlled.

## Clean32-arm result

Each rate/admission row aggregates8 arms,256 reads and128 writes. Reads arrive
at50/s or200/s, event writes at25/s or100/s respectively. All16 low-rate arms
pass; all16 high-rate arms fail.

| Rate | Admission | Read errors | Write errors | Responses >100ms | Stale rejections |
| --- | --- | --- | --- | --- | --- |
| Low | off | 0 | 0 | 0 | 0 |
| Low | on | 0 | 0 | 0 | 0 |
| High | off | 123 | 0 | 135 | 136 |
| High | on | 0 | 91 | 8 | 0 |

The no-admission high-rate run uses batches1..4; admitted attempts use1..6.
Neither realizes the earlier best-case batch16 amortization. All warmups pass;
all901 successful journals and421 acknowledged writes survive orderly reopen.
Successful frontiers contain no observed future evidence. Verifier checks all
1024 read and512 write samples, artifact hashes and reopen counts.

Do not interpret zero admitted read errors as successful service rescue:91/128
high-rate event writes failed, reducing downstream work. The previous atomic
run had106 read errors and84 write errors with admission, but this cross-run
comparison changes settlement and stop-after-error behavior and is descriptive
only, not evidence of a general improvement.

## Failure accounting and next lead

The91 write errors comprise:

- 5 transaction errors reporting deadline expiry during index insertion.
- 8 later failures requiring reconciliation after those transaction errors.
- 37 plain deadline errors (may include queue cancellation or admission).
- 41 later failures requiring reconciliation after a plain deadline error.

The latter guard is intentionally too conservative when a transaction provably
never entered, but the five index errors are genuinely unresolved outcomes from
the caller's perspective. Do not simply remove the guard or treat every timeout
as rollback. Next reconcile failed event IDs and durable snapshots from retained
DB copies, then distinguish safe pre-commit rejection from uncertain commit.
After that, consider draining additional pending jobs after obtaining admission
and measuring journal batching separately. Retain the entire100ms arrival budget.

`node research/public-task-pilot/check-batch-queue.mjs` verifies the clean artifact.
The old diagnostic artifact can be supplied as an explicit argument. All seven
whole goals remain open. No production, module, whitepaper or publication change.
