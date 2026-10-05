# Fixed-arrival admission screen: partial success, overload failure

## Results

All32 arms completed and the verifier checked source hashes and1024 read plus512
write records. The reusable researchadmission component was exercised around
actual service calls in the runner, not installed in the daemon API.

| Scheduled arrival rates | Admission | Read errors /256 | Write errors /128 | Stale rejections | Maximum read ms | Maximum write ms |
| --- | --- | --- | --- | --- | --- | --- |
| 50 reads/s +25 writes/s | off | 0 | 0 | 32 | 58.389 | 2.318 |
| 50 reads/s +25 writes/s | on | 0 | 0 | 0 | 36.766 | 32.364 |
| 200 reads/s +100 writes/s | off | 68 | 0 | 88 | 252.864 | 46.267 |
| 200 reads/s +100 writes/s | on | 112 | 83 | 0 | 130.132 | 109.308 |

Counts pool control/experimental packing, future/visible writes, and two
repetitions at each rate. They are workload counts, not independent statistical
trials. At the lower rate, all eight admitted arms pass the predeclared strict
no-error/no-stale/100ms screen. All eight admitted high-rate arms fail. At the
lower rate unleased service calls also succeed: the32 stale rejections are
internal retries, not32 failed requests.

For experimental packing with visible writes, lower-rate read p95 drops from
56.761/55.802ms without admission to36.019/35.241ms with admission. Admission also
unnecessarily delays compatible future writes: their experimental p95 rises
from27.651/27.807ms to34.467/34.535ms. Writer waiting is included, not excluded
from the latency result.

At the higher rate all195 admitted errors are deadline exceeded. Without
admission,67 reads hit deadlines and one exhausted stale-snapshot retries.
Across the artifact114 operations never entered their callback; these are
counted errors, not silently dropped measurements. Successful reads have matching
journals and zero observed future evidence. Context deadlines are cooperative,
so return times can exceed100ms; those overruns explicitly fail the screen.

## Interpretation

The coarse lease removes optimistic snapshot conflicts but does not remove
computation or queueing. Exclusive writes can block new readers while earlier
readers drain; at this higher fixed arrival rate the admission queue becomes
harmful, including to writes that did not fail without admission. Zero stale
rejections therefore cannot be used alone as the success criterion.

The earlier closed-loop success remains valid for its workload but is not an
open-loop throughput guarantee. This new result rules out promoting the coarse
gate unchanged as a general scheduling rescue. It does not falsify the retrieval
quality experiments or imply the full daemon has this capacity limit.

## Audit and scope

TASK_LEXICAL_OPENLOOP_PROTOCOL.md was written before dispatch. Each operation
has an independent scheduled arrival, with duration/deadline measured from that
arrival. Stored DispatchNS includes scheduler lag; WaitNS includes that lag and
admission delay. NS excludes post-return verification, while wall completion
includes it. The artifact retains unsuccessful operations and source hashes.

```sh
node research/public-task-pilot/check-task-lexical-openloop.mjs
```

This is a short32-read burst with at most48 tasks per arm, memory storage,
hash embeddings and repeated public facts. No durable backend, unbounded queue,
production traffic, actual agent response or general fairness claim is tested.
The arrival phase is fixed, not randomized. The verifier's integrity PASS means
records match declared checks; it does not override the failed performance arms.

Next leads: distinguish snapshot-compatible ingestion from invalidating
mutations under an atomic admission protocol; evaluate bounded read/write
batching with measured writer-age limits; shorten the precommit calculation
window without publishing a stale packet. None may weaken freshness checks,
silently discard work, or count load shedding as successful service completion.
Durable-backend testing remains required. All seven full goals remain open.
