# V10 shift-failure information diagnostic

Frozen 2026-10-01 after observing the v10 failure. This is a consumed-data,
hindsight diagnostic, not a new confirmation gate or a predictor candidate.
Use only the v10 confirmation records for `shift128` and `shift256` in both
input modes. Preserve all 24 trajectories per cell and all four arms.

For each archived frame, reconstruct the role-0 input and outcome RNG from
the frozen v10 evaluation seed. Verify exact equality with the stored outcome
and verify every final observed mask/value agrees with the reconstructed
input. Fail the diagnostic on any mismatch. Recompute the outcome probability
from the declared pre/post majority rule and scenario noise. The full-input
oracle is this probability at the actual nine-bit input.

For each arm's *actual selected view* `(mask, value)`, compute the ideal
conditional forecast by enumerating all 512 inputs under the declared
uniform or latent-correlated input mass, restricted to that mask/value. The
oracle may know the true mechanism and change time, but the selected view must
be the one the arm actually acquired, not a hindsight-selected view. Report
realized Brier for model, selected-view oracle and full-input oracle, both
over all post-change frames and the first 64 post-change frames. Report mean
observed-bit count, support/mass checks, the source hashes and runtime.

The view oracle isolates information loss from the observed coordinates
*relative to this declared generator*. Model minus view-oracle Brier mixes
fitting/calibration error with sampling noise and is not a causal effect.
The full-input oracle is an unattainable reference under the six-coordinate
budget. Do not use these hindsight probabilities to fit, select or score a
new arm on the same data. These findings cannot validate generalization,
real-agent benefit or latency.
