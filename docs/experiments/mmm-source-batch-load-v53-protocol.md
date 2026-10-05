# Opt-in source-owner batch-read load v53

Frozen before execution. v52 supports testing amortized source reads, not assuming
loaded success. Four rotated arms per trial, three trials: off; prepared raw
durable/union guard; source owner with point reads and atomic batch discard;
source owner with transactional batch resolution/readback and atomic discard.
The latter uses an explicit new constructor; all default APIs remain controls.

Keep v50's workload and limits unchanged: 192 real Recall calls, four readers,
96 future-dated Observe writes spaced2ms, fixed as-of, 50 overlapping visible
events, K50/pack10, queue64, groups up to four already-ready frontiers, 20ms
guard-entry budget and parent-context callback I/O. Isolated LibraVDB/SQLite with
durable FULL writes. No labels, fitting, private data or production changes.

Batch resolution must retain canonical model/seed/epoch, original identity and
exact immutable retry checks. No caller learner ID enters source admission.
Validate the cold envelope under the real guard and keep actual owner originals
for post-guard readback. Completion waits for all readback and durable discard;
unknown/error/uncertain state cannot become a cache miss or successful completion.

Before measurement, test matched warm training histories, preserved original
forecasts across constructor changes/reopen, missing/malformed batch results,
source conflicts/duplicate requests/cancellation, and before/after-commit
uncertainty. Use equal per-label publication barriers for warm parity, not just
equal final label counts. Run source/control load accounting, full race and vet.

Record all raw request/write/guard/age samples, named phase durations, group sizes,
accepted/drop/expiry/error and original/terminal counts, BatchSourceReads flag,
source snapshots/hashes and runtime metadata in exclusive JSONL. No warmup or
trial exclusion and no threshold tuning after results.

Unchanged screens: each batch-source trial requires accepted-observation age
p95<=250ms and serving-read p99<=1.10 times its same-trial off control. Report
all drops/expiry and write tails; read success alone cannot erase missing work or
writer harm. Integrity must pass with no unexpected errors. Go test PASS means
accounting/integrity, not latency success. Even a finite pass leaves broader
traffic, warm learning, feedback/history authority and whole-direction success
unproven. Preserve every result, including failure.
