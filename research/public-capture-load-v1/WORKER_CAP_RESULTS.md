# Two-worker cap: rejected

The unchanged full-frontier public capture fixture was run on the repaired
control with explicit GOMAXPROCS=10 and GOMAXPROCS=2, in both arm orders. All
eight runs passed functional/availability/durable-ledger assertions. Every cap
run failed the original absolute recall timing screen AND the prospectively
declared paired rescue screen. Do not deploy this setting as a latency rescue.

| Frontier | Pair | Default p99 ms | Cap p99 ms | Ratio | Rescue <=0.90 |
| ---: | ---: | ---: | ---: | ---: | --- |
| 50 | 0 | 113.841208 | 277.756375 | 2.439858 | FAIL |
| 50 | 1 | 121.843583 | 263.460416 | 2.162284 | FAIL |
| 200 | 0 | 158.314500 | 359.355167 | 2.269882 | FAIL |
| 200 | 1 | 157.869417 | 362.902083 | 2.298748 | FAIL |

Cap publication age p99 86.25-93.71 ms remained below 250 ms but was substantially
worse than the matched default's 34.81-36.81 ms. Cap capture p99 90.09-96.68 ms
was also worse than default 39.66-46.48 ms. Init time rose from 15.26-15.48 s to
31.47-32.86 s. All arms completed 256 writes with exactly 46,766 raw capture
text bytes; actual capture/Recall overlaps were 203-205 at frontier 50 and
212-213 at frontier 200. The slowdown is not a reduced-data passing screen.

The diagnosis survives, but this remedy does not: capture index construction
blocks readers, and reducing worker parallelism makes that construction slower
in this workload. The results do not prove fewer workers are always bad, nor
that more cores alone remove corpus-dependent reconstruction.

Next viable structural lead: bound affected-index construction itself, using
transaction-compatible incremental/delta publication or immutable smaller
partitions with merged retrieval. The pinned library has a four-shard option
and a DeltaIndex path; neither is validated here as a rescue. Any proposal must
preserve full candidate scoring, availability filtering, atomic event/runtime
state publication, reopen correctness, invalidation and workload scale. A Flat
index mechanism control could isolate rebuild cost but is NOT a substitute for
ANN serving on million-to-billion-event corpora.

This is a consumed-design host/workload pilot, not independent outcome evidence.
The earlier mask-candidate and default absolute failures remain unchanged.
Production is untouched; all seven whole goals remain OPEN. Verify frozen
hashes, raw traces and ratios with `node verify-worker-cap.mjs`.
