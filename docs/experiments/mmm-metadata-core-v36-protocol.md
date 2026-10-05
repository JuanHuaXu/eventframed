# Head-Scoped Immutable Metadata Core V36

2026-10-03. All seven goals OPEN. V35's16normal trials failed original adoption
gates; its frozen auditor prime-target failure and post-hoc repair are preserved.
No production/private corpus/whitepaper/frozen source changes. Usage9%.

## Hypothesis And Boundary

V35 removes repeated getter-owner reads but JSON-copies4.16-4.41MB/trial.
Competing native I/O/admission/durable publication costs remain. No proven sole
cause. V36 tests immutable head-scoped reuse rather than another request copy.
The chosen invariant is that runtime-head motion invalidates the read core;
journal-only publication preserves Head/Log/Sources/Certificate AFTER first
certificate binding. V23's native implementation changes that certificate only
from nil to its first value, so a nil-certificate core is never cached. Runtime
mutations are excluded by the original read admission interval. Source changes
without head motion are not supported; unaccounted native changes must reject.

Cache at most one deep-copied core, atomically published and never mutated;
each request retains its own active token/scope/as-of/nomination and expiry.
Old core roots may precede later journal-only publications but MUST be verified
full committed prefixes with read fields identical at the request's head.
Sealed Search/native validation and all native metadata getters/full posterior
equality stay intact. No FULL durability, journal wire/readback, chain check,
native receipt or queue work removed. Native gap/poison/foreign/expired scope
fail closed. No in-place mutation or exposed core maps in runtime code.

One cache pointer plus active reader references; deep-copy O(log+source bytes)
per built head, not per request. History/source caps unchanged. Recorded
diagnostic build/proof arrays grow with trial size and are not a bounded-memory
production observer. Owner acquisition, builds, hits and PHYSICAL copied bytes
measured. Existing Metadata.CopiesBytes is physical, not logical per-request
proof bytes. Independent auditor adapts only that legacy byte-check input, then
checks physical bytes against separately recorded builds. Never relabel a
counter to hide copy work. Additional diagnostic formatting remains measured.

## Frozen Tests And Evaluation

Before normal trials: inverse-AST full workload equivalence; native law/decision
differential; first-certificate nil transition; same-head reuse; runtime motion;
old-root current-head fields; no borrowed mutable authority maps; query/vector/
old-as-of rejection; expiry/handoff; native gap and poison; barrier/cancel/Close,
reopen and race. Inherited V35 three-getter and validity semantics stay frozen.

16NEW trials:2reps x future/visible x joined/archive x request-copy/shared-core.
All underlying arms use V35 projection; request-copy is the comparator. Same
128writes/Recalls at4ms/8workers,16mixed outcomes16ms/150nominees, full100/250ms
original gates and source/law checks. Fixed arm order/two reps limit timing
attribution; completed harness is not adoption. Preserve every negative result.

Independent frozen prefix replay, raw law/wire/nomination/audit checks and core
build/reuse/source/timing/counter controls. Prime key/independent ack limitations
remain explicit; delivered controls must target recorded delivered requests.
Frozen auditor must pass its own corruption controls before normal data.

Inspiration: [Go immutable copy-on-write](https://go.dev/src/sync/atomic/example_test.go)
and [RCU publication/reclamation](https://www.kernel.org/doc/html/v6.6/RCU/whatisRCU.html).
This is Go-owned immutable reuse, not Linux RCU or an inherited latency guarantee.
