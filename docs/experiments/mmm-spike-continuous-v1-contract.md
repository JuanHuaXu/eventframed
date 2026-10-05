# Uninterrupted mixture experiment

Freeze before collection: source indices0-7, all21scenarios/two phases/two
schedules, all256frames. Same spike prior, 64-origin window, fresh fit on
admitted-window changes, cap1024 and predictive integration settings. Expert
prior half, learning rate1 and excess-loss allowance .01 per issued forecast
remain unchanged. 672 trajectories,172,032 forecasts.

Retain expert feedback and the pending-outcome budget ledger across the whole
256-step trajectory. Compare with the same forecasts processed using resets
every32steps, Markov alone, unguarded continuous weights and the fixed half
mixture. Reset controls drop earlier-block feedback, exactly as previous
block experiments did. No current/future outcome access and no missing-label
reserve release. Use stored issue-time expert predictions only.

Reuse the audited raw forecast blocks at0,128,224; fit the missing blocks at
32,64,96,160,192. This is a compute-saving replay, not free inference: account
for every fit in all eight blocks, including cached work. Each block's initial
fit is recomputed in the defined collection method even if the window did not
change at the boundary. Since the fitter always initializes from its frozen
prior, equal retained windows must produce exactly equal state; assert that
in the collector. This extra boundary work is charged. It does not reset the
expert weights or guard ledger in the continuous policy.

All data are consumed exploratory data. Report whole-stream, each32-frame
block and terminal64 scores, harms and trajectory-bootstrap uncertainty.
Global realized budget bounds prefixes, not arbitrary subwindows. Do not
mislabel a late-window loss as violating the global bound or hide it under
early accumulated gains. Preserve iteration-capped states and flag them.

Validation: source/key/order checks on cached blocks; all-clock as-of unit
tests; independent moments and origin audit for new raw fits; explicit
continuous-vs-reset negative control; prefix audit including hidden outcomes;
poisoning of unavailable labels; replay of the aggregation. All production
systems, endpoints and the whitepaper remain untouched.
