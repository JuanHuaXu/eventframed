# Durable phase diagnostic v30

Status: v29 performance failure replicated; bounded group transactions warranted
for investigation, not yet implemented or validated by this experiment.

Repeated the complete nine-arm workload with monotonic API phase timers. Small
durable read-only/write accounting tests passed three race repetitions first.
The durable arm admitted 78/192 observations in every trial, dropping 114 each;
all 11,700 admitted candidates had real admission and explicit discard writes.
No execution errors, stale rejection, deadlines, labels or fitting occurred.

| Trial | Entry sum (ms) | Callback sum (ms) | Admit API sum (ms) | Integrity readback sum (ms) | Discard API sum (ms) | Age p95 (ms) |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 109.276 | 966.316 | 398.159 | 76.181 | 441.860 | 844.548 |
| 1 | 112.865 | 929.853 | 392.966 | 66.528 | 415.554 | 817.444 |
| 2 | 113.619 | 950.763 | 388.894 | 67.216 | 443.898 | 833.596 |

Admission plus discard account for about 87% of callback time; readback accounts
for roughly 7-8%. The APIs include validation, retry lookup and COMMIT, not just
fsync; this is not a direct measurement of disk latency. The batch still performs
two individually committed writes per candidate. Removing integrity readback
alone is unlikely to remove the backlog and is not the chosen remedy.

All ages fail 250ms. Non-durable group4 again admitted all 576 observations.
Thus the scheduling rescue cannot be generalized to actual durable admission.
No durable-learning or real-agent quality claim follows from either result.

## Next transaction design

Investigate a bounded atomic append operation on the research ledger, without
changing existing single-record semantics, WAL or FULL synchronous durability.
Bound both entry count and aggregate bytes before beginning a transaction.
Preserve input order, exact byte-level retries and terminal-after-admission
requirements, including earlier admissions within the same transaction. Reject
any conflict without partially committing earlier entries.

The durable wrapper must validate and preserve original forecasts, acknowledge
only after commit, and stop on uncertain commit outcomes until reopen/replay.
Batch admission and typed discard need separate contracts; no absence is a
negative label. Unsupported log implementations must fail before mutating worker
state. Tests must cover cancellation, conflicts late in a batch, rollback,
idempotent retry, after-commit uncertainty and process restart. Do not claim a
SQLite batch makes the SQLite/LibraVDB pair one atomic transaction.

Only after those boundaries pass should the same grouped load comparison measure
individual transactions versus batch transactions. Keep every integrity check
and retain the prior failures; no cap/deadline tuning on these outcomes.

## Artifact

`mmm-durable-phases-v30.jsonl`, SHA-256
`434e219910d26e8223b0bc78206680502ace4eefc6d1264165700841f365fabe`.
Independent verification checked all nine distinct cells, embedded hashes,
request/write counts, outcome conservation, exact durable operation counts and
durable phase sums contained within callback duration. No deployment or push.
