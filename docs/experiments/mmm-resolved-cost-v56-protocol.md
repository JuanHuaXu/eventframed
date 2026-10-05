# Resolved admission isolated cost v56

Frozen before measurement. v55 removed one duplicate admission preflight but
failed two of three loaded age screens. Admission means improved in two trials
and regressed in the third; unchanged cleanup code also varied. Separate
primitive admission cost from service/queue scheduling before another patch.

No runtime changes. Compare OpenSourceOwnerBatchReads against
OpenSourceOwnerResolvedAdmissions. Both use VerifyAndDiscardBatch. Two fixed
batch sizes50/200, cold and warm states, three trials, two rotated arms:24 cells.
Each cell has32 cycles of fresh admission, exact retry and verified cleanup;
all phases are measured separately with monotonic clocks. Request construction
and test assertions are outside spans. No service traffic or concurrency.

Warm cells receive64 deterministic synthetic labels with a publication barrier
after each before measurement. Both arms must have identical hashes of training
originals and measured originals. Cold cells get no labels. No fitting or labels
during timing; this isolates trained prediction cost, not concurrent learning.
Seed42, epoch1, deterministic public source keys/features/times, same requests
within each comparison. Actual returned worker originals must match retries,
expected IDs/readiness, stored readback during cleanup, and first/last readback
after replay. Terminal retries after reopen must all be true. Label, pending,
queue and ID counts must survive reopening. Record bytes and restart duration,
which are separate from per-operation spans and not bounded-history claims.

Capture all raw timings, source snapshots/hashes, model-history hashes and
runtime metadata in exclusive JSONL. Run fixture accounting/race checks first.
Do not overlap measurement with other tests initiated by this task.

Diagnostic screen only: at200 events, new-admission mean must improve at least
10% in every cold/warm trial; exact-retry mean may not regress more than5% in
any50/200 cold/warm trial. All identity/forecast/replay checks are mandatory.
Report all phases, sizes and trials, even if the screen fails. These are frozen
engineering thresholds, not statistical confidence bounds or new latency SLAs.

An isolated pass cannot rescue v55, prove writer non-harm, or complete direction6.
If cost improves but load does not, inspect matched arrival schedules and phase
contention; if not, investigate where time moved before adding another feature.
All seven directions remain open. No production, private data, push or deployment.
