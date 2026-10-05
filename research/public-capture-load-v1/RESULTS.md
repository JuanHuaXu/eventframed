# Public capture, full frontier: latency screen failed

All eight ordinary-build runs passed the functional checks: each Recall exposed
exactly 50 or 200 pre-packing candidates; future-dated captures never entered
the frontier or packet; all 64 labels were observed published before Close;
all 128 durable rows and same-epoch replay agreed. All eight runs FAILED the
frozen recall p99 <100 ms timing screen. Publication age passed its <250 ms
screen. The mask candidate also FAILED paired non-regression in one 200-event
comparison. It is not a loaded-service rescue.

| Frontier | Pair | Arm | Recall p99 ms | Publication p99 ms | Capture p99 ms |
| ---: | ---: | --- | ---: | ---: | ---: |
| 50 | 0 | Control | 127.881250 | 35.553292 | 39.991375 |
| 50 | 0 | Mask reuse | 116.424833 | 34.946542 | 40.158292 |
| 50 | 1 | Mask reuse | 115.005209 | 35.540209 | 39.864750 |
| 50 | 1 | Control | 114.563791 | 35.073416 | 39.625834 |
| 200 | 0 | Control | 157.943916 | 37.802375 | 45.962291 |
| 200 | 0 | Mask reuse | 156.878250 | 35.792167 | 45.017250 |
| 200 | 1 | Mask reuse | 176.371334 | 44.929125 | 47.266542 |
| 200 | 1 | Control | 157.645375 | 38.749500 | 46.072083 |

Candidate/control p99 ratios: frontier 50, 0.910414 and 1.003853; frontier 200,
0.993253 and 1.118785. The latter exceeds the frozen 1.10 ceiling. Keep every
negative cell; there is no post-hoc deletion or threshold adjustment.

Each run initialized 1,000 public DESIGN capture replicas in 15.11-15.37 s,
then completed 256 future CaptureTurn writes and 64 full Recall calls with one
guarded worker label per call. Actual capture/Recall interval overlaps were
204/256 at frontier 50 and 213/256 at frontier 200. Packing retained 6-10 and
10 records respectively; the tap confirms full pre-packing frontiers. Full
individual monotonic timing samples, bytes and counts are in `results.json`.

## Diagnostic evidence

The additional instrumented control run is NOT a new timing verdict. In its
whole-process CPU profile, HNSW batch insertion accumulated 67.96 sampled CPU
seconds; foreground Recall accumulated only 0.62 sampled CPU seconds. CPU
totals span multiple workers and include initialization, so they are not
exclusive wall-time shares. Recall-scoped blocking accumulated 5.37 seconds.
Mutex contention attributed to CaptureTurn/Store.Put releases accumulated
about 4.99 seconds. Idle monitor/channel totals are not serving delay.

The pinned verified LibraVDB v1.6.13 transaction path uses index deltas where
supported, otherwise builds replacement affected-collection indexes. This
fixture's HNSW construction is visible in the profile. Store.Put holds its
write/event locks across the transaction, making capture work a plausible
direct cause of foreground reader delay. Source and profile evidence refute
input-mask scans as the main remedy here; they do not apportion every delay or
prove all transaction/index modes behave the same. A separate frozen two-worker
cap screen tests a partial contention hypothesis without weakening durability.

## Interpretation and limits

The component microbenchmark improvement is real within its own fixture; it
does not establish a faster full service. The old one-event future-writer
fixture fits the absolute timing screen, while this larger public-capture
fixture does not. Both facts remain in the record.

This is a constructed timing workload, not historical replay or public-fact
accuracy. The 288 inputs are DESIGN capture templates, not independent facts;
see the explicit metadata erratum in `AUDIT.md`. Internal retry counts and peak
RSS are not captured. Closed-loop writer cadence is not an open-loop backlog
guarantee. Publication labels are synthetic mechanics probes, not real agent
outcomes, and same-epoch replay is not cross-epoch learning continuity. All
seven whole goals remain OPEN, with goal 6 finite latency contradicted here.
Production and private data are untouched.

Run `node verify.mjs` to recompute hashes, counts, overlaps, quantiles and ratios
from the retained transcripts. `profile.mjs` and its diagnostic manifest record
the distinct post-screen profiling procedure.
