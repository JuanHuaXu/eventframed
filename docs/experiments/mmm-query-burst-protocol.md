# Evidence-triggered query timing: frozen exploratory protocol

Status: scheduler contract stage; no efficacy result or production promotion.
All seven research directions remain open. The v120 corpus is consumed data,
including its phase1. This experiment cannot supply fresh confirmation.

## Question and falsifier

The factorial diagnostic finds delay cost in selected shifts. This does not
prove adaptive acquisition helps: fit publication is still every32 frames,
surprise may reflect irreducible noise, and bursts may spend on stale evidence.
The proposed rescue fails if it cannot improve equal-cost controls while
meeting the existing non-harm gates. No threshold tuning after observing results.

## Frozen scheduling contract

Use the original256-frame experiment and four-expert learner/Markov mixture.
Queries occur after the current forecast and any zero-delay natural delivery;
paid evidence arrives at the next clock. A trigger uses ONLY newly delivered
outcomes and their originally issued served forecasts. An outcome is surprising
when its assigned probability is strictly below0.2. Each origin is processed
once, even if both natural and paid delivery become available together.
The trigger includes paid evidence; any self-sustained acquisition feedback is
part of the policy and must be reported, not silently excluded.

There are eight non-carrying budget blocks: clocks0..31 have3 tokens,
32..63 through192..223 have4 each, and224..248 have4, total31.
No query is permitted before clock8 or after248. A trigger permits one query
per clock for that clock and the following3 clocks, subject to remaining tokens.
At the end of a block, force spending when remaining clocks equal remaining
tokens. Empty candidate pools do not consume tokens; missed tokens expire.
Only successful selections consume tokens. Future arrivals, missingness flags,
latent probabilities, case identities and change times are unavailable to the
scheduler. It accepts only current clock, arrived-surprise bit, and pool presence.

Use the same recent32-origin unavailable-evidence pool for ALL new timing arms;
exclude already paid origins and never acquire the current frame. Candidate
selection controls are random, entropy, and weighted expert disagreement using
issued forecasts and current pre-query weights. Compare fixed every8 clocks
against burst timing for each selector; retain natural/no-query control.
The wider pool changes the previous experiment, so its output is contextual
only, not the matched control. Full immediate feedback should buy zero labels.

Integration clarification before quality collection: a paid label older than
the retained mixer journal updates the training evidence and surprise observer,
but does not reopen a censored mixer entry. Record Mixer=false on that arrival.
Both timing arms preserve the same original journal expiry rule. The first
integration test exposed this boundary as an invalid Markov delivery, not as
a quality result. No acquired evidence is removed from subsequent expert fits.

## Fairness and evaluation

Record actual query counts, query clocks, trigger source, empty-pool losses,
fit origins, and arrived evidence. A quality comparison requires equal successful
query counts per paired trajectory. If counts differ, report the mismatch and
do not claim equal-cost superiority; do not discard trajectories to repair it.
The bounded per-block budget is known ex ante, not computed from future labels.
Report stationary burst activation separately from changing-stream activation.
Maintain fixed fits/windows, forecast-before-current-label ordering and as-of
poisoning checks. Preserve original experiments unchanged.

The primary candidate is burst/disagreement. Compare with periodic/disagreement,
periodic/random, periodic/entropy and natural. Use prior all-cell0.01 Brier
non-harm tolerance, and delayed terminal changing cases1,2,4,5,7,8,19,20
mean gain at least0.005 with positive paired lower bound. Mean +/-3.5SE over32
trajectories remains exploratory, not simultaneous or anytime coverage.
Report total selection/training runtime and paid label cost, not just scheduler
cost. No broad success on a single case or control.

## First implementation boundary

The initial change implements and tests only the budget state machine. It does
not yet connect surprise delivery to learner refitting or produce quality data.
The scheduler uses O(1) space/time per frame; pool search and fitting are separate.
