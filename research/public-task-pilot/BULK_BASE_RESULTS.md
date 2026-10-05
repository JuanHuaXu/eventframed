# Bulk derived-base construction passes the finite component screen

The research overlay changes only the disposable base builder: one synchronous
transaction constructs the whole captured HNSW base instead of individual
durable inserts. Authoritative writes remain synchronous, and the delta cap64,
trigger32, lease cap8, retirement cap2 and fixed-arrival workload are unchanged.
The default source builder and production daemon are not modified/promoted.

Before timing, the full researchindex suite passed three race-enabled repetitions
under bulk-base-overlay-v1. The subsequent isolated run completed all8 arms.
Raw aggregate: generation-bulk-results.json, with matching per-arm sidecars,
source hashes and retained stores. The existing verifier passes all samples.

## Results

All2048 reads and1024 writes succeed within100ms. Every read finds the expected
seed ID in top10. All1024 acknowledged records and final revisions survive orderly
reopen; no failed-present records, build errors or audit errors occur.

| Corpus | Read gap | Max read ms, repeats0 /1 | Max write ms, repeats0 /1 |
| --- | --- | --- | --- |
|200|20ms|8.43 /27.35|8.21 /8.69|
|200|5ms|11.09 /18.88|18.94 /10.88|
|800|20ms|23.93 /24.45|7.50 /7.78|
|800|5ms|29.17 /10.06|17.33 /9.45|

All32 background builds complete in42.29-117.11ms. Background builds may exceed
100ms individually; the declared budget applies to arrival-to-response reads and
writes, all of which meet it. Total worker drain is recorded rather than hidden.
The high-rate write stream has320ms delta headroom after the trigger, so these
observed builds finish before the capacity failures seen previously.

Previous sequential construction rejected401 writes and built in1.34-6.18s.
This is a descriptive cross-run rescue, not a paired speedup estimate: the new
run also accepts more records, changing later graph sizes. No gates were relaxed.

## What this does and does not establish

The finite compaction-inclusive storage/ANN component screen now passes. This is
not whole-goal completion. Inputs remain synthetic32-dimensional vectors with
public-fact metadata,200/800 initial records and short repeated arrival streams.
The test does not validate semantic recall,768-dimensional costs, long-run
stability, million-record graphs, adversarial distributions or memory ceilings.
Bulk transaction staging and graph construction still have corpus-sized work.

Full EventFrame integration remains:5w1h records, authoritative idempotency,
properly-scored laws, residual dependencies and synchronous journal writes must
be tested with this architecture before reporting a daemon throughput rescue.
Next broaden the component screen and wire the relevant service contracts;
retain all failed earlier artifacts and the frozen sequential control.
All seven whole research goals remain open. Nothing was pushed or deployed.

Verifier: `node research/public-task-pilot/check-generation-load.mjs research/public-task-pilot/generation-bulk-results.json`.
