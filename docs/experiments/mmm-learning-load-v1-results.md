# Actual learning load v1 results

**FAILED the frozen completion screen in every pair.** Serving latency and
read/write error gates passed, but the learning consumer did not keep up.

See [protocol](mmm-learning-load-v1-protocol.md) and
[raw six-arm artifact](mmm-learning-load-v1.jsonl). Three paired trials used fresh
temporary persistent stores, 64 concurrent-reader requests and32 future writes
per arm. On arms used the real temporal feedback bridge and background refits.

| Trial | Off p99 ms | On p99 ms | On/off | Frontiers completed | Labels completed | Dropped |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 27.997 | 26.038 | 0.9300 | 34/64 | 1700 | 30 |
| 1 | 31.059 | 29.873 | 0.9618 | 33/64 | 1650 | 31 |
| 2 | 29.012 | 29.127 | 1.0040 | 34/64 | 1700 | 30 |

All arms had zero reported read/write/bridge errors and overlapping writes.
Every admitted frontier's50 labels completed; no worker failures were reported.
Only51.56%-53.13% of offered frontiers completed, below the frozen80% requirement.
The bounded queue protected serving by dropping work; that is not successful
learning-throughput integration. No increase in queue capacity is adopted here.

The audit recomputed nearest-rank p99 from all64 latency samples per arm, checked
32 write samples, source hashes, and on-arm conservation:
admitted+dropped=64 and completed labels=50*admitted. The raw artifact retains
all failures. These finite timings are not population p99 guarantees.

## Separate profiling diagnostic

A second run, with CPU profiling, is retained separately as
[profile-run outcomes](mmm-learning-load-v1-profile.jsonl) and
[CPU profile](mmm-learning-load-v1.cpu). It also failed completion (33,35,36
frontiers), and is not fresh confirmation or a replacement for the original run.

Of4.37 seconds sampled CPU, the background Adapter.Feedback/observation.Fit path
accounted for about.16 seconds cumulative (3.66%). Scheduler/runtime and store
work dominated the profile. This does NOT prove fitting is cheap in wall time or
identify the blocking critical path: CPU profiles omit much waiting. The temporal
store proof takes a read lock, and the consumer performs proof checks for each
candidate label; contention is a candidate explanation, not an established cause.

Next instrument admission, per-frontier feedback, and worker-drain waits before
choosing a batching or scheduling rescue. Do not discard labels, skip validation,
or change the finite success threshold to turn this failure into a pass.

## Scope

These are repeated deterministic fixture outcomes for workload measurement,
not independent real-world evidence, calibration, or learning-quality validation.
No OpenClaw/provider generation was used. Models and replay tracking remain
memory-resident even though the event store is persistent. Publication, crash
recovery and real-task representation remain open. Nothing was deployed/pushed.
