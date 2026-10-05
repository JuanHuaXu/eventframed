# Request lease scheduling experiment

Freeze before running. Repeat the prior controlled-ingestion fixture with the
correct nomination bound: N through N+16 for in-window writes, exactlyN for
future writes. Preserve earlier failed artifacts. Add paired scheduling arms
without leases and with request-level weighted-semaphore leases. Four permits;
each Recall holds one, each CaptureTurn writer holds all four. The existing
golang.org/x/sync/semaphore dependency provides context-aware acquisition.

This admission wrapper exists in the research runner only, not the daemon.
It prevents concurrent ingestion while a leased Recall computes and commits its
journal. It is not a change to snapshot compatibility or retry count. All readers
and writers in this experiment must participate; uncoordinated mutations remain
outside its protection. A read permit spans the entire Recall, not just the
pre-commit region. It is a deliberately conservative scheduling candidate.

Two repetitions; N50/200; four workers; future/in-window writes; ordinary/task+
lexical mode; lease off/on.32 arms,1024 reads,512 writes. Reader duration begins
BEFORE acquisition; writer duration likewise includes wait. Record wait separately,
all errors, stale rejections, full-frontier availability, nomination and journal
checks. Warmups and seed captures remain outside measured phases.

Finite gate: no leased service/write errors or future leakage, no leased stale
rejections, and report whether either reader or writer maximum exceeds100ms.
Also report read and total completion time, so read gains cannot hide writer
starvation. This is not an open-loop fairness/latency guarantee: each writer
waits5ms after its prior completed write; admission delays change that schedule.
The total completed work remains fixed. Future writes use leases too, explicitly
exposing the cost of unnecessarily blocking compatible ingestion.

Do not promote to production on a pass. Durability, uncoordinated writers,
tenant isolation, cancellation under queued admission and broader load remain
separate requirements.
