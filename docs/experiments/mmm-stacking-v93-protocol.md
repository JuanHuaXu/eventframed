# Exact LOO stacking v93: frozen component protocol

Frozen before results. Test the binary Brier stacking specialization in
research/predictive-stacking-proposal.md. Use exact leave-one-out subset and
parity predictions with all likelihood weights recomputed without the held-out
label. Optimize a convex mixture weight on the sum of LOO squared errors.

Four paired arms: unchanged generic subset; skeptical-prior BMA (parity prior0.1);
unregularized stacking lambda0; regularized stacking lambda1. The sum objective
adds lambda*a^2, not a mean-loss-scaled penalty. Lambda1 is a provisional one-unit
penalty toward the generic default, not a tuned or universally justified value.
Only lambda1 determines the v93 advance verdict. Report all other arms under
the same criteria. No alternative winner selection or post-result retuning.

Use v92's ten cases,64 fits per phase/case, nested16/32/64 labels, disjoint
phase-specific feasible rule pools and full/fixed-mask63 exact expected scoring.
Seed base2038119300 plus phase*1000000 + case*10000 + index*10 + role, roles0..2
as before. Verify disjoint roles from v90/v91/v92. New fits, same finite generator
family. All3840 four-arm records, stacking weights and prediction/training hashes
must be stored. Reuse fitted LOO components for the two lambda arms; no outcomes
or scoring-domain information enter that shared fit.

## Unchanged requirements

For parity3/parity4/complemented parity4 at n32/n64 in BOTH phases, require a
positive lower mean-minus3.5SE bound on
(L_control-L_candidate)-0.2*(L_control-L_oracle), full-input expected Brier.
For ALL cases, phases, sample sizes and both views require the paired upper
mean-plus3.5SE bound on candidate-minus-control Brier <=0.01. Small sample and
non-parity failures can veto advancement. Report every observed regression.
Keep all predecessor verdicts; no gate changes based on these data.

Bounds are approximate across-fit normal summaries, not confidence sequences.
A pass only authorizes a separately frozen delayed-stream integration test.
Whole-direction requirements, v88, real-world tasks and serving are unchanged.

## Validity and cost

LOO fits remove the held-out label from both counts and subset evidence. Verify
against explicit refits on small and contradictory fixtures; compare held-out
predictions when only the excluded label changes. Prove the quadratic weight
solution with independent objective checks, including zero disagreement,
endpoints and finite parameter/input bounds. Check partial mixture coherence,
symmetry and immutable snapshots. Require2..256 eligible samples.

All labels in the iid window have arrived before fitting. LOO is not an as-of
historical forecast; a temporal extension needs forward/blocked validation and
regime-shift tests. Weights are predictive optimization weights, not posterior
model probabilities. Input-law misspecification and sparse validation uncertainty
remain possible. No causal or calibration guarantee follows from stacking.

The consumed v92 majority index37 tail may be checked as a diagnostic only.
Do not remove it or treat it as fresh confirmation. Report tail harms on fresh
fits too. Freeze source/protocol hashes, replay, race/vet, and benchmark exact
shortcut versus explicit refits, as well as full fitting and lookup. Benchmarks
exclude observation/queueing/publication; no concurrent experiment. No production,
OpenClaw, install, commit or push changes.
