# Expert-weight / loss-ledger two-by-two ablation

Motivation: fixed-share improves continuous weights but terminal harms persist.
Separate static versus fixed-share weights from global versus32-step ledger.
Freeze all four cells, no rate/allowance tuning. Use the same672 trajectories
and172,032 stored forecasts. Keep .01 per step and alpha_t=1/t, eta=1.

In the local-ledger cells, expert weights remain continuous and use arrived
feedback from all earlier origins. Only the guard ledger resets at fixed
32-step boundaries. Earlier-block predictions remain covered by their prior
worst-case reserves; neither missing labels nor pending losses are reclassified
as successes. They simply cannot supply credits to a later block. Each block
and each global prefix must satisfy its realized bound. Do not claim an
arbitrary sliding-window bound or conditional expected non-harm.

Compare all four cells, the existing reset-both32 control and incumbent.
Report whole-stream and terminal64 effects, scenario harms and pointwise
trajectory uncertainty. This is consumed-data mechanism diagnosis and a
possible rescue, not untouched validation. Model fitting costs unchanged.
