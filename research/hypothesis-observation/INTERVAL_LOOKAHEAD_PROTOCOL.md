# Frozen two-query interval acquisition pilot

Extend the one-step interval-aware observer with a receding two-renewal
lookahead. At each decision minimize the model-expected sum of posterior
target-class Gini risks after the next two observations, permitting the second
source choice to depend on the first hypothetical outcome. With one renewal
left, use the existing one-step criterion. Depth is capped at2, total
history length10, no free observations.

Keep the joint model, prior, four initial reports and16-credit budget from
INTERVAL_ACQUISITION_PROTOCOL.md. Policies are lookahead, target(one-step),
random, entropy and fixed. Presample paired tapes. Fresh seed base2026091402,
two splits, three noises(.10,.20,.30), four masks(0,5,10,15),64 episodes/cell:
1536 episodes. Do not drop the difficult mask5 case.

Primary lookahead comparisons retain the preceding84 gates versus random
and entropy: final paired lower gain>=-.01 in all cells; learning-area paired
lower gain>0 in masks0,5,10. Add24 final nonharm gates against one-step across
all cells. Total108 gates must pass. Report paired area/final differences
against one-step descriptively too; no post-result threshold changes.
Intervals remain mean +/-3.3 SE over64 trajectories, descriptive rather than
simultaneous/confidence-sequence guarantees.

Check independent depth2 enumeration, depth1 equivalence, terminal bounds,
history immutability and prefix replay. Full run and rerun must match.
Hypothetical branches use the declared predictive model, never the actual
generator or unobserved tape. Account acquisition and compute separately:
this protocol does NOT claim equal planner CPU or production performance.
Prior failed studies remain unchanged. No production work or publication.

