# Bound worker v7: motion-proof diagnostic

Status: frozen pre-run diagnostic, no daemon or serving change.

Question: does unchanged-source validation alone distinguish a harmless
future-only ingestion from a backfill that changes the as-of observation set?
The answer controls whether a target-independent epoch seal would be safe.

Fixture: one witnessed original forecast bound to a committed query journal
and a source event. Capture its snapshot and prediction time. Repeat on the
in-memory and persistent event stores under the research publication wrapper.

1. At the original target, both event-source validation and publication
   compatibility must pass.
2. Ingest a new event available strictly after the prediction time. The
   original source remains valid and the publisher's as-of compatibility must
   pass.
3. Ingest a different event available before the prediction time. The original
   source event remains byte-identical, but the publisher's as-of compatibility
   must reject the old snapshot. Record whether source validation still passes.

If step 3's source validation passes, it is insufficient by itself to authorize
cross-snapshot reuse: a v7 rebinding protocol must additionally persist and
check the original dependency snapshot's as-of motion. Do not weaken the v5/v6
exact-target seal before this result is known. No latency or answer-quality
claim follows from this diagnostic.
