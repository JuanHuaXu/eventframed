# Transactional source reads v52

Frozen before execution. v50/v51 failed the loaded source-owner age screen;
profiling showed repeated SQLite shared-memory locking on point reads, not proven
SQL parsing dominance. Test one bounded read transaction with prepared repeated
queries before changing any owner default or loaded path.

Twenty-four rotated cells: three trials, sizes50/200, all-hit/all-miss requests,
point-read control and transactional source batch. Each fresh ledger contains
1,000 distinct bound 1,024-byte valid JSON storage fixtures (not learned forecasts)
written through the prepared append path with the unique source index enabled.
Each cell issues32 groups; deterministic lookup indices are shared across arms.
Input generation and comparison against expected keys/payload bytes are outside
timing; ordinary result allocation, query/validation and transaction work remain
inside. No excluded warmup, trial, failure or changed workload after results.

Record all read-group durations, result counts, returned payload bytes, source
snapshots/hashes and runtime metadata in exclusive JSONL. Correctness requires
every hit's full original key/payload and every miss's bound Source plus zero
Entry, with exact caller order. Default point reads remain the control.

Before measurement, test all five source fields and indexed SEARCH, duplicate
payload independence, missing versus error behavior, a concurrent inserted row
visible only to the next transaction, late cancellation/panic cleanup, encoded
request and aggregate payload caps, oversized/type/binding/identity corruption,
missing index and ambiguity. Run ledger/learner/service race and vet checks.

Performance screen: at size200, mean read-group time must improve by at least20%
for BOTH hit and miss cases in EACH trial. Report size50 and p95s as well. A pass
only warrants opt-in owner integration; it cannot establish loaded completion
age, serving non-harm, warm learning, current evidence authority or a whole
research direction's success. Byte bounds and transaction/identity semantics
must not be weakened for speed.
