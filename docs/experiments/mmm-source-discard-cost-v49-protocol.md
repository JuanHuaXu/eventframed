# Source-key batch discard v49

Frozen before execution after v48 exposed per-item cleanup as the largest phase.
Retain v48's 50/200 sizes, three rotated trials, 32 cycles per cell, cold actual
originals, exact retries, readback, restart and full accounting. Both arms now use
the source owner and its unique index. The control retains per-item Discard; the
candidate changes only cleanup to the new source-key DiscardBatch. Admission,
retry and per-source original readback stay identical. No raw durable escape
hatch, durability relaxation, label, fitting or production access.

Batch discard resolves every source, then uses the existing bounded atomic
terminal batch with snapshot preflight. Same source lookup work, one terminal
transaction instead of N. Test missing/duplicate/early/zero/conflicting-terminal
and canceled late members, mixed retries, reopen, immutable identity retention,
and before/after-commit acknowledgment uncertainty before measurement. Reuse the
original single-item behavior as a negative control, not silently replace it.

Record exclusive JSONL with complete source/hash snapshots, BatchDiscard flag,
all phase samples and runtime metadata. Do not remove warmup samples or select
favorable trials. Correctness is mandatory; Go test PASS alone only certifies
those checks. Performance screen: reduce mean whole-cycle time by at least 20%
at size 200 in EACH trial, not merely the discard phase. Report size50, all phase
p95s, restart and disk size. A pass warrants further integration investigation,
not loaded service success; the original 250ms age and serving non-harm targets
remain unchanged and still need actual queue/guard/retrieval/load measurements.
