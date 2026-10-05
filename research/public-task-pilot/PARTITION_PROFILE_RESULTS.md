# Profile findings: construction CPU and allocation churn

Diagnostic protocol: PARTITION_PROFILE_PROTOCOL.md. Raw CPU/allocation profiles
are retained in `partition-profile-results.json.stores`; textual pprof reports,
raw-profile hashes and memory summaries are in `partition-profile-analysis`.
The capture script verifies runner hashes and sidecars before extracting reports.

These two profiled arms are NOT performance validation. They use the isolated
no-sync-derived overlay, keeping authoritative writes synchronous. No competing
test/build workloads ran during collection. Setup is excluded from CPU sampling.

| Initial records | Compaction sampled CPU | Read sampled CPU | Write sampled CPU | Allocated MiB | GC cycles | Total GC pause ms | Largest GC pause ms |
|---|---|---|---|---|---|---|---|
|3200|3.30s (48.89%)|1.49s (22.07%)|0.37s (5.48%)|1342.80|48|154.75|8.50|
|6400|5.25s (53.30%)|2.18s (22.13%)|0.64s (6.50%)|1526.77|34|57.32|10.88|

Percentages use all sampled CPU, including unlabeled runtime/library work. They
are not wall-time fractions or lock-wait estimates. At6400 total sampled CPU is
9.85s over5.34s wall time, reflecting parallel work.

## Evidence That Changes The Next Action

Within compaction at6400, DotProductNEON accounts for2.39s flat CPU; FreeList.popFree
accounts for0.88s. Neighbor selection/heuristic stacks account for substantial
additional cumulative work, but cumulative entries overlap and must not be summed.
These point to graph construction and allocation, not merely fsync.

The differential sampled allocation profile attributes roughly0.41GiB to storage
cloneEntry,0.22GiB to encoder-buffer growth in WriteUint32,0.19GiB to encoder
acquisition and0.18GiB to committed graph-WAL collection. These are allocation
sites across the interval, not live memory or proof of a leak. Allocation profiles
are sampled and GC-delayed; use the exact runtime counter delta for total bytes.

Write-labeled sampled CPU is dominated by syscalls (0.50s of0.64s at6400), but CPU
profiles omit sleeping I/O/lock time. They do not exonerate authoritative storage
from admission stalls. Largest observed stop-the-world pauses are under11ms;
the profiles do not establish GC as the cause of prior100ms deadline clusters.

The profiled6400 arm had29 capacity rejections and one read error; all483
acknowledged writes reopen. The3200 arm acknowledged/reopened512. These remain
diagnostic observations, not replacements for unprofiled failures.

## Next Investigations

1. Trace candidate construction allocations and encoding size estimates; compare
   existing upstream behavior before proposing a narrow semantics-preserving fix.
2. Consider reusing safe derived structures or reducing repeated pair-distance
   calculations, but preserve graph quality and measure fresh recall controls.
3. Use bounded phase timing or a separate blocking trace for admission outliers;
   do not attribute them to CPU/GC solely from this profile.

No backend or production code was patched here. All seven whole goals remain open.
