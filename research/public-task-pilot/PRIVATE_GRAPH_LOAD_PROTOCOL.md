# Private graph sustained-load rescue

Freeze before execution. Two repetitions, initial N=6400, 768 dimensions,
4096 append writes at 100/sec and 8192 seed self-queries at 200/sec, CPU4,
100ms deadlines measured from scheduled arrival, no retries. At most 64 admitted
write calls, eight reader leases and two retired roots. Every successful write
atomically persists vector/ID and revision to fresh local libravdb authority.

The candidate intentionally replaces eight initial partition graphs plus tail
flush/merge with one serial captured HNSW graph and private incremental edits.
Initial connectivity comes from the N6400 initial serial-work-control capture;
vectors use the original workload's deterministic unnormalized SHA-byte values.
Cosine is scale-invariant; this does not imply identical floating-point traces.
New levels are frozen SHA-derived exponential levels (ML=1, cap31), not tuned
from outcomes. Query ef=400, k=10, evaluation cap=20000. No background repair.

Shared serving gates: zero request errors, zero responses beyond scheduled
100ms, zero successful self-query misses, every acknowledgement survives reopen,
no failed write present, and revision=1+acknowledgements. Record all failures and
timings. The previous merge-count condition is architecture-specific and not
applicable: it is replaced by checking every acknowledged incremental mutation.
This is not an unchanged implementation comparison or end-to-end agent test.

Record scheduled total, writer wait/preparation/persistence, lease acquisition,
search and observed revision. Reopen and verify vector equality, not only IDs.
Persist complete per-arm results before evaluating gates. No seed retuning,
retries or relaxed limits after inspecting outcomes. Failed outputs are retained.
