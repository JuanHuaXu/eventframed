# v98 consumed-stream forecast headroom

Read-only diagnostic of the two switch directions, both v97 phases and all32
indices:128 already consumed streams. Verify parent artifact SHA256
6b4559a63d7b6c1e1b2603513a90dcf4c82281b27e48430405522a1fb48380e5,
every parent source hash, and exact reproduction of all128 original records.
No learner or prediction is changed. Full all768-record parent replay was
completed in v97; this diagnostic does not claim to rerun the other640.

Retain256-step pre-outcome forecast traces and bank weights for each view.
Compute metrics in eight32-step blocks: generic64, generic32, actual bank,
oracle convex hull of the four BANK expert forecasts, and best constant
generic64/generic32 mixture. Also report mean expert weights and oracle alpha.
The hull contains generic64,Boolean64,generic32,Boolean32, not unrelated controls.

Oracle forecast is projection of simulator q onto the scalar forecast hull.
Constant-pair alpha minimizes block expected Brier analytically, using alpha0
when both endpoints coincide throughout the block. All oracle computations
occur after the unchanged stream is scored and cannot enter learning.

Verify hull loss<=fixed-pair loss<=both fixed endpoints within1e-12; unit-test
interior, endpoints, ties and invalid numbers. Recompute all diagnostics from
traces in an independent evaluator. Summaries over32 paired trajectories use
z=3.5 approximate intervals as diagnostics only, not a new confirmation test.
No quality gate is reclassified. Full diagnostic replay, race smoke, vet and
source-hash checks are required. No production code or configuration changes.
