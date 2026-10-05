# One fixed-size higher-precision replication of evidence routing

Historical proposal, now executed in [v102](../docs/experiments/mmm-replication-v102-results.md).
The fixed-size run passes its finite screen; v101 remains FAIL under its own
protocol. No further sample-size ladder is proposed for this screen.

Proposal, not a new validated result. v101 remains FAIL104/106. Both failed
recovery cells now meet the mean-gain threshold but not the positive confidence
lower bound. Changing the algorithm again would confound an untested precision
question with another mechanism change.

## Why one replication is justified

In v101 parity-to-majority, the paired routed-minus-generic improvement means
are0.007642 and0.005140, with standard deviations0.013895 and0.010221 over32
trajectories. The approximate z=3.5 interval half-widths are0.008597 and0.006324.
The same-run improvements versus comparative-only also have positive descriptive
lower bounds in both phases. Those are consumed diagnostics, not substitutes
for the frozen generic-control success criterion.

A fixed128 trajectories per cell would halve these estimated interval widths
if variance remains similar. This is a planning approximation, not a guarantee
of significance or of a mean above0.005. In particular, a true effect right at
the minimum threshold can fail the empirical mean criterion even with more data.

## Proposed freeze

- Exactly128 trajectories per phase/case, all12 cases and both views:3072 full
  streams, not only extra samples in the two favorable-looking recovery cells.
- All twelve v101 arms, frozen routing kernel, priors, model windows, fit cadence,
  rejection budget and evidence volume per trajectory remain unchanged.
- All106 quality gates remain: same mean thresholds, non-harm limits and paired
  z=3.5 construction. Report every gate, not only the two formerly failing ones.
- Predeclare fresh disjoint seeds before any generation, perform no interim
  quality look and no early stopping, and replay all records exactly afterward.
- Repeat the same four finite null screens with fresh seeds; preserve the
  stronger per-outcome comparative-state equality check.
- Do not pool v101 with the new run, relabel its failure, select a favorable
  phase, increase n again after seeing outcomes, or claim anytime confidence.

After this one replication, decide from its fixed result whether evidence is
still too weak, mean benefit is insufficient, or protection fails. An inconclusive
result remains inconclusive; it does not authorize an endless sample-size ladder.
Mechanism-level work on model freshness and short-window variance remains a
separate lead. Independent generators, delayed/partial observation, real tasks,
external certificates and persistent serving are still required by the roadmap.
