# Event transaction batching: promising component, service untested

The normal embedded backend supports amortizing synchronous transaction work.
No original store/service/module file was changed by this experiment. The new
method is research-only and is not called by the daemon or CaptureTurn.

## Correctness

Three race-enabled repetitions pass payload/vector parity, duplicate handling,
final snapshot parity, per-new-event ingestion motion, orderly reopen, rejection
before writes, and mixed concurrent ordinary Put. One transaction publishes all
new members and its final snapshot; every returned result uses that snapshot.
Only new unique IDs increment version/epoch, including within-batch duplicates.
The full libravdbstore package also passes once under the race detector
(`go test -race ./internal/store/libravdbstore -count=1 -timeout=180s`).

Limitations: no injected commit failure, crash test, queue, per-member deadline
handling or service replay yet. A commit error is explicitly an unknown outcome:
the research contract requires reopen/reconciliation before reuse. Validation
can leave empty collection metadata, but tested rejected batches insert no event.

## Frozen timing screen

Apple M4, CPU4, normal synchronous libravdb v1.6.13. Three one-operation runs;
each operation inserts128 already-available public-fact events into a fresh DB.
Setup/close excluded; first collection initialization included. This is a small
engineering screen without confidence intervals or randomized arm order.

| Batch size | Median ms /128 events | Amortized ms/event | Speedup |
| --- | --- | --- | --- |
| 1 (original Put) | 926.002 | 7.234 | 1.00x |
| 2 | 476.053 | 3.719 | 1.95x |
| 4 | 234.940 | 1.835 | 3.94x |
| 8 | 130.994 | 1.023 | 7.07x |
| 16 | 68.971 | 0.539 | 13.43x |

Every batched arm clears the predeclared20% median-improvement screen. These
amortized times are not individual response latencies. The benchmark's repeated
public fact is a storage fixture, not independent semantic evidence.

## Next test and rejection criteria

Proceed to bounded queue integration, not production promotion. Batch only
already-pending compatible writes and account for the oldest member's deadline.
At100 event writes/sec, waiting for16 arrivals from an empty queue costs150ms
before committing. That would fail the100ms requirement even with fast commits.
Do not add that wait to obtain the measured best-case throughput.

Journal durability is still a separate workload: prior phase traces observed
roughly6ms in journal_commit per successful low-rate read. This is not a proven
fixed lower bound, but it is enough to warn that event batching alone may not
rescue200 reads/sec. Retain the same journal accounting and fixed-arrival screen;
do not hide failed or unentered callbacks or move acknowledgement before sync.

Raw artifact: event-batch-screen.json, with source/protocol hashes and all15
benchmark measurements. Verification: `node research/public-task-pilot/check-event-batch.mjs`.
All seven whole research goals remain open.
