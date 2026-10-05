# Bounded immutable runs: merge prerequisite

Status: research kernel, not a serving or throughput rescue.

The atomic pair cache did not reliably speed construction. Another direction is
to build a small immutable run from a drained delta, then merge runs less often,
instead of rebuilding the whole shard each time. This trades build work against
read fanout, retained storage and later compaction debt. Those costs must remain
visible at the original load and global delta limits.

Source: Sarkar et al., *Constructing and Analyzing the LSM Compaction Design
Space*, PVLDB14(11),2021,
[primary paper](https://www.vldb.org/pvldb/vol14/p2216-sarkar.pdf).
The paper separates trigger, layout, granularity and movement policy, and
describes retaining latest valid versions through compaction. Its results concern
LSM key-value engines, not ANN graphs. Our graph adaptation inherits no ANN
recall, construction, or latency guarantee from that source.

## Correctness boundary

Simply querying top k from every overlapping run is incorrect: tombstones and
demoted updates in newer runs can shadow the entire older top k. An old result
below that prefix can then be the actual best remaining result.

`RunMergePlan` captures complete run manifests newest first, including tombstones.
Construction copies manifests, establishes latest-owner membership, and counts
each run's live records L and shadowed live records S. It requests min(L,k+S)
candidates. An exact sorted prefix of this length contains enough unshadowed
results for global top k. ANN prefixes retain their approximation; the kernel
cannot certify unseen candidates. A run manifest must match its immutable graph.

The kernel limits run count to16, k to200, and each prefix to328. A plan exceeding
the caller's prefix cap fails with ErrCapacity, requiring compaction or an
explicitly different design, not silent truncation. Fully shadowed runs currently
also request their bounded full live prefix; skipping them is a future optimization.
At query time, no complete manifest scan is needed. With C total requested
candidates, merge work is expected O(C) hash validation plus O(C log C) sorting.
Manifest preparation uses O(total manifest entries) time/storage off-path. This
cost and retained vector/index storage cannot be called constant with corpus size.

## Tests

`TestRunMergeHiddenPrefixAndDemotion` covers tombstones, downward score changes,
insufficient-prefix rejection, capacity rejection and caller-input isolation.
`TestRunMergeRandomExactOracle` compares300 generated histories (up to8 runs,
50 IDs, ties, updates, tombstones) with exhaustive latest-version ranking.
`TestRunMergeRejectsInvalidInputs` rejects invalid manifests and nonfinite,
duplicate, unknown, deleted, unsorted or missing candidates.

Full ordinary suite: `go test -race ./internal/researchindex -count=3` passed
in17.200s, including all three new tests. These tests use exact synthetic score
lists, not live ANN runs or an integrated service performance workload.

## Next gates

Build real small-run graphs, bind graph handles and manifests atomically, keep
old readers alive, carry writes arriving during a build, and prove durable
revision/recovery behavior. Bound total runs and retained resources; compaction
must catch up during sustained growth, not merely delay saturation. Repeat
the same load with updates/deletes as well as appends, plus exact-oracle recall
checks. No production integration, speed claim or whole-goal completion yet.
