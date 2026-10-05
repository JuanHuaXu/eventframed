# Prebuilt generation/index pair with reader retirement

Added internal/researchindex/serving.go. The wrapper is the sole compaction owner
of its durable coordinator. It builds the replacement HNSW from a captured base
before publication. Under one serving lock it publishes the rebased generation
and replaces the matching index handle. Acquire cannot observe the intermediate
pair. Newer writes remain in the delta through the existing rebase operation.

Read leases retain both a historical view and its matching index handle. Retired
graphs close only after their final reader releases. A declared retirement cap
rejects further compactions when reached. Close refuses active readers, writes,
or builds; the caller must drain them and retry. Historical leases may still read
their old state after a later unknown write, but fresh acquisition is quarantined.

## Tests and bug hunt

Real HNSW tests cover old/new answers across replacement, concurrent writes
during a paused build, rejected publication after uncertain persistence, bounded
retirement, release-after-close behavior and orderly shutdown. Persistence in
these serving tests is a controlled callback; the preceding durable coordinator
tests separately use real transactions. This is not yet one combined real-store
serving performance experiment.

Audit found and reproduced a retirement accounting bug: the final reader removed
an index from the retirement count before its potentially slow Close completed.
The regression stalls Close and verifies a new build is rejected. It initially
failed with "closing graph escaped retirement bound"; the patch retains the
charge until successful close. Reader-free replacement follows the same rule.
An error retains the charge rather than assuming reclamation. Repeated Close
also returns the recorded error rather than falsely reporting cleanup success.
No low-level close-failure injection is claimed; the regression stalls cleanup.
After the final changes, the full researchindex suite passes three race-enabled
repetitions: `go test -race ./internal/researchindex -count=3 -timeout=120s`.

## Remaining limits

This is not daemon integration or a validated throughput rescue. The serving
mutex can wait on a durable operation, and current graph construction is a slow
sequential derived-database build. No fixed-arrival service benchmark, broad
semantic recall test, idempotency integration, abrupt-crash test, or compaction
scheduling policy has run for this wrapper. Graph-retirement count is bounded;
caller-retained lease count and historical delta memory are not independently
capped yet. Unreleased readers deliberately cause backpressure, not eviction.

Next measure a combined real persistence + serving workload and implement bounded
lease admission before interpreting memory/latency claims. Keep the original
service and research controls intact. All seven whole research goals remain open.
