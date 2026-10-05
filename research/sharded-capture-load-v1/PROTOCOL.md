# Four-shard transaction pilot

Frozen before execution, 2026-10-05. Goal 6 remains the full original target,
not this pilot. Control `f3231fab2244c5c6bca8f7f822e5669ac58d13cd`; candidate adds
the existing `libra.WithSharding(true)` option to new event collections only.
No dependency, existing database, daemon setting or production file is changed.
The pinned LibraVDB v1.6.13 has four fixed shards with ID-based FNV-1a routing,
queries all shards, and merges top-k results. Transactions route affected IDs
to their shard. This is a smaller-index hypothesis, not constant-cost ingestion
or a proof that a billion-event corpus meets a 100 ms envelope.

## Patch audit gate

Confirmed symptom: prior full public-capture load recall p99 115-176 ms, while
all functional publication/availability/ledger checks passed. Profile evidence
connects HNSW construction under capture-held store locks to reader delay.
Mask reuse and fewer workers did not rescue loaded latency. Those failures are
preserved, not replaced by this pilot. Competing explanations include index
construction, lock choreography, metadata hydration and journal work.

Recommendation under investigation: four smaller transaction index partitions
may reduce capture-held lock time. Falsifier: capture/recall tails fail to improve
at equal completed work, or any chronology/atomic-publication/reopen check fails.
Do not remove persistence, score fewer candidates, bypass guards or claim ANN
outputs are bit-identical. Both historical records and live envelopes need
correct durable routing; this is not repairing bad rows or migrating old data.

Related work was checked read-only: current upstream master
`18515d4ce0620b24fd2512b0f9023b8b1f4d8355` still has index-delta and rebuild
paths; PR 6 supplies HNSW/async-indexing machinery but does not establish that
this atomic event/runtime transaction avoids reconstruction. No automatic
dependency upgrade is made. Pinned source and runtime tests are authoritative
for this experiment. eventframed's published main has no WithSharding use.
The upstream issue/PR list was read via GitHub's public API because cached web
pages were stale; no exhaustive undisclosed-fix claim is made.

## Matched load

Copy the IDENTICAL frozen public-capture fixture from public-capture-load-v1
to both isolated checkouts. The raw template and formatted fixture hashes must
match that experiment. Only 288 public DESIGN capture templates and question
strings are used, not private/confirmation records or oracle answers/IDs.
The synthetic clock and labels are timing/worker mechanics, not real quality.
Use 1,000 seed replicas, 256 future captures, 128D/SQ8, RecallK 50 and 200,
PackK 10, 64 sequential calls/labels, live offered-before-guard publication age,
complete 128-row ledger and same-epoch replay. Full pre-packing frontier and
future exclusion remain mandatory. Explicit GOMAXPROCS=10 in both arms.

Two paired repetitions: control/candidate then candidate/control, at each
frontier size, fresh stores each time. Keep every raw sample/failed exit. No
parallel benchmark, threshold retuning or deleting unfavorable cells. Frozen
absolute gates remain recall p99 <100 ms, publication age p99 <250 ms. Additional
pilot rescue screen: candidate/control recall p99 <=0.90 in all four comparisons,
with all functional checks passing. Report bytes, initialization, completed
writes, actual overlap and capture quantiles. Closed-loop writer rate changes
with service cost; this is not equal offered open-loop load. A finite nearest-
rank p99 over 64 calls equals max and has no population-tail guarantee.

## Functional guard scope

Before timing, run existing full libravdbstore regression tests in both arms,
vet and non-optional persistent admission/batch/temporal race guards. Require
named suites to actually execute; no skipped opt-in audit is a pass. Preserve
any test failure even if load is subsequently favorable. Tests that pin exact
source hashes or ANN ordering are not silently excluded; investigate failures.

Add a candidate-specific topology/reopen check in libravdbstore: four occupied
shards in the new event collection; Get/Search/ListAll consistency and no
future crowding; full raw payload and snapshots survive close/reopen; duplicate
retry is idempotent and conflicting payload fails closed; delete remains
tenant-scoped and survives reopen. No automatic old-store migration. This is
not crash-injection recovery, cross-epoch learning transfer, real agent utility,
ranking-quality equivalence or sustained/million-record validation.

All seven whole goals remain open even if this pilot passes. Any viable rescue
must later cover the larger corpus, visible mutations, transfer/recovery,
contract-network paths and real outcome tasks without narrowing the target.
