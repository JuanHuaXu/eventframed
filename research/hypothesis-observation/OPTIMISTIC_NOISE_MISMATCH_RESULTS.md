# Optimistic transfer fails upward noise misspecification

Under the [frozen protocol](OPTIMISTIC_NOISE_MISMATCH_PROTOCOL.md), all predictions
retain the .20 measurement model. Actual noise varies only in the evaluation
world; neither true noise nor counterfeit mask enters the predictor.

| Actual noise | Gates passed /180 | Protection failures /160 | Genuine gain failures /10 | False-confidence failures /10 | Worst Brier harm vs local |
| --- | ---: | ---: | ---: | ---: | ---: |
| .10 | 180 | 0 | 0 | 0 | -.001931 |
| .15 | 180 | 0 | 0 | 0 | .003938 |
| .20 | 180 | 0 | 0 | 0 | .010000 |
| .25 | 94 | 78 | 8 | 0 | .016650 |
| .30 | 79 | 91 | 10 | 0 | .023101 |

Overall 713/900 gates pass: the combined stress screen FAILS. Negative harm in
the .10 row means even the worst cell improves over the local control. Both
candidate and local control remain misspecified in the .25/.30 worlds; this is
not comparison against an oracle that knows the true noise.

All-counterfeit false-confidence reduction passes in every tested world.
That does not imply proper-score protection: the two measures diverge.
At .30, all-genuine gains range from -.003181 to .001332, below the .005
minimum for every allocation. This is not merely a confidence-interval miss:
these are exact expectations under the finite generating laws.

## Interpretation

The prior 180/180 result remains correct in its stated matched model, but does
not establish robustness to an underestimated error rate. The model's finite
ambiguity family covers counterfeit masks, not arbitrary measurement noise.
Its conditional certificate was never a certificate for these outside-family
worlds. This experiment changes that assumption rather than finding an algebraic
error in the projection bound.

The failures motivate expanding protection beyond a fixed likelihood family.
Do not widen the harm allowance or tune noise to make this screen pass.
Passing five grid points would not establish coverage between them, nor under
omitted roots, adaptive selection or arbitrary real observations.

## Verification and scope

Full output replay is byte-exact. The .20 results (including all candidate cells)
are deep-equal to the archived optimistic-transfer run. An alternate conditional
risk identity agrees with direct latent-state squared-loss summation for all
3200 Brier values, maximum absolute difference 6.11e-16; gate decisions agree.
This checks scoring algebra and deterministic replay, not an independent
implementation of inference or the measurement model.

Artifacts: [full results](optimistic-noise-mismatch.json),
[verification](optimistic-noise-mismatch-verification.json),
[verifier](verify-optimistic-noise-mismatch.mjs).

No serving performance measurement was made, no production changes were made,
and no whitepaper claims were promoted. All seven research directions remain open.

## Next falsifiable rescue

Test an outcome-simplex guard: constrain regret against each of the four
point-mass outcome laws, while preserving the same local baseline, optimistic
target, .01 allowance and all 900 gates. Brier regret is affine in the target
law, so vertex protection implies protection for every distribution over the
declared four outcomes, without trusting a measurement likelihood. The remaining
question is whether this stronger protection leaves enough movement to pass the
genuine-gain tests. This is our elementary convexity argument, not a borrowed
empirical guarantee. It does not promise accuracy, authenticate reports or cover
outcomes absent from the defined scoring space.

