# Lazy repair matrix results

The existing deletion routine builds a full pairwise distance matrix before
checking whether neighbors have enough connections. The separate overlay defers
that matrix until the first neighbor below the unchanged repair threshold. Once
needed, the original full matrix and subsequent selection code run unchanged.
This avoids unused computation; it does not truncate repair or loosen thresholds.

## Correctness and work

The32-operation serial probe passed in6.478s and matched every captured node and
global state against the serial control. Published-pair calls fell from21131 to96
across its16 deletions. Cases that actually need repair still compute distances.
This is a work-count observation, not a general worst-case bound.

Three race runs matching `TestDelete|TestReclamationConcurrentSearchDeleteReinsert`
passed in1.677s. These provide adjacent deletion/concurrent reclamation coverage,
not proof for every concurrent mutation schedule. The matrix now observes vectors
later; the existing mutable concurrent algorithm has no exact snapshot-equivalence
proof. Controlled state equality here is for the serial fixture.

## Frozen cost screen

Overall FAIL, with3/4 comparisons passing. Initial and final hashes match in all
pairs. Fresh order: control0,candidate0,candidate1,control1, CPU4. The same32 inserts
and8 forced-entry deletions are timed per corpus, with counters disabled.

| Repeat | N | Seed ratio | Insert ratio | Delete ratio | Screen |
| --- | ---: | ---: | ---: | ---: | --- |
| 0 | 800 | 1.0584 | 0.9757 | 0.1102 | fail |
| 0 | 6400 | 1.0140 | 0.9775 | 0.2487 | pass |
| 1 | 800 | 0.9841 | 0.9862 | 0.1332 | pass |
| 1 | 6400 | 0.9894 | 0.9902 | 0.2732 | pass |

Ratios are candidate/control. Deletion totals improved73-89% across all four
comparisons. N6400 eight-deletion totals fell from1.140/1.166ms to0.284/0.319ms.
One initialization observation regressed5.84%, exceeding the predeclared5% ceiling.
Do not drop that phase or relax the criterion after seeing the result. Separate
controls are needed to characterize timing variance and any reproducible overhead.

## Decision

Retain lazy matrix preparation as a strong research candidate; it is not yet a
passed promotion screen or sustained serving rescue. Unlike the entry summary,
it removes most measured pair-distance work in these fixtures without requiring
new index metadata. Broaden controlled timing and concurrent/lifecycle testing
before integrating it into private preparation. This does not itself provide
private graph writes, durability, byte budgets, or steady-state throughput.

Artifacts: `lazy-repair-results.json`, `lazy-repair-comparison.json`, the four
`lazy-repair-cost/*.json` arms, and `check-lazy-repair-cost.mjs`. Normal backend and
daemon code are unchanged. All seven whole research goals remain open.
