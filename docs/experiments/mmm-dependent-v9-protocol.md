# Dependent-input v9 protocol

Frozen before evaluation. This is a specification-assumption stress test, not a
claim that the existing uniform missing-bit integration is correct off-generator.
No production behavior or v8 source changes.

Run the unchanged v8 four-arm policy under three input distributions:

- fair: nine independent Bernoulli(.5) bits, reference control;
- biased: independent bits, even positions Bernoulli(.2), odd positions (.8);
- clustered: latent fair bit, each coordinate independently flips it with .1.

All distributions have full support. Fit the incumbent on 4096 samples from the
matching distribution. The correlated case is not an unseen covariate shift:
the observation policy's uniform missing-bit assumption is the suspect. Preserve
v8 label equations, noise, audit probability, six-coordinate observation budget,
rolling64/long256 fitting, inner and outer weights and pre-outcome journaling.
Do not give any arm the hidden full input before its forecast. Full input is
recorded solely for audit/replay after forecasting.

Four scenarios: stable05, shift128, recurring, interaction (original indices
0,2,6,8). Two fitting seeds, four streams per fit, two splits, three generators:
192 streams of512 frames. Fit base2026104201, design2026104202,
confirmation2026104203. Seed case index=10*generator+original scenario;
existing role encoding applies. No adjustment between splits. Design and
confirmation have separate streams but reuse the two fitted incumbents; this
is not independent training-set confirmation.

Report paired per-stream full/post Brier and accuracy for all four arms,
by generator, scenario and split. Post starts at128/256 for changed scenarios,
and256 for stable. Primary stress-screen criteria for each retained arm:
every generator/scenario/split full and post mean Brier harm <=.01 versus fixed;
every generator/split shift128 post mean improvement >=.005. Preserve failed
criteria. These are finite pilot screens, not confidence bounds or sufficient
completion evidence for direction1. No parameter selection from these results.

Audit: fair-input stream parity against v8, replay determinism, raw input and
label regeneration, <=6 observed coordinates, forecast bounds, no label delivery
before its forecast, and source hashes. Runtime is instrumented fitting time,
not a claim about serving latency. Possible follow-up: estimate a conditional
input distribution from past audits; do not silently replace it with the oracle.
