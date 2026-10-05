# Local libravdb: low-rate pass, high-rate failure

## Result

The actual local libravdb adapter completed all32 arms. Each used a fresh,
retained database, the incremental diversity overlay, public facts and hash32
embeddings. No production service or remote database was accessed.

| Arrival rate | Admission | Read errors /256 | Write errors /128 | Maximum read ms | Maximum write ms |
| --- | --- | --- | --- | --- | --- |
| 50 reads/s +25 writes/s | off | 0 | 0 | 32.546 | 18.162 |
| 50 reads/s +25 writes/s | on | 0 | 0 | 43.104 | 37.817 |
| 200 reads/s +100 writes/s | off | 137 | 32 | 161.710 | 126.954 |
| 200 reads/s +100 writes/s | on | 110 | 79 | 111.969 | 104.402 |

Every lower-rate arm passes the predeclared screen; every high-rate arm fails.
Both warmups succeed in every arm. The admitted high-rate condition has zero
stale rejections but still189 errors. Admission therefore does not provide the
same rescue on this durable workload that it did on the memory-store workload.
The low-rate admitted maximum43.104ms is a finite observation, not a tail bound.

All777 successfully returned journals match their complete JSON digests after
close/reopen, and all401 acknowledged writes are present by ID. There are zero
recorded reopen errors. Successful frontiers contain no observed future events.
All1024 read and512 write outcomes are retained, including358 errors and98
operations canceled before callback entry. No error is omitted from denominators.

## Failure scope

The higher-rate failures include context expiry in search, journal persistence,
and event insertion/index transaction commit, as well as admission waiting.
Without admission, some reads also retry invalidated snapshots. These messages
identify affected boundaries, not a measured decomposition of storage cost.
We have not yet isolated whether durable commit time, index work, locking or
allocation dominates the additional overhead.

Orderly reopen checks only acknowledged results. Failed/canceled writes can
have uncertain outcomes; this run does not establish that they left no rows,
partial indexes or version inconsistencies. Post-cancellation reconciliation
and a fresh recall after recovery remain necessary. There is no power-loss,
crash, disk-full, torn-write or multi-process validation here.

The in-memory performance result remains valid for its stated scope. This run
contradicts extending it to this durable high-rate workload without further
changes. Lowering offered load, dropping requests, weakening freshness, or
acknowledging before the durable boundary would not rescue the failed criterion.

## Audit

DURABLE_OPENLOOP_PROTOCOL.md preceded dispatch. The exact current adapter source
is hashed, including existing worktree changes. The runner keeps service call
time separate from close/reopen verification. Durations include scheduling and
admission wait and use scheduled-arrival deadlines. These are short bursts with
fixed arrival phases, not steady-state capacity trials.

```sh
node research/public-task-pilot/check-durable-openloop.mjs
```

Raw data: durable-openloop-results.json. Retained isolated databases live in
durable-openloop-results.json.stores, contain public fixtures/research metadata,
and are not source-distribution artifacts. Reopen counts are checked against
successful outcomes by the verifier; integrity PASS does not mean latency PASS.

Next: instrument named durable phases and check ambiguous cancellation outcomes
against the retained stores before proposing a persistence rescue. Longer-load
and whole-agent learning validation remain open. None of the seven full goals
is complete. No production or whitepaper changes were made.
