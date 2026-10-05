# Actual learning under persistent load v1

Freeze before the first run. Three paired trials, alternating off/on order.
Each arm gets a new temporary persistent LibraVDB store and 50 fixed public dummy
events. Four reader goroutines issue 16 Recall requests each (64 total), recall50,
pack10, with distinct session IDs and a fixed query as-of. One writer inserts32
new events strictly one hour after query as-of, sleeping2ms between writes. Require
at least one write overlapping readers. No production service or private data.

On arm: frontier tap capacity16; actual temporal feedback bridge; one consumer
admits all captured candidates, then releases explicit deterministic fixture
labels to the background learner. Wait for those labels to complete before
consuming the next frontier. Journal/session IDs distinguish repeated load
requests. These repeated fixture outcomes measure workload ONLY, not independent
evidence, calibration, or answer quality. Close the tap after readers/writer stop
and drain the queued observations before closing the learner.

Frozen finite screen, each pair: no read/write errors; overlap>0 in both arms;
on p99 <=1.10*off p99; at least80% of64 frontiers admitted and completed; no
unexpected bridge/worker failures. P99 is nearest-rank ceil(.99*n). Preserve all
raw latency samples, errors, drops and model completion counts, including failures.
The bridge's bounded tombstone capacity is not exceeded by64 distinct journals.

Opt-in test via EVENTFRAME_LEARNING_LOAD_ARTIFACT, exclusively created before
work starts. Record source hashes/text and append each finished arm immediately.
Model-fitting and service timing run without race instrumentation; run separate
race regressions. This is a finite workload screen, not a population p99 bound,
an OpenClaw test, persisted feedback recovery, or a real-task learning result.
