# Six-millisecond confirmation and four-millisecond phase trace v31

Date: 2026-10-01. Research-only Goal 6 follow-up to the
[v30 rate screen](mmm-recall-cadence-v30-results.md). V30 passed
one matched 6 ms pair and failed 4 ms frontier freshness despite
zero tap drops. This protocol does not modify the learner or the
frozen v30 gates.

Run three independent 6 ms matched learning-off/on pairs with
rotated order off/on, on/off, off/on. Retain the same 192 full
200-event Recalls, eight workers, 256 future-only writes, queue64,
selected-channel32, reorder cap32, admission-channel16, guarded
SQLite WAL/FULL journal, 64 selected labels and 6 ms offer cadence.
The confirmation component passes only if every enabled trial
finishes 192 unique frontiers, 64 durable labels, zero tap drops,
all as-of/nomination/replay/mutation/phase checks, measured median
offer gap within 25% of 6 ms, on-arm offer p99 <100 ms per trial
and pooled, pooled on/off p99 <=1.10, and pooled frontier and
feedback ages p99 <250 ms. Report each trial and pooled metrics.

Separately run one fresh enabled-only 4 ms diagnostic fixture.
It must preserve all lifecycle/no-future/durable checks and report
offer gap, offer p99, frontier age p99, and p50/p99 for tap wait,
tap-take-to-feedback-offer, guarded feedback, and publication
wait. Log those four components for the five labels with greatest
frontier age, preserving per-label conservation. Do not add
component p99s to infer a critical path. If 4 ms now passes
freshness, call the v30 failure nonrepeatable pending more runs;
if it fails again, identify the dominant component from paired
per-label traces. This diagnostic is not a second confirmation of
the v30 screen or a license to change its threshold.

Production remains untouched. Even confirmation at 6 ms is a
finite synthetic capacity result, not arbitrary bursts, missing
frontiers, recovery, real-agent outcomes, or whole Goal 6 success.
