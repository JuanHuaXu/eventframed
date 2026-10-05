# Published-LSN batch journal v15: frozen Goal 6 rescue screen

Date: 2026-10-02. V14 measured roughly 5.43 ms median for one
LibraVDB journal write and 0.98 ms for readback under 4 ms Recall
offers. This test-only candidate batches journals in one native
LibraVDB transaction. It is distinct from v22's SQLite journal
with 8 ms dwell and a separate publication guard.

## Architecture and invariants

Freeze batch cap at four (the Recall worker count), maximum dwell at
1 ms after the first queued journal, and a bounded 500 ms internal
transaction context. Service callers retain the v10 read-admission
lease from Search through durable acknowledgement. The batch worker
holds the adapter owner permit, checks the READY marker and exact
latest LibraVDB LSN, verifies every entry's captured snapshot and
as-of against the current Store state, and checks duplicate IDs and
payload conflicts both within the batch and against stored journals.
It inserts only new records in one LibraVDB transaction, checks that
the state snapshot is unchanged, reads back every entry, then commits
one SQLite marker-LSN transition. No caller receives success before
both durable transitions and readback. If either commit has uncertain
outcome, fail closed; a later exact retry or reopen must resolve it.
The batch worker must drain on orderly Close. A canceled caller may
receive an uncertain error while the transaction finishes; it may
not receive success before commit. No production code is changed.

## Controls and load

Before load, test exact duplicate retry, conflicting same-ID payload,
mixed-horizon rejection, interruption after the LibraVDB transaction,
and reopen/READY behavior. Verify all-or-none insertion on a
transaction error. Report actual batch-size distribution and
durable acknowledgement order.

Retain v10's private 256D rows, 128 visible writes and 128 full
Recalls at nominal 4 ms offers, four Recall workers, cap-16/16-ms
event writer, published-LSN Search, full residual-enabled service,
per-version top-150 oracle, and all freshness/semantic checks.
Run two fresh normal trials and predeclared race-correctness mode.
The unchanged gate requires 128/128 acknowledged writes and
Recalls, zero stale/oracle/journal/future/ack-before-offer violations,
writer offer-to-ack p99 <250 ms, Recall call and offer-to-done p99
<100 ms, and published-view max age <250 ms. Record actual offer
gaps, queue wait, batch p50/p99, journal transaction cost and final
READY state. A pass is a finite private component, not Goal 6
completion or a production deployment. If any gate fails, preserve
the negative result and diagnose without retuning this cohort.
