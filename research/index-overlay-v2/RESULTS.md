# V2 Geometry Pass, Service Rejection

The24 frozen SQ8 geometry commands complete3072 individual durable commits
and3072 top50 searches. All24cells pass mean tie recall>=.95 and regret<=.005.
Lowest control/candidate recall99.453125%/99.65625%; worst mean regret
.000594102/.000430893. These are synthetic geometry, not actual answer gains.

Mean durable commit cost3.74-4.93ms candidate vs9.05-33.17ms control,
including every synchronous overflow compaction. Quiet-query p99 .574-1.397ms
candidate vs .192-.311ms control: a real serving-cost tradeoff. Initial bulk
construction27.43-135.62ms candidate vs10.81-39.62ms control is slower.
Report retained/RSS NOT CAPTURED; do not call this loaded freshness validation.

Common aligned-key repair passes12cases ordinary/rename/aligned/unaligned,
and candidate/control selected database/core ordinary/race gates. V1 remains
a partial failed study. A broad preflight regex accidentally selected the cost
test without its required configuration; preflight-all-after.txt preserves
that harness selection error. Explicit required preflight list then passed.

SERVICE REJECTION: control16adjacent tests pass ordinary/race. Candidate passes
six service tests and vector hydration, then StoreRoundTripAndAvailabilityGate
panics: Collection.Stats expects standard RawVectorStoreProfile typed fields,
but the research wrapper supplies only extra overlay diagnostic fields. The
service runner stops; no public quality trial or candidate service-race claim.
V2's component benchmark does not prove an integrated usable backend.

Fix the profile adapter in fresh V3 copies, preserve this failure and V2 source
pins, and rerun identical geometry and integration gates. All seven OPEN.
