# Source-owner cost v48

Frozen before execution. This is an isolated cold durable-lifecycle diagnostic,
not a loaded service benchmark or feedback/learning experiment. No production
changes, private data, outcomes, model fitting or background ingestion.

Twelve fresh-ledger cells: sizes 50 and 200, three trials, prepared raw durable
control and source owner v47, rotating their order by trial. Each cell performs
32 cycles with distinct service journal/event keys. A cycle is actual admission,
exact admission retry, original readback, then typed unlabeled discard. Every
original and retry must agree field-for-field; IDs remain sequential. All pending
records are discarded between cycles without advancing the evidence clock.

Measure monotonic durations for each phase and their sum. Control uses existing
snapshot-read batch admission, Admissions and snapshot-read batch discard. Source
owner uses its batch admission, then per-source Lookup and Discard, the APIs it
actually exposes. This intentionally measures the current lifecycle integration
cost, not an attribution of all overhead to the unique index. Document call and
transaction granularity; no hidden raw Durable escape hatch for the source arm.

Keep request generation and equality checks outside timed phases. Use fixed
nine-bit features, baseline 0.6, epoch 1, seed 42, one frozen public timestamp and
shape-valid service snapshot. Include returned-record allocations and ordinary
API validation in timings. Record all 32 samples per phase, full source snapshots
and SHA-256 hashes, runtime metadata, final database size and restart duration in
exclusive-create JSONL. After close/reopen, full canonical replay must succeed
with zero labels/pending/queue/failures; first/last originals must remain exact.

Correctness gates: zero mismatches, duplicate originals, invented labels or lost
terminals, complete 12 cells and 384 cycles. No excluded trial or warmup removal.
The existing 250ms completion-age target is NOT relaxed. A cycle p95 above 250ms
rules out that budget even before queue/guard/retrieval work; a value below it
does not establish loaded completion or serving non-harm. Report all phase
overheads and restart/disk costs, even when correctness passes. Go test success
means the accounting/correctness checks passed, not production readiness.

Use results to choose the next specific optimization or authority-integration
step. Do not tune a learning policy from this cold fixture, change defaults, or
claim the earlier 94.7% learning result has been revalidated.
