# Handoff Boundary V27/V28: Technical Results

2026-10-03. **V27 control assertion FAIL preserved; corrected V28 technical
diagnostic PASS. No early-release implementation or performance rescue.**
All seven whole goals remain OPEN.

## V27 Failure And Narrow Correction

[Original protocol](mmm-handoff-boundary-v27-protocol.md),
[original run](../../research/handoff-boundary-v27/run.json),
[original log](../../research/handoff-boundary-v27/test.log).
V27 expected a changed query to have no belief before any insert changed the
evidence epoch. It failed at `query` with one belief. This was same-epoch
ordinary member-posterior use, not cross-epoch witness reuse. No runtime or
mathematical defect follows. No raw JSON was produced because the test failed
before final artifact emission; failed code, freeze and log stay unchanged.

V28 places the future-only insertion BEFORE query/vector/selection controls.
This makes those controls exercise cross-epoch transport, as intended in the
earlier witness studies. Neither rule, posterior, context key nor compatibility
gate was changed. It is a new separately frozen technical diagnostic, not an
overwrite, successful replay or a new quality confirmation cohort.

## Boundary Evidence

[New protocol](mmm-handoff-boundary-v28-protocol.md),
[run](../../research/handoff-boundary-v28/run.json),
[raw](../../research/handoff-boundary-v28/raw.json),
[race log](../../research/handoff-boundary-v28/test.log).
1135 source/protocol/runner files frozen before test. Generated Go-AST proxy
covers all31current EventStore methods; independent AST test checks observation
first, correct tag and actual delegation. Missing Snapshot is detected. A
deliberate dynamic post-boundary Snapshot is also detected. Interface compile
success alone would not have proven coverage, because embedded methods can
silently inherit uninstrumented forwarding.

| Case | Service-facing store calls | Post-handoff calls | Frontier | Packed | Beliefs |
| --- | ---: | ---: | ---: | ---: | ---: |
| Prime |456|0|150|10|0|
| Learned |456|0|150|10|1|
| Future-only |456|0|150|10|1|
| Changed query |156|0|150|10|0|
| Changed vector |156|0|150|10|0|
| Changed selection |156|0|150|9|0|
| Visible insert |156|0|150|10|0|

Every final recorded interface call is PutBayesianJournal. Native live test
assertions compare full durable journal wire with the captured wire copy and
every packed forecast with its journal decision. The raw diagnostic stores
counts/call names/hash, not full probabilities for an independent offline law
oracle. It does not inherit the normal V26 auditor's scored-law coverage.

Post-hoc call counts locate a cost candidate: accepted/default frontiers perform
150Anti-Pigeon lookups,150posterior lookups and150residual lookups, plus six
common operations. Declined-certificate cases omit the first300lookups. Counts
are not execution times or evidence that those lookups dominate latency.

## Proof Boundary

This observes forecast-computation calls through the SERVICE-facing EventStore
interface. The underlying journal worker necessarily still performs native
database writes/readback and marker/witness commits after handoff. Those are
durability operations, not later forecast dependency reads. No claim that
all storage I/O stops at the boundary is made.

Default config only, optional async processors disabled, seven serial cases
under race, leases unchanged. No concurrent mutation, general optional-capability,
power-loss, historical-journal acceptance or latency proof. Packing uses local
values afterward on this path, but every owned dependency, version transition,
admissibility/cancellation/reopen rule and all original loaded gates must be
established before any earlier lease release. Current JournalSnapshotCompatible
rejects posterior/residual/semantic version motion; that guard remains intact.

Next prospectively examine complete owned handoff artifacts plus historical-
journal semantics, and profile/batch/cache metadata reads only with valid
snapshot/absence/mutation coverage. Do not replace missing evidence with a
blanket version bypass or assume456calls mean456expensive database transactions.
