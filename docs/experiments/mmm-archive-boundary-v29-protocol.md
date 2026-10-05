# Historical Owned Archive V29: Technical Protocol

2026-10-03. Prospective technical experiment, not a fresh quality/latency cohort.
Production, older frozen adapters, and all seven whole-goal gates unchanged.

Previous goal turn: concrete progress (V26 eight load trials and V27/V28 boundary
evidence). All seven goals remain OPEN. Weekly usage checked at 7%.

Patch gate: load failures confirmed; read-to-ack exclusion contributes blocking,
but native append and owner duration remain alternative bottlenecks. V28 only
establishes the default-service boundary, not an early-release theorem. Chosen
hypothesis: owned handoff can end admission without changing historical laws;
falsifier: an intervening acknowledged mutation changes journal/packed laws,
blocks until journal ack, admits stale selected evidence, or invalidates reopen.
No upstream deployment repair: this is an isolated, non-shipping research fork.

One test-only archive adapter captures full-wire JSON and request/vector/frontier
binding while external/native read admission and owner establish the current
published snapshot. The capture seals original provenance and full-wire digest.
Release read admission ONLY after that copy. Append independently under owner;
check complete durable hash-chain ancestry from genesis through capture to the
current published head. Archive original snapshot unchanged. This permits
historical record insertion, NOT residual/posterior/certificate current validity.

Go AST generation copies V25 native append, changing precisely its admission
predicate to the owned-capture predicate; original byte readback, native receipt,
snapshot/LSN checks, marker transaction, and failure behavior remain. Archive
capture row and witness row commit with the marker in one FULL SQLite transaction.
Cross-database atomicity is NOT asserted. Default service and original selected
feedback guards stay unchanged. Current full-stream feedback may use a historical
commitment; selected feedback after epoch motion must fail.

Five barrier cases: future-only insert, visible insert, ordinary outcome, outcome
plus visible insert, and cancellation plus ordinary outcome. In each, mutation
must acknowledge before archive barrier releases, reader count must be zero,
caller must not return early, saved wire and original snapshot must agree, and
native/witness reopen must preserve them. Non-cancel packets pack ten of exactly
150 recorded decisions with byte-identical laws. Cancel must persist its accepted
capture but deliver no packet. Instrument all31EventStore methods as V28 did;
no service-facing store access after Put handoff in tested default configuration.

Six authority corruptions: wrong owner, changed wire, changed snapshot, changed
frontier binding, unknown root, corrupted committed chain. Every one must reject.
Freeze runtime source, generator, generated file, protocol, runner, and checker
before exclusive raw/log outputs. Independent checker verifies full journal hash,
snapshot consistency, temporal barrier order, packed laws and original validity
controls. Timing under race is diagnostic, not a latency/adoption result.

Research basis: [SQLite isolation](https://www.sqlite.org/isolation.html) describes
snapshot isolation and rejects stale read-to-write upgrades. This design does not
upgrade a stale database transaction: it inserts an owned historical observation
in a new current write transition. That distinction is an EventFrame research
contract and needs tests; SQLite alone does not prove multi-store correctness.

Remaining before loaded adoption: batch archive publication, overload bounds,
closure/cancellation under concurrency, interrupted commits and reopen controls,
optional asynchronous configs, and the unchanged 100/250ms/freshness load gates.
