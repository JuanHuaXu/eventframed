# Prepared Immutable Base And Bounded Overlay (DESIGN)

Freeze before measurements. All seven original goals remain OPEN.

## Reasoning Gate

Confirmed: V3 HNSW lacks DeltaIndex, and each explicit transaction materializes
the historical collection then reconstructs its graph. Four shards reduce the
constant, not corpus dependence. Hypotheses: graph construction, record-state
materialization, and durable WAL cost. This prototype isolates the first two;
it does not claim that they exclusively explain earlier latency.

Upstream master was inspected at 18515d4ce0620b24fd2512b0f9023b8b1f4d8355.
PR6 implements async indexing but documents eventual ANN visibility. PR9 adds
SQL/transaction support. No listed issue or PR supplies a transactional HNSW
DeltaIndex rescue. Do not substitute eventual visibility for committed search.

## Candidate And Invariants

Only fresh isolated dependency copies, inherited repaired V3 source. New
research-only HNSW wrapper: immutable graph plus at most64 changed-ID slots;
nil slots are tombstones. Prepare copies the bounded overlay and caller vector
payloads. Commit only publishes prepared state; Abort cannot expose it. Base
nomination excludes shadowed ordinals, then all eligible overlay vectors are
scored and merged BEFORE final top-k. No silent cap reduction or dropped put.
Overflow stages a complete synchronous graph compaction before WAL publication.
This is NOT asynchronous compaction or a bounded corpus-memory guarantee.

Readers and prepare/commit are serialized by the wrapper lock. Retired base
graphs close before the next preparation or on Close, never through a reader.
Canonical records remain durable storage's responsibility. The prototype's
canonical-vector generation additionally retains O(ND) private RAM; record
metadata is not duplicated into the index. Unsupported legacy snapshot formats
must fail closed rather than be interpreted as a new generation.

Falsifier: any precommit visibility, abort mutation, deleted/upserted shadow
leak, ordinal collision, filtered overlay bypass, failed WAL mutation, missing
reopen record, race, future-ID exposure, reduced nomination count, or different
complete learned law where inputs/accepted state match rejects that capability.

## Preflight And Cost

Before timing: standalone prepare/abort/commit, upsert/delete, copied-vector,
dimension/finite/ordinal validation, filter and threshold, overlay overflow,
persistence/reopen, canceled preparation and concurrent readers. Ordinary and
race, plus existing relevant transaction/recovery regressions in both arms.
Add a storage commit failpoint to prove failed WAL cannot publish the prepared
index, and a cross-collection atomicity check. Preserve failures and repair
records; no thresholds may be retuned after collection.

Frozen component timing:32D and128D;256 and1024 initial records;128 individually
prepared commits per cell; top50 queries after each commit; three repetitions,
alternating control/candidate execution order; GOMAXPROCS10. Control retains
ordinary full graph rebuild semantics. Include initial graph construction,
all compaction commits, peak retained payload and total allocations separately.
Report per-command wall and raw per-operation timings. Search numerical quality
uses independently computed FP64 cosine; tie-aware recall>=.95, mean regret
<=.005. These are finite synthetic geometry screens, not agent usefulness.
No automatic speedup verdict: ratios are descriptive until full service load.

Then retain a public EventFrame service integration preflight using identical
inputs and full frontier/packet/forecast evidence if component gates permit.
Do NOT call cold-law equivalence evidence about learned posterior/residual law.
Background freshness, sustained open-loop queues, kill/recovery, large corpus,
RSS, untouched agent outcomes and all other original scientific gates remain.

## Sources

FreshDiskANN https://arxiv.org/html/2105.09613v1 sections5.1-5.6 motivates mutable
recent data plus a long-term graph and background merge. Its scale results do
not transfer. HNSW https://arxiv.org/abs/1603.09320 supplies the graph algorithm.
Upstream visibility contract:
https://github.com/xDarkicex/libravdb/blob/18515d4ce0620b24fd2512b0f9023b8b1f4d8355/docs/research/async-wal-indexing-plan.md
