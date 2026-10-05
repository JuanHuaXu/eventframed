# Guarded source-owner load v50

Frozen before execution. Nine rotated cells: three trials each of off, prepared
raw durable + union validation + post-guard readback/discard, and source owner
v49 with the same union guard and post-guard readback/batch discard. No serving
forecast influence, feedback, labels or model fitting. Synthetic public fixture
only, isolated real LibraVDB and SQLite, no production instance/config changes.

Keep the v42/v43 workload: 192 Recall requests, four readers, K50/pack10, 96
future-dated Observe writes spaced 2ms apart, fixed as-of, 50 overlapping visible
events, queue64, groups of up to four already-ready frontiers, no wait to fill.
Admission remains under the real service publication/as-of guard with a 20ms
entry budget. Actual callback I/O uses the parent context. Every accepted original
is read back and durably discarded before completion; no hidden cleanup backlog.

The source arm accepts no caller learner ID. Validate cold preview envelopes
under the guard, then retain actual source-owner outputs for readback. Check
their IDs are consecutive and unique for the distinct-journal fixture, reject
unexpected retries or warm records, and compare all other original fields. Never
write a relabeled preview. Exact retry/restart behavior has separate v47/v49 tests.

Record full request/read/write/guard/age and named phase samples, attempts,
acceptance/drops/expiry/staleness, original/terminal counts and group sizes with
source hashes in exclusive JSONL. Freeze the source/control choice per cell and
rotate cell order; no warmup/trial exclusion. Both durable arms retain prepared
FULL-commit writes and all per-journal/query/feature/publication checks.

Screens remain unchanged: each source trial must have completion-age p95 <=250ms
and request-read p99 <=1.10 times its same-trial off control. Report completion
counts, drops/expiry and write tails alongside those screens; do not hide writer
harm behind a read-only pass. Zero unexpected errors or record mismatches are
mandatory. Go test PASS establishes accounting/integrity only, not these latency
screens. No population guarantee, warm learning success or whole research
direction completion follows from nine finite cold cells.
