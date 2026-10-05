# As-of prediction-error evidence boundary

Before training an error-informed critic, verify a strict observation projection
on all2688 consumed v120 records. Decision clock160; issued origins128..159;
forecast channel10 is the 64-label segmentation model fitted at128 and issued
before each current label was read. This is an auxiliary historical learner,
not a calibration certificate for the newly refitted query-decision law.

The replay adapter may inspect delivery metadata solely to determine whether
a label arrived by160. It outputs issued origin/X/probability plus observed
origin/label/arrival for received labels only. It never exports Q, source identity,
future delay, permanent-missingness, unresolved Y, or publication-after160 data.
Current frame160's label is excluded even if immediate. In a runtime integration,
the same boundary would use actual receipt records, not retrospective metadata.

Freeze ten descriptive features for each nominated origin152..159: observed
coverage; observed positive fraction; mean Brier; signed mean residual mapped
to[0,1]; coverage in152..159; Brier in that recent interval; local kernel weight
divided by32; local Brier; local signed residual mapped to[0,1]; mean observed
age divided by32. Local weights are1/(1+9-bit Hamming distance to the candidate).
Empty diagnostics use positive fraction.5, Brier.25, mapped residual.5, age1;
coverage/weight0 explicitly signal absence. These are observed-feedback summaries,
not unbiased full-stream risk estimates or statistically certified confidence.

Tests: hand-computed losses, empty and delayed-boundary cases, strict field
traps, current/future/unarrived-Y and all-Q poisoning, unresolved delivery changes,
ownership, invalid projection inputs, and a non-noop arrived-label change.
Across every source record, independently enumerate eligibility, compute features
through a different reference, and verify all candidate vectors and ownership.

Additionally reconstruct each historical128-clock segmentation fit from as-of
evidence and compare all32 issued probabilities to channel10. Poison excluded
fit-time outcomes on selected fixtures. Record maximum numerical discrepancy,
source/code hashes, race tests, complete counts, and byte-exact projection replay.
Do not infer critic efficacy or a whole-goal pass from this component boundary.
No production, paper, commits, or pushes.
