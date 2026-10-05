# Mixed-mutation feedback boundary v1

Frozen 2026-10-01 before the combined service test. This is a correctness
preflight for research goal 6, not a serving-tail or throughput screen.

Run the same sequence on the in-memory and real persistent LibraVDB stores,
through the research publication wrapper and the actual service/tap/temporal
feedback bridge. Seed one visible event and bind the original bridge. Recall
and admit one as-of-now frontier. Add a future-only event: feedback for the
visible candidate must still complete, while scoring at a time when the new
event is visible must fail. Recall/admit a second as-of-now frontier. Delete
the original visible event through Service.Delete: the old bridge must reject
its pending feedback, score and further admission. Its worker count must not
increase from the rejected label.

Add one new visible event, bind a fresh bridge, recall a fresh journal, admit
its frontier and complete one explicitly supplied label. Verify the fresh
bridge's worker completion and that the old bridge still rejects its old
pending feedback. Record all identities and errors; no label should be
inferred from absent feedback, and no future label may update the first
bridge. Re-run with `-race` and package vet. Production and service defaults
remain untouched.

Passing establishes this finite mutation/rebind boundary only. It does not
establish automatic crash recovery, arbitrary mutations, sustained mixed-write
load, p95/p99 serving latency, or agent-task answer quality.
