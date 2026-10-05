# Rate adaptation diagnostic v77

Post-hoc mechanism analysis of v76, not fresh confirmation and not a rescue.
Retain all 10,240 v76 streams and source hashes. No production edits. Compare
fixed .25, actual predictable, and simulator-only oracle rates, all on the
actual candidate's query and augmentation snapshots and selected observations.
The oracle knows the current generator law, including change time; it is NOT
an implementable detector. It changes rates only, never the observation policy.

All generator probabilities are multiples of 1/60. Independently enumerate the
60 midpoint outcomes for each channel to recover its ternary law exactly.
For each sign compute true expected log-factor and its maximizing rate on the
same all-channel factor-safe interval as v76. Use 32 derivative bisections.
Test enumeration against explicit probabilities for every scenario and phase,
and compare the optimizer against a 1001-point grid.

Record per-stream first alarms for fixed, actual and oracle arms; verify the
actual rate tape hash and first alarm against v76, and fixed first alarm against
v76's fixed-augmentation arm. Record favorable-sign mean rate, zero-rate count,
and true expected log-factor in six windows: last64 pre-change steps, post
[0,32), [32,64), [64,128), [128,256), [256,384). For null scenarios only the
last64 steps exist. Means use actual window sample counts, not padded zeros.
These are conditional one-step growth diagnostics, not forecast Brier scores.

Lag is supported by a large early oracle-minus-actual growth deficit that
shrinks later. Persistent estimation loss is supported by a late deficit.
Carryover is relevant when early-window growth does not explain delay: all
three arms retain the same eight starts and no hindsight reset is allowed.
Do not claim a unique cause from aggregate comparisons. The oracle isolates
rate knowledge but retains history-dependent augmentation and allocation.
No success criterion from v76 is relaxed or retrospectively relabeled.
