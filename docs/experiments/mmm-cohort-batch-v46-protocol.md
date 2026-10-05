# V46 durable cohort batching

Prospective isolated Goal 6 experiment, 2026-10-04. V45 shows journal handoff
dominates Recall blocked stacks; admission waiting and archived owner contention
are also substantial. This does not prove that four-entry commits cause the tail.
Competing explanations include storage latency, capture contention, native index
work, eager publication, and overload. Test rather than assume the batching lead.

Fork the complete test-only joined/archive implementations and serving wrappers
using Go ASTs. Change exactly two collection bounds, four to eight entries; keep
the one millisecond collection timer, admission eight-reader/512-pending caps,
128-job queue, full native readback and SQLite durability, owned original wires,
acknowledgement-after-commit, epochs, cancellation and current-head checks.
No production/default implementation changes. Independent inverse AST comparison
must cover every copied function. Preserve earlier source and negative results.

Run both old cap four and new cap eight for two repetitions of all eight V44
visibility/storage/eager cells: 32 trials. Every trial still offers 128 Recalls,
128 writes, 16 outcomes and 150 nominees at the original rates and worker count.
Alternate cap order by repetition, not by observed outcomes. Exact commands,
sources, host, raw traces and independent checks are retained in an exclusive
root. Old gates remain unchanged: offer/call/outcome p99 below 100ms; outcome
max, writer p99 and used freshness below 250ms; no omitted operations or errors.
Future cases must actually use learned and transported corrections, while
visible cases retain the original fail-closed behavior (not a resolved utility
gap). Two repetitions are a bounded diagnostic, not a population-tail guarantee.

The new auditor may change ONLY the old batch-bound predicate and histogram,
using the recorded declared cap (4 or 8). All chain, law, wire, ancestry, lease,
returned-ack, queue, freshness and physical-copy checks remain inherited. Run
all inherited corruption controls against an actual positive arm, plus invalid
cap and over-cap controls. Reject missing factorial cells and skipped tests.
The setup prime acknowledgement remains an explicitly unbounded historical
instrumentation gap; do not imply a stronger certificate than the trace proves.

Falsifier: eight-entry batches do not form, do not improve queue-inclusive tails,
harm a sibling storage/eager cell, or violate an invariant. No adaptive tuning,
threshold weakening, workload thinning, early lease release, or silent adoption.
All seven whole goals stay open unless their original complete criteria pass.

## Native-Cap Correction

The initial complete run in `research/cohort-batch-v46` is a preserved failed
experiment, not a latency rescue. The worker fork formed batches larger than
four but delegated to unchanged four-entry native commit adapters. Fifteen of
sixteen candidate trials failed closed and omitted offered work; all sixteen
old control trials passed their correctness audit but not latency gates. Their
short candidate timings are unusable. One candidate happened not to encounter
the forbidden size; it cannot establish correctness of the overall design.

The corrected prospective freeze uses NEW roots with suffix `native-cap`.
Fork both test-only native commit adapters as well, changing ONLY their size
bounds to eight and matching diagnostic messages. Worker bounds, native bounds
and all function bodies are independently inverse-compared: two worker and two
native bounds, no marker/receipt/snapshot/validation changes. Add a positive
eight-journal FULL durable publication/reopen test and negative checks proving
the old adapters still reject eight and the new adapters reject nine. Repeat
the entire 32-trial design, without excluding failed cells or weakening gates.
