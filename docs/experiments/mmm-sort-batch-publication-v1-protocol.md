# Batch incremental publication v1: frozen private component protocol

The test-only, single-owner incremental sort-key gate may accept 1..16
same-tenant pre-embedded EventFrames. LibraVDB commits the new subset in one
receipt-bearing transaction. The gate must bind every new event to that one
commit LSN, append each row hash in input order to the SQLite journal, and
advance one READY marker transaction only after the whole DB batch commits.
Exact duplicates cause no new journal entry; a duplicate-only batch causes
no DB commit or marker change. All input events must retain per-event runtime
version and ingestion motion, even though they share one publication LSN.

Functional controls: all-new batch; exact retry; mixed duplicate/new batch;
conflicting duplicate; unkeyed or bypassed write; injected interruption
after DB commit and after SQLite commit; close/reopen verification; an
exact-LSN as-of query that includes an earlier eligible event and excludes
a future available-at event. Failed or uncertain cross-DB transitions must
deny READY until reconciled; do not silently republish an unjournaled row.

After functional controls, run a private paired isolated cost screen. Start
two independent Stores with the same 259-row or 1,027-row event tape. For
each size, compare eight 16-event groups: control performs 16 receipt-bound
journaled single appends; candidate performs one receipt-bound journaled
batch append. Rotate arm order by group. Report group completion p50/p99,
sum of arm times, candidate/control ratio, and quiet capture p99. Neither
arm shares a database path or overlaps another owner. No acceptance gate is
imposed on this exploratory component.

This screen is not loaded Recall, a 4 ms offered-rate experiment, immediate
per-event acknowledgement, multi-owner safety, or a power-loss proof. A
batch acknowledges all its events at completion, so any later offered-rate
study must include batching dwell and queue time. Preserve negative results
and keep production untouched.
