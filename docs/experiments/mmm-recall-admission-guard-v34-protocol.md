# Guarded admission sub-operation profile v34: frozen protocol

Date: 2026-10-01. Research-only Goal 6 diagnostic after
[v33](mmm-recall-admission-wait-v33-results.md) located the
4 ms freshness backlog in the selected handoff upstream of
serial guarded admission. Before changing any authority or
durability rule, split the guard's elapsed time.

For each selected label record five monotonic intervals whose
sum must exactly equal guarded admission duration:

1. guard entry through candidate/source validation to callback entry;
2. `durable.AdmitBound`, including its ledger read, forecast and append;
3. `durable.Admission` readback;
4. `s.ValidateResearchAdmission` of the persisted record;
5. callback return through guard exit.

Run one fresh enabled-only 6 ms fixture and one 4 ms fixture with
the unchanged v29 learner: 192 full 200-event Recalls, eight
workers, 256 future-only writes, 64 bound durable labels,
queue64, selected channel32, reorder cap32, admission channel16,
and single guarded SQLite WAL/FULL journal. Require all existing
zero-drop, exact-nomination, no-future, as-of, mutation rejection,
journal reopen, replay and phase-conservation checks. Report
actual offer gap, serving/freshness p99, p50/p99 per sub-operation,
and the five oldest labels' corresponding durations. No guard or
storage operation may be skipped for this diagnostic.

The lead is that durable append or duplicate post-record source
validation accounts for much of the roughly 18 ms serial
admission. If neither dominates, the lead is rejected. Even a
dominant step is not automatically removable: a later rescue must
preserve its invariant or replace it with an equivalent checked
contract, including no-future and durable authority. Production
remains untouched.
