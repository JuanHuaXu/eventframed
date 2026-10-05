# Unchanged deadline-scheduler replication v39: frozen protocol

Date: 2026-10-01. Research-only Goal 6 replication after
[v38](mmm-recall-deadline-gate-v38-results.md) passed learning
freshness but missed the frozen writer-tail ratio by 2.9
percentage points. Do not change the scheduler, its 200 ms
frontier and 30 ms writer soft deadlines, test fixture,
or v38 success margins.

Run four fresh matched pairs at nominal 4 ms offers, rotating
control/deadline, deadline/control, control/deadline,
deadline/control. Each arm uses eight workers, 192 full
200-event Recalls, 256 future-only writes, 64 bound durable
labels, queue64, selected channel32, reorder cap32,
admission channel16 and one guarded SQLite WAL/FULL journal.
Require exact nomination, no-future/as-of, mutation rejection,
journal reopen, durable replay, 64 labels, 256 writes and zero
tap drops. Log actual median offer gap, Recall offer p99,
frontier and feedback age p99, writer offer/call p99,
writer total time and writer starts during the offer window
for every arm. Compute pooled metrics and each matched pair's
writer-p99 ratio.

Reapply the unchanged v38 pooled gates: every candidate trial
Recall p99 <100 ms, pooled candidate/control Recall p99 <=1.10,
pooled frontier and feedback age p99 <250 ms, writer offer p99
and summed total time each <=1.25 times control, and candidate
writer starts during offers >=80% of control. Report whether
every candidate trial also keeps frontier age <250 ms as a
robustness observation; do not substitute that for the pooled
gate. A replication pass does not retroactively erase v38's
failed run: it would establish mixed evidence requiring
independent load-shape confirmation. A repeat writer-tail
failure disfavors further tuning of this deadline pair.

These are fresh runtime trials on the same deterministic
synthetic corpus, not independent data generators or real-agent
outcomes. Production remains untouched.
