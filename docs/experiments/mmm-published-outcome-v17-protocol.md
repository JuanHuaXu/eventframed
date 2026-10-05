# Published-LSN Bayesian outcome transition v17: frozen component

Date: 2026-10-02. The v16 full-Recall load fixture has no feedback.
`Store.ApplyBayesianOutcome` durably changes runtime, posterior and
residual versions plus the LibraVDB LSN. Calling it through the
v16 adapter's embedded store would bypass the SQLite published-LSN
marker, leaving the Search gate unready. This is a research-only
publication-lifecycle test, not a full Goal 6 learner screen.

First demonstrate the direct-bypass negative control on a private
READY gate: a committed outcome must move the database LSN and
make the old marker unready. Then add a test-only transition that
serializes with admitted event writes and read-to-journal Recalls,
checks the current READY marker and exact before-LSN, applies the
outcome once, checks the returned snapshot and exact after-LSN,
reads back the outcome and posterior, durably advances the marker,
and only then publishes a new immutable view. The current LibraVDB
transaction can advance multiple LSNs (21 to 28 in
the first observed outcome), so an assumed one-LSN-per-transaction
guard is invalid. An exact duplicate
must not move LSN or marker. Conflicting idempotency content must
not be accepted. A commit/marker interruption must fail closed;
reopen may report READY only when its marker matches the durable
LSN. A direct bypass committed before the transition must not be
laundered. Concurrent out-of-band writers are outside this single-owner
component; their exclusion still needs an enforceable writer fence or a
transaction receipt that identifies exactly which LSN range belongs to
the outcome.

Exercise the transition through `Service.ObserveBayesianOutcome`
after a real committed Recall journal, using full-stream feedback
with observed and available times after that Recall's as-of.
Check that its journal attribution, posterior/residual versions,
durable outcome, published snapshot and subsequent Recall pin
agree, including after reopen. Check an old-as-of Recall does not
consume feedback available later. Keep production source and the
v16 frozen test unchanged. Run focused normal and race tests,
ordinary store/service package tests and vet.

A component pass enables, but does not substitute for, the next
loaded test with concurrent writes, Recalls and outcome deliveries.
That later test must include offered-label to observed publication
age, scored-forecast freshness, no-future leakage and unchanged
<100 ms Recall offer p99 under load. All seven whole goals remain
open until their own evidence gates are met.
