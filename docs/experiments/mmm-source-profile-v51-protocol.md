# Source-owner profile v51

Diagnostic follow-up to the failed v50 load, frozen before execution. Repeat the
same nine rotated cells without changing code, workload, source/control selection
or thresholds. Write a fresh exclusive JSONL and collect Go CPU and allocation
profiles plus the test executable in a temporary directory. Preserve the v50
source snapshots and compare their hash maps exactly with the new artifact.

Run without competing test processes from this task. Profiling overhead and
ordinary machine variability remain possible; this is not an uninstrumented
confirmation or a new success claim. Report complete counters and latency screens,
then inspect source lookup/decoding and SQL call stacks. Use evidence to choose
whether transactional prepared source reads are warranted; do not assume an index
scan, parsing bottleneck or learning effect from aggregate wall time alone.
