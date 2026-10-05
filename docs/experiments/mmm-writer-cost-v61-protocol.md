# Bounded native writer concurrency v61

Frozen before measurement. Pinned LibraVDBv1.6.13 has synchronous WAL durability
by default and signals its flusher for admitted writes. The10ms fallback flush
interval and adaptive1ms coalescing window are internal, not public options.
Streaming FlushInterval is unrelated. Async-index group-commit targeting is a
different mode and is not enabled. Do not edit module cache or disable sync.

Test the supported per-collection write-execution limit: Open retains2, explicit
OpenResearchFourWriters uses4. Queue depth64, synchronous durability, no async
indexing, snapshot locks and identity/receipt checks remain unchanged. This
tests admission concurrency, not direct control of WAL timing or global slots.

Run paired storage-shaped public journal workloads:1 or4 submitting goroutines,
128 distinct journals each with50 placeholder decision records,3 trials,rotated
2/4-slot arms:12 cells. Payloads are not scored forecasts. Capture per-call and
whole-workload monotonic time, payload bytes, complete record digest and source
hashes in exclusive JSONL. Reopen every cell via default Open and verify every
record; retain exact-retry and conflict checks. Test16 acknowledged concurrent
journals surviving immediate process exit without Close. Process exit is not
hardware power-loss simulation; source durability contract remains load-bearing.

Run native-store race tests and vet before measurement. No competing test run,
production, private data, daemon configuration, plugin or dependency changes.

Isolated screen: at4 submitters, whole-workload elapsed time must improve at
least10% in every trial; at1 submitter,p99 may not regress more than10% in any
trial. Integrity required in all cells. Retain all failures; no threshold tuning.
This cannot rescue serving writer tails by itself. Full scheduled mixed-service
tests, resource limits and active learning remain open before any adoption.
