# Ordered admission-wait attribution v33: frozen protocol

Date: 2026-10-01. Research-only Goal 6 diagnostic after
[v32](mmm-recall-pre-feedback-v32-results.md) falsified a
feedback-queue explanation for the 4 ms freshness failure.
V32's dominant pre-admission bucket mixes journal lookup,
selected handoff and waiting for earlier ordered admissions.

Add timestamps only to the v29 research consumer. For each selected
frontier, divide the existing tap-take-to-admission-start interval
into (1) tap take through successful committed-journal lookup and
session validation, (2) lookup end through admission-coordinator
receipt, including bounded channel send/wait, and (3) receipt
through start of guarded ordered admission, including waiting for
earlier indices and request setup. These three nonnegative
durations must sum exactly to the v32 before-admission duration;
all existing v27/v32 phase equalities must also hold.

Run one fresh enabled-only 6 ms fixture and one 4 ms fixture.
Keep 192 full 200-event Recalls, eight workers, 256 future-only
writes, 64 durable bound labels, queue64, selected channel32,
reorder cap32, admission channel16, and the guarded SQLite
WAL/FULL journal. Preserve zero-drop, exact nomination,
no-future, as-of, visible-mutation and durable replay checks.
Report actual offer gap, serving/freshness p99, p50/p99 of all
three new stages, and those stages for the five oldest labels.

Hypothesis: the 4 ms delay is predominantly stage 3, serial
ordered-admission backlog. This is falsified if journal lookup
or handoff accounts for most of the oldest labels' delay.
Do not alter label count, cadence, gate, order or production code.
This is not a capacity pass or performance rescue.
