# V3: Geometry Pass, Integration Rejected

All24 frozen geometry commands finish:3072 individual durable commits and3072
top50 queries across32/128D,256/1024initial records and3repetitions per arm.
Every mean tie-recall/regret screen passes; minimum candidate mean tie recall
.9965625. Independently recomputed row arithmetic and10evaluator/tamper checks
pass. These are consumed synthetic geometry measurements, not population bounds.

Candidate mean commit3.760-4.889ms; paired mean commit ratios .146-.431.
Candidate quiet query p99 .684-1.239ms. Maximum candidate commit142.531ms:
the synchronous compaction spike exceeds100ms even though the mean improves.
Peak RSS and peak retained payload NOT CAPTURED. Initialization and whole test
allocation appear separately in evaluation.json; oracle/logging allocations
must not be called index-only allocation.

The actual service/store preflight FAILS. Both control ordinary/race16case
checks pass. Candidate ordinary passes14cases but returns zero results in
TestStoreRoundTripAndAvailabilityGate and TestDeleteRetentionCompactAndBackup.
The runner stops; no candidate service race or public quality trial runs.
The six service cases are memory-backed; the ten store cases use embedded
LibraVDB. Core/table checks did not prove this sink boundary was correct.

Investigation in fresh V4 reproduces ErrEmptyIndex from an empty physical base
before the live overlay is visited. Preserve V3's frozen sources and failures;
V4 is a separate repair, not retrospective evidence that V3 integrates safely.
No loaded freshness, background compaction, total-RAM bound, recovery guarantee,
agent utility or WHOLE goal is validated. All seven goals remain OPEN.
