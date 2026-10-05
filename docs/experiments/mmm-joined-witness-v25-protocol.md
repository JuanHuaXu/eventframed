# Joined Journal And Witness V25: Prospective Protocol

2026-10-03. All seven original goals OPEN. V24 fair rotation fails all eight
normal screens: outcome391-407ms; offered Recall173-192ms. No adoption or
consumed-cohort scheduling retuning. This tests a different work-reduction
mechanism under the original workload and gates.

## Pre-Patch Reasoning And Boundary

Confirmed: V23 native journal worker first commits its marker, then each
reader reacquires the owner for a separate witness JSON copy/transaction.
V24 changes ordering but does not remove this serialized work. Competing
causes are native journal serialization, duplicate SQLite commits/full-state
copying, and full read-to-ack writer exclusion. Existing profiles and both
future/visible controls support investigating publication work, not assuming
this will resolve every source of latency. Falsifier: joint publication fails
unchanged latency gates or violates native byte readback/provenance semantics.

New test-only adapter and separate copy of the sealed V15 native batch method.
No upstream production patch is proposed; existing V15/V23/V24 sources are
unchanged. No installation, production/private sessions, live configuration,
whitepaper, global skills or Git publication. Authorization is the active
isolated research goal, not a broad production change.

## Mechanism And Safety

Both arms enable V23 conditional transport and use ordinary original admission,
not V24's failed external scheduler. Control retains the V24 cancellation
terminalization rule and original two-stage publication. Joined stops only the
isolated idle native worker before serving, starts one bounded replacement
worker (128 jobs; batches<=4; dwell1ms), and copies the submitted wire journal
and complete query/vector/frontier/snapshot/as-of binding before enqueue.

Under the original owner, prepare one bounded copied witness state for all
batch bindings. Perform the unchanged native commit receipt and full byte
readback. The SQLite marker update, hash-chain record and whole witness-state
update commit in ONE FULL SQLite transaction. Only then publish the immutable
view and acknowledge all callers. Whole read-to-durable-ack leases remain held.
Already-canceled calls do not submit; an accepted handoff terminalizes before
returning caller cancellation. No early delivery or mutable private-key reuse.

[SQLite WAL](https://www.sqlite.org/wal.html) supports transaction atomicity,
but this is NOT one transaction across libravdb and SQLite. Native-before-
marker/witness interruption must leave no acknowledged packet, clear the live
view, and remain non-READY after actual reopen. A fully committed joint SQLite
transaction can reopen as READY even if its prior caller never received an ack.
This distinguishes inspection-open from serving-ready; no marker repair.
No hung-device or power-loss claim is established by injected failures.

Witness caps remain10000 mutations,512 bindings,200 source keys. Missing or
conflicting bindings fail closed; stored posterior epochs never change. Reuse
remains conditional on exact actual request and complete as-of runtime history.
Visible motion still lacks admissible bounds and blocks reuse. Synthetic
certificate coverage is assumed, not demonstrated. Copy/hash cost still grows
with retained bounded witness state; no billion-record complexity claim.

## Controls And Frozen Cohort

Before freeze: race tests for valid future/visible reopen, changed query/vector/
selection and future evidence; cancellation in a blocked handoff; four-journal
atomic witness/readback, duplicate and conflicting bindings; native-after-DB,
before-witness rollback and after-SQLite interruption; unaccounted/gapped runtime
must block. Preserve compile/test mistakes and repair only the new fixture.

Eight NEW normal trials:2reps x joined/control x future/visible. Identical V24
workload:200 eligible normalized256D+17future rows initially;128writes and128
full Recalls independently offered4ms apart,8read workers;16mixed outcomes
independently offered16ms apart to one worker. AsOf=offer; fixed query/settings;
no dropping, load throttling, condition-on-success filtering or label tuning.
Event batches<=16/16ms; journal batches<=4/1ms in both arms.

Unchanged gates: write offered p99<250ms; Recall call AND offered p99<100ms;
outcome offered p99<100ms/max<250ms; view max<250ms. All128/128/16 work items
must acknowledge. Exact150nominee oracle, no future labels, pin/journal wire
agreement, Beta arithmetic, immutable sources, hash-chain continuity and
runtime conservation. Future must have genuine cross-epoch served beliefs;
visible must transport zero. First scored use and censored labels reported;
later drain cannot be mistaken for a fixed-window freshness win.

Freeze all internal Go sources, this protocol, collector, runner and independent
checker before normal collection. Checker reuses sealed V23 law/nomination/
timing arithmetic unchanged; additionally verifies all129 journal bindings,
wire hashes, class attempts, batch bounds/unique IDs, batch-to-chain links,
acknowledgment chronology and publication counts. Corrupted tapes must reject
missing work/binding, future/pin/source/chain violations, forged timing/gates,
duplicate batch IDs and premature ack. Original V23/V24 tapes remain sealed.
Adoption requires BOTH joined future repetitions pass all gates and controls.
No accuracy, real-agent outcome, general error-control or whole-goal completion
claim from this synthetic loaded serving study.
