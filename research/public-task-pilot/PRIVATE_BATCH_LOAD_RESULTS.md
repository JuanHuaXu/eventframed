# Bounded collector load rescue: FAILED

The collector is implemented with a maximum of four requests per transaction,
10ms collection window in the load test, and 64 outstanding calls including
active work. Inputs are cloned after bounded admission. Expired queued requests
are rejected before preparation. Participating cancellations and earliest
deadlines govern the atomic batch; Submit waits for the real callback outcome.
Close rejects new submissions, drains admitted work and joins the worker.

Ten race repetitions of collector tests passed (1.340s after the final slot
release ordering correction). Tests cover full grouping, capacity, queued expiry,
late cancellation during a successful commit and shutdown. The original suite
also ran ten race repetitions before that correction (1.376s).

Two full sustained-load arms completed in 87.996s under
PRIVATE_BATCH_LOAD_PROTOCOL.md. Both failed the frozen gates. Raw artifacts:
private-batch-load-results.json and the two per-arm sidecars.

| Outcome | Repeat 0 | Repeat 1 |
|---|---:|---:|
| Acknowledged / recovered writes | 109 / 109 | 1236 / 1236 |
| Failed writes found durable | 0 | 0 |
| Read errors | 7952 | 5474 |
| Write errors | 3987 | 2860 |
| Late writes | 1 | 68 |
| Reopen audit errors | 0 | 0 |

Repeat 0 has four deadline failures followed by 3983 recovery-required write
rejections. Repeat 1 includes 72 cancellation errors, 40 deadline errors, four
members of one commit-canceled batch, then 2744 recovery-required rejections.
That batch waited about 60ms and prepared for 39.7ms before its commit failed.
Counts are per request, not independent commit events. Batch IDs identify shared
timings; they must not be summed across members.

All 1345 acknowledgements recovered, but neither arm sustained useful service
for the full stream. Large variability (109 versus 1236 acknowledgements) is
not a robust rescue. This candidate combines norm reuse and batching, so it does
not isolate either contribution. Aggregate latency is dominated by cheap
rejections after quarantine and must not be promoted as successful p99 serving.

## Decision

Retain the implementation and failed evidence as research-only. Do not weaken
quarantine, raise deadlines or call this production ready. Time-box the current
performance branch and return to the unresolved learning/real-task directions.
Possible future performance work includes bounded speculative preparation or
vector-kernel acceleration, but neither is validated or silently adopted here.
All seven whole research goals remain open.
