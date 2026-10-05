# Compaction-inclusive load: read latency passes, writes saturate

Use generation-load-v2-results.json and its eight matching per-arm sidecars.
The initial failed setup attempt is documented in COMPACTION_LOAD_SETUP_CORRECTION.md;
its empty aggregate is not a completed result. V2 bounded only unserved setup
transactions and preserved all measured workload settings.

## Outcome

All8 arms fail the full gate. All2048 reads return successfully, find their seed
ID in top10, and finish within100ms. Across1024 attempted writes,623 succeed and
401 are explicitly rejected with "research index delta full". No failed writer
is present after reopen; all623 acknowledgements and durable revisions agree.
No build errors or audit errors are observed. Read maxima remain below59ms and
write-response maxima below67ms, but short rejection latency is not throughput.

| Corpus | Read gap | Write errors, repeat0 / repeat1 |
| --- | --- | --- |
|200|20ms|8 /9|
|200|5ms|64 /64|
|800|20ms|64 /64|
|800|5ms|64 /64|

Each row has128 writes per repeat. Background builds take1.34-6.18 seconds.
All15 recorded builds and final worker-drain time are included in the artifact.
At100 writes/sec, the32 free delta slots after the trigger provide only320ms
of headroom; even the smallest observed build exceeds that. At25 writes/sec,
the1.28s headroom is still below the initial200-record build time. This arithmetic
explains why merely moving rebuilds to the background is insufficient here.

## Next rescue

The derived HNSW builder currently inserts records one-by-one into a separate
synchronously persisted database. That introduces repeated durable operations
while constructing a disposable prepublication index. Test bulk construction
using existing backend APIs, preserving atomic serving-pair publication and
authoritative durability. Measure build cost and sustained load again; do not
just enlarge the delta until this short experiment ends before saturation.

This remains a storage/ANN component fixture: synthetic vectors and public-fact
metadata, no full EventFrame journals, forecasting,5w1h or idempotency contract.
Passing reads does not close goal6 or establish the daemon's intelligence claims.
All seven whole goals remain open. Production and dependencies are unchanged.

Verification: `node research/public-task-pilot/check-generation-load.mjs` checks
source hashes, sidecars, all3072 operation samples and per-arm gates.
