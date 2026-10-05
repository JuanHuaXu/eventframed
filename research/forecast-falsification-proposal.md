# Version-scoped forecast falsification after v98

Proposal only, not implemented or validated. The consumed v98 diagnosis shows
both available gains and a short useful lifetime for the short-window generic
model. Increasing its permanent prior or choosing oracle block weights is not
an admissible rescue.

## Research basis

[Shafer, Shen, Vereshchagin and Vovk, Test Martingales, Bayes Factors and p-Values](https://www.probabilityandfinance.com/articles/33.pdf)
explain nonnegative test martingales, sequential likelihood ratios and the
distinction between their current values and running maxima. This supports
testing a stated probabilistic forecast law against accumulated outcomes with
an explicit error budget. It does not certify an unrejected model, establish
causality or supply a Brier non-harm guarantee for a downstream gating policy.
The linked author version predates the2011 Statistical Science publication.

## Proposed EventFrame specialization

Test the four raw bank experts separately for each32-step published model
version and observation view. Under the stated null, the expert's issued
probability p_t is the true conditional Bernoulli probability in that view's
pre-outcome filtration. This null is not assumed proven; rejection is evidence
against it. Correlated experts are not independent corroboration.

Use the declared neutral alternative q_t=1/2 initially. For a monitor starting
at s within the version, after receiving outcomes define

    M_s,t = product_{u=s..t} [(1/2) / P_expert,u(Y_u)].

Before s, its value is1. The conditional expectation of each likelihood-ratio
factor is1 under the null. Mix the32 possible starts with fixed equal initial
mass, including dormant monitors at1. Do not replace this mixture by an
uncharged maximum, or restart a fresh unit of evidence whenever a loss is large.

For the first bounded experiment there are4 experts x8 versions x2 views.
Allocate total error budget0.01 across these64 tests. A version-level mixture
crossing6400 rejects that version for subsequent predictions. Account for the
start mixture within the test rather than spending the same error probability
again on every start. Freeze and test probability-endpoint handling; silently
flooring denominators while claiming the original forecast null is invalid.

The error interpretation requires each tested conditional-law null and honest
pre-outcome predictions. Models in these experiments may already be
misspecified, so it is not a guarantee of a1% empirical rejection rate. Nor
does repeated model publication grant unlimited fresh testing budgets.

## Gating and evaluation boundary

Mask rejected versions and renormalize existing bank weights over remaining
raw forecasts; use an explicitly neutral0.5 forecast if none remain. Continue
recording raw forecasts and outcomes for every expert, including rejected ones.
Publish a new version on the existing fixed schedule; non-rejection of that
version is permission to test it, not a certificate of truth. No simulator
change point, future label or oracle headroom enters a gate.

Keep the ungated bank as a paired control. Its old cumulative-loss theorem
applies to its own output, NOT to the gated forecast or neutral fallback.
The new candidate must pass all106 original quality gates and dedicated
rejection/re-entry, missed-feedback, sequence and null-control tests. Measure
gate delay and false rejections, with all64 version tests accounted for.

First validate a literal monitor reference and finite null simulations, then
use fresh design and confirmation streams. A neutral alternative is deliberately
limited: it can detect forecasts losing to ignorance, not every form of
miscalibration. If it fails, preserve that finding rather than retune threshold
or alternative on consumed confirmation. This is a falsification layer inspired
by Anti-Pigeon's role, not a replacement for its external sharing authority.
