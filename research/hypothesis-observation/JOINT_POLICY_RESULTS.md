# Joint forecasting/acquisition still meets a finite-model barrier

[Protocol](JOINT_POLICY_PROTOCOL.md), [three-control results](joint-policy-game.json),
[verification](joint-policy-verification.json), [forecast tests](test-joint-policy.mjs).

Unlike the frozen-law game, this oracle chooses the optimal forecast at every
state as well as its next observation. It minimizes weighted Brier risk by
normalizing the weighted joint outcome masses, then solves the observation
recursion. No realized source regime or target is supplied at prediction time.
The candidate is a decision-theoretic predictor, not an ordinary posterior
under the earlier single fixed prior. The uniform mixture samples a complete
forecast/policy pair once before the observations, not after seeing truth.

## Results

| Screen | Numerical lower bound | Final mixture upper bound | Failed rows |
| --- | ---: | ---: | ---: |
| Random + both entropy versions | .000955125216 | .001588103113 | 92/465 |
| Random + archived entropy | .000438703724 | .000975549888 | 31/310 |
| Random + normalized entropy | .000501222337 | .000961739687 | 33/310 |

The two-control diagnostics retain the same128 rounds and learning rate32;
only the control set changes. They are scope-isolation follow-ups on consumed
worlds, not fresh empirical confirmations:
[archived entropy](joint-policy-two-entropy.json),
[verification](joint-policy-two-entropy-verification.json),
[normalized entropy](joint-policy-two-tie_entropy.json),
[verification](joint-policy-two-tie_entropy-verification.json).

The optimized payoff is worst constraint violation: final Brier minus control
minus.01, or area Brier minus control. The whole screen requires nonpositive
final rows and strictly negative area rows. Positive valid lower bounds exclude
feasibility, independently of whether the iterative optimizer converged.

All three lower bounds were independently reproduced to within5.56e-17. They
are strong numerical evidence of a barrier, not outward-rounded real-arithmetic
proofs. Joint prediction substantially lowers the conflict relative to freezing
the forecasts, but does not eliminate it. The extra entropy control is not the
sole cause: each two-control game retains a positive witness.

## Scope and verification

The Brier minimization covers arbitrary four-class forecasts at each state,
including stage-dependent rules. The observation oracle covers deterministic
evidence-adaptive sufficient-state policies; linearity extends each witness to
mixtures. This sufficiency relies on the fixed iid-or-root-copy family. It does
not extend automatically to arbitrary temporal source processes or richer data.

All games replay byte-for-byte. Independent recursive oracles visit48048 states
per witness and use the closed-form minimum m-sum(v_y^2)/m rather than direct
squared loss. Mixture reconstruction checks465/310/310 rows. Weighted forward
scores agree with oracle objectives within6.11e-16. The forecast unit test checks
1024 excess-risk identities, two boundary cases and four invalid inputs;
observation tests exhaust64 contingent policies in32 tiny fixtures plus a tie.

This is model-based design evidence. No claim of empirical calibration,
actual-agent performance, implementation latency or a general intelligence
impossibility follows. It does not erase the original failed screen.

## Next research decision

Do not continue changing acquisition or forecast heuristics within this same
finite class hoping to beat a positive best-response witness. First certify the
witness with outward rounding or exact arithmetic. Then investigate which
additional observable information would remove the conflict, maintaining equal
access and equal costs for the controls. Changing the information model would
be a new experiment, not retroactive validation of this one. Other open research
directions remain active. No production, paper or archived code changes.
