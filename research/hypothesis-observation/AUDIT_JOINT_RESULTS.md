# Joint noisy-audit model: remaining numerical barrier

[Protocol](AUDIT_JOINT_PROTOCOL.md), [game](audit-joint-game.json),
[verification](audit-joint-verification.json).

Optimizing forecasts and acquisition together with the mandatory noisy source0
audit does not pass this finite screen. The128-round mixture passes160/160 final
nonharm comparisons but120/150 positive-area comparisons. Thirty area rows fail.

The best numerical lower bound on worst violation is .0003260816432802405;
the mixture upper bound is .0007974899248260609. The independent recursive oracle
returns .000326081643280296. Positive lower bounds do not require convergence,
but these are floating-point witnesses, not outward-rounded certificates like
the separate old-model certificate. The old certificate is not transferred to
this different information contract.

## No future-audit shortcut

Each of16 root patterns has one pre-audit forecast, followed by both possible
audit-result subgraphs. The pre-audit joint masses have no audit conditioning.
The mandatory audit transition has four identical internal aliases solely to
reuse the existing four-action oracle; none selects or skips an outcome.
Post-audit forecasts and choices may depend on the observed result.

There are41200 states:41184 after-audit states plus16 pre-audit states. The audit
consumes2 credits and leaves five renewals. Six pre-action losses and one terminal
loss preserve the16-credit contract for candidate and controls. Forecasting and
query decisions are optimized; there is no statewise safety guard left to blame.

## Verification

Full byte-exact replay,310 mixture-score checks and source hashes pass. The
separate recursive oracle explicitly scores pre-audit loss once, then sums the
two audit branches, rather than using the main graph's aliases. It visits all
41200 states. The graph audit checks20295680 class-mass branch equalities and
all action/level transitions, including audit marginalization and terminal depth.
Weighted forward/backward objectives agree within7.22e-16.

The hypothetical channel's.8 sensitivity/specificity remains an assumption.
Nothing here validates an actual source-authentication method. Mandatory audit
cost may consume useful observation capacity; it is not a free information gain.

## Research scope

The greedy pilot's failure was not merely insufficient planning: a numerical
best-response witness remains after joint optimization. More heuristics inside
this exact model are not a justified next step without falsifying the witness.

However, this universal per-regime screen is stronger than direction7's original
statement, faster learning at equal acquisition cost. Its failure must stay
visible but must not be presented as impossibility of that broader research
goal or of actual-agent learning. Existing matched-model successes also remain
limited evidence. Before adding more hypothetical sensors, prioritize an
observation/provenance mechanism whose information and reliability can actually
be measured, and define the deployment/task population explicitly. Any new
experiment must preserve honest case-level reporting, fair control access and
costs, and cannot retroactively turn these failed screens into passes.

All seven whole directions remain open. No production, paper or frozen source
changes were made.
