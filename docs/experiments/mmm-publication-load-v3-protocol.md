# Ingestion-only publication load v3

Controlled architectural ablation, NOT a general store integration. Repeat v2's
three rotated off/16/64 arms at64 and192 requests, with the same50-candidate
frontier, actual feedback/refits,32/96 future writes, and unchanged80% completion,
250ms learning-age p95 and10% paired serving-p99 gates.

On arms install a test-only store wrapper AFTER initial policy binding and seed
ingestion. During measurement the only data mutation is ordinary Put. Each Put
announces future-ingestion intent before delegating to the real persistent store,
then validates and publishes the committed version. Errors/duplicates quarantine
rather than inventing rollback. Journal writes do not change version/data semantics.
The bridge asks the coherent publication proof without first waiting on Snapshot.
Off serving and all ordinary serving Snapshot calls retain their original behavior.

The wrapper does not intercept general mutation methods and lives only in a test
file. Composition, posterior, graph, agency, deletion, policy and crash-recovery
paths are OUTSIDE this ablation, not implicitly proved safe. Passing would support
the latency hypothesis and motivate a full adapter audit, not production readiness.

Use exclusive artifacts and retain all arms, age samples, failures and source
text/hashes. No queue/drop/label policy changes from v2; no private or production
data and no generation-model call. These repetitive fixtures measure workload,
not independent learning evidence. Test pending-write and uncertain-outcome
boundaries separately before accepting even the narrow architectural result.
