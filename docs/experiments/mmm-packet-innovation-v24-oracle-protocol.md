# Packet innovation v24: post-hoc exact model diagnostic

Date: 2026-10-02. The primary v24 screen has failed its aligned-protection
interval in both splits. Preserve that decision. This additional analysis
integrates over every possible one-label-per-member vector under the declared
independent Bernoulli model; it does not fit a proposal or replace the screen.

Before computing the diagnostic, require score separation strong enough to
prove each innovation proposal always packs the first ten positive monitored
members in baseline order, filling any shortage from the highest unmonitored
members. Check that all monitored positive scores exceed every unmonitored
score, monitored negatives remain below the ten highest unmonitored scores,
and within-sign score order follows baseline order. Require the current policy
to demote every monitored member below those ten unmonitored scores even at
its most permissive possible elastic scale .5; its actual scale lies in
[.5,2.5]. A failed separation check prevents this closed-form interpretation.

Let P_i(k) be the distribution of successes among the first i monitored
outcomes, capped at ten. Update that distribution by the exact Bernoulli
recursion. The expected utility sum contributed by the first ten positive
members is sum_i p_i^2 Pr(N_{i-1}<10). The first p_i is the probability that
the training outcome is positive; the second is usefulness of an independent
future outcome. Add each final shortage probability times the utility sum of
its highest unmonitored fillers, then divide by ten. Compute the baseline and
current expectations directly from their fixed packets.

Test the dynamic program against exhaustive enumeration on small populations,
including all-zero, all-one and shortage cases. Recompute this conditional
expectation for all 48 recorded hidden populations and report it alongside
the empirical outcome-world screen. The exact expectation relies entirely on
the synthetic independence and fixed probabilities. It supplies no external
calibration, safety, confidence-sequence or agent-task guarantee. Do not move
any primary threshold or call either proposal accepted from this diagnostic.
