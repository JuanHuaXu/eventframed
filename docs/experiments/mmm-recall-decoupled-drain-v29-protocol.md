# Bounded decoupled tap drainage v29: frozen protocol

Date: 2026-10-01. Research-only Goal 6 candidate after the
[v28 failure](mmm-recall-ordered-pipeline-v28-results.md). The single
v28 consumer blocks tap drainage during guarded admission and lost
55 observations in a focused race run. That coupling is the leading
explanation, not proof that it is the only bottleneck.

Use the unchanged v28 full-Recall fixture and gates: eight workers,
200 live events, 192 offers at a measured nominal 8 ms cadence,
256 future-only concurrent writes, queue-64 nonblocking tap, single
guarded SQLite WAL/FULL journal, offered indices 0,3,...,189 selected
for exactly 64 synthetic labels, and three rotated learning-off/on
ordinary trials. A dedicated goroutine takes and validates every
committed tap observation, sending only selected observations into
a FIFO channel with capacity 32. The ordered admission goroutine may
hold at most 32 selected out-of-order observations beyond that channel.
The feedback goroutine and its capacity-16 admission channel are
unchanged. These are finite, declared bounds, not an unbounded replay
or a change to the queue-64 serving handoff.

Admissions and terminal feedback must each be in increasing offered
index/`AsOf` order; each ID must be durably admitted before terminal
feedback. Later predictions may omit earlier unprocessed feedback but
must not incorporate future feedback. The tap drainer must not silently
discard unselected observations from accounting: exactly 192 unique
committed session IDs, 64 selected labels, and zero tap drops are
required. Capacity overflow, missing selected frontier, duplicate ID,
stale/backdated admission, or incomplete durable replay fails closed.

Retain all v28 serving, as-of, exact nomination, journal reopen,
visible-mutation, durable ledger/replay, phase-conservation and
freshness checks. Ordinary component pass requires every on-arm
offer p99 <100 ms, pooled on/off p99 <=1.10, and pooled frontier and
feedback ages p99 <250 ms. Separately run a focused `-race` lifecycle
trial; race-build timings are not compared with the ordinary gates.
If the drainer still drops work, distinguish tap->selected-channel
backpressure from journal lookup and producer rate before another patch.

A finite pass would not prove arbitrary lateness, never-arriving
frontiers, cross-epoch transfer, power-loss recovery, real-agent
outcomes, or OpenClaw serving. Production remains untouched.
