# Consumed publication factorial diagnosis

Use all192 original publication-v1 trajectories and both schedules. No fresh
confirmation claim. Four arms cross withholding newest16 audited training labels
(off/on) with resetting BOTH inner and outer selectors at publication (off/on):
all-carry, all-reset, holdout-carry, holdout-reset. No calibrated weights are used.
Heldout fits still compute and discard their calibration, preserving the frozen
component's objects; this diagnosis is not a compute optimization.

Unchanged and holdout-reset metrics, every issued forecast and latent tape must
match the parent exactly. Preserve same audit/arrival/missingness/expiry/gate/
publication clocks and fit availability. All arms choose their own observations.
Thus factorial contrasts include resulting acquisition changes; they are not
controlled direct effects with observations fixed.

For each phase/case/schedule report post/full Brier for all four arms, reset
harm at each data setting, withholding harm at each reset setting, and the
difference-in-differences interaction. Verify the two algebraic decompositions
of endpoint harm. Reconstruct scores and as-of partitions independently.
Do not use these consumed-data contrasts as new adoption tests, tune holdout
size, discard cases, or replace the original failed results.

Record source and parent hashes. No production, paper, commit, remote or private
data changes. The original seven research goals and criteria remain unchanged.
