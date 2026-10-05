# Snapshot/proof timing diagnostic

Separate replay of the frozen learning-load workload, instrumenting only the
research bridge's store calls. Ordinary serving uses its original store handle.
The wrapper forwards the optional temporal proof without weakening it.

**Outcome: still FAILED throughput in all three pairs** (34/64 frontiers each).
See [raw artifact](mmm-learning-proof-v1.jsonl). No runtime rescue was applied.

Aggregate milliseconds across each on-arm consumer:

| Trial | Admission | Snapshot reads | Temporal proof | Journal reads | Feedback submission |
| --- | ---: | ---: | ---: | ---: | ---: |
| 0 | 274.147 | 267.133 | 0.348 | 6.437 | 1.495 |
| 1 | 255.582 | 248.152 | 0.267 | 6.912 | 1.612 |
| 2 | 247.296 | 239.563 | 0.417 | 7.225 | 2.234 |

Snapshot/proof totals include bridge creation, admission and feedback, so they
are not strictly disjoint subintervals of admission. Feedback totals are small,
however, and snapshot reads plainly dominate measured bridge wall time. Largest
single snapshot reads were17.230,14.532,15.751ms. Proof computation and journal
retrieval are not the dominant costs in this workload.

The persistent store Snapshot method acquires writeMu.RLock. Ingestion holds
writeMu.Lock across its transaction. Journal persistence also holds a read lock;
a waiting writer can therefore delay subsequent snapshot readers. This identifies
store-lock contention, not neural/model arithmetic, as the leading admission
bottleneck. The earlier CPU profile alone could not reveal these waits.

## Rescue boundaries

Do not remove snapshot checks or reuse stale dependencies to make timing pass.
A lock-free snapshot would need coherent publication across every mutation and
must address in-progress commits; replacing a getter with an unsynchronized
struct read is not safe. That is a broader store change, not a local one-line fix.

A separate bounded-buffer experiment is a lower-risk lead for finite write
bursts, provided it reports memory, completion delay, and a longer workload so
increasing capacity is not confused with increasing steady-state throughput.
No queue-size change or new throughput claim has been adopted in this diagnostic.

No real-task accuracy, model generation, production deployment or push occurred.
The seven-direction research objective remains open.
