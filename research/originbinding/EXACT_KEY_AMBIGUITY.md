# Exact origin keys do not certify independent evidence

Status: negative research diagnostic. No serving or packing policy change.

The earlier public fixture showed that two externally bound records with the
same producer and lineage can have distinct exact relation keys yet be merged
by the legacy packet-correlation fallback. A tempting rescue is to let two
different verified keys bypass that fallback. The paired diagnostic in
`fallback_test.go` falsifies that rule on the same fixture mechanism:

| Second observation relative to `A > B` | Bound keys | Legacy correlation | Required interpretation |
| --- | --- | --- | --- |
| `A  >  B` | different | correlated | formatting repeat; do not count twice |
| `A < B` | different | correlated | distinct, conflicting relation; do not erase either |

Both payloads have valid fixture-created bindings to one origin, one producer
and one tool lineage. Both occupy one packet slot under today's default
selector. Merely prioritizing key inequality would separate both pairs: it
would repair the conflicting-relation example while falsely treating the
formatting repeat as fresh independent evidence. The test is an explicit
counterexample to that **rule**, not a proof that relation-aware packing is
impossible. It checks the raw `Registry.Apply`, `epistemic.Correlated` and
`packing.Select` boundaries; it does not assert that an agent saw both facts.

The safe next contract needs a trusted capture witness for claim/relation
identity or explicitly distinct observed occurrences, bound to the event
payload and source lineage. A different hash is only inequality of bytes.
When either witness is absent, the existing conservative fallback remains.
Any future override must preserve genuine repeats and Anti-Pigeon certified
separation, and must be tested with contradictory operators, formatting
variants, unbound/tampered records and actual packet selection. The offline
registry is supplied by the fixture controller; it is not a real provenance
authority or an agent-answer validation set. Retaining two conflicting records
for inspection must not count them as independent corroboration.

No production code, daemon default, whitepaper or remote changed. Goal 5
remains open: this diagnostic prevents one unsafe rescue but does not produce
untouched, outcome-labeled agent improvements.
