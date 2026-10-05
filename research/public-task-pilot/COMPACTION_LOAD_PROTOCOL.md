# Combined compaction-inclusive load screen

Research-only storage/ANN component test, not EventFrame's scored service. Normal
synchronous libravdb v1.6.13, immutable HNSW bases, flat authoritative collection,
32-dimensional SHA256-derived vectors, public-fact metadata. No private data or
LLM calls. FixedCPU4. Initial corpus200 or800, read gaps20ms or5ms, two repeats.

Per arm:256 independently scheduled reads and128 independently scheduled writes,
with100ms deadline measured from scheduled arrival. Writes occur every second
read interval; each inserts one unique ID. No retries or dropped observations.
Delta cap64, lease cap8, retired base cap2. A single background worker polls every
5ms and starts compaction at32 delta entries. Its build is synchronous within
that worker, so only one is active. No accumulation hidden outside capacity.

Read operation: acquire a coherent serving lease, search top10 for a known seed
vector, release lease. Record top10 self-hit and leased revision, errors and
scheduled-arrival latency. This is not semantic recall. Writes include real
records+revision transactions. Record capacity rejection and unknown-outcome
errors separately by preserving error strings.

Record all background builds, errors and durations, and wait for the worker to
settle before closing. Report operation latency and total drain wall time; do not
drop slow compaction from the evidence. Reopen the authoritative store, compare
all128 attempted IDs with acknowledgements, and verify final durable revision
equals200/800 initial state revision1 plus successful writes. Present failed
writes or snapshot disagreement are audit failures, not silently repaired.

Passing requires zero operation errors, zero >100ms responses, all self hits,
zero compaction errors, and successful acknowledgement/reopen checks. Failure
of any arm rejects full rescue for this tested component. Even a pass does not
validate EventFrame's journal, forecast,5w1h payload or idempotency integration.
No concurrent agent-started tests/builds during measurement.
