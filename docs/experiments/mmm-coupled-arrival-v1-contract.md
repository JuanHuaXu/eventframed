# Arrival selector and observation coupling diagnosis

Consumed publication-v1 trajectories192, both schedules, original seeds/fits.
Three arms: unchanged conservative control; arrival-updated outer selector
choosing its own observation mask; same arrival update on the original mask.
Inner selector state is copied from control before each coupled prediction,
so this isolates OUTER selector/acquisition coupling, not all feedback policies.
All arms share actual arrived audit training and external Anti-Pigeon evidence.
Structural split caps the long weight and discards pre-split outstanding credits;
publication alone does not invalidate the experimental algorithm-role losses.

Preserve original control forecasts/outcomes/masks and metrics exactly against
the captured parent. Fixed-view arm must reproduce the independent JS tracking
replay. Score all outcomes, including missing labels; those labels never train.
Record per-frame cost and every arm; no adoption threshold, fresh confirmation,
seed replacement or tuning. Primary comparison is coupled minus fixed-view,
including observation-cost changes. Preserve v104 full-carry failure.
