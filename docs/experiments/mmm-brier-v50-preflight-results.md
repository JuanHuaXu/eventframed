# V50 Brier primitive: correctness preflight

2026-10-04. Research-only, isolated from the sealed V49 normal candidate.
The primitive implements the binary strong Brier substitution described in
[the exploration note](mmm-brier-v50-exploration.md), with positive static
expert priors and a single unresolved, owner-bound ticket. No interleaved
delayed forecasts or ordinary Bayesian-posterior interpretation is claimed.

`research/brier-v50-preflight/` preserves four exact source copies, race/vet
commands, terminal logs, and a separate source/log/test-root hash readback.
Both commands completed with code zero; all six required test roots executed,
without skips. Correctness work ran during V49's excluded reference-audit
phase, not during candidate/control scientific timing.

- Independent two-outcome positive-part root search agrees with the binary
  closed form in 1,024 deterministic cases; endpoint/domain, outcome symmetry,
  and per-outcome loss domination checks pass at declared numerical tolerance.
- All 4,096 twelve-bit label paths under three priors check the immediate
  regret inequality against every expert at every prefix. Advice depends only
  on already revealed labels; this is a finite numerical test, not a new proof.
- Foreign, forged, duplicate, stale-epoch, and backward-time tickets are
  rejected without mutation. Original advice is owned, not aliased. Pending
  and total-trial caps, invalid inputs, and epoch resets pass adjacent controls.
- A 2,048-observation reversal preserves recovery of log weights after severe
  linear-weight underflow. Extreme finite priors and common log offsets pass.
- Sixteen outcome-fork positions leave forecasts identical before the divergent
  label is revealed. No future outcome enters the prediction API.

No broad world experiment, sample-efficiency win, priority improvement,
packet/agent benefit, adaptive recovery, or serving-latency claim follows.
Performance and delayed/nested integration remain separate work. The current
hybrid's possible loss/objective mismatch remains a hypothesis, not an isolated
cause of its quality failures. All seven whole goals remain open.
