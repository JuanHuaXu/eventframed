# v109: log-score advice, frozen quality protocol

Freeze before fresh generation. Preserve v108's full768-trajectory design:
12 cases, two phase-disjoint rule pools,32 trajectories per cell, paired
immediate and jitter0..31/missing0.2 schedules (1,536 runs). Initial16 labels,
256 scored frames, switches at128, as-of long64/short32 publication every32,
six-coordinate acquisition and separate full training audits are unchanged.

Fresh base2130110900+phase*1000000+case*10000+index*10. Rule/input/outcome/
delay/missingness roles0..4. Audit effective seeds against prior v90-v108
learner/null allocations. V107 used consumed data. No optional enlargement,
replacement seeds, half-life/prior/rate tuning or interim outcome inspection.

## Seven arms

0 generic;1 conservative;2 full carry;3 arrival Brier;4 Brier neutral. These five
must match their frozen v108 counterparts exactly on consumed compatibility
streams.5 log/no-neutral and6 log/neutral are the only new candidates.
Both preserve priors, .001 per-arrival share, published models, gate thresholds,
version policy and observation rule. They replace eta=.5 Brier selector loss
with the Bernoulli log likelihood; there is no unconditional age discount.
The [component contract](../../research/log-advice-component-contract.md)
specifies meaning and limits.

Before each forecast, advance clock, release due earlier labels, expire
unresolved age>=32 entries, and publish as-of models. Acquire all seven arms
before sampling the current outcome. Admit immediate labels afterward. Flush
256..287. Keep issue-time probabilities for feedback, not current-model scores.
Record expected and realized log loss secondarily, using strict full support
and log1p(-p). No silent floor. Proper Brier remains the primary decision score.

## Gates:518 per candidate,1,036 total

Paired mean +/-3.5 standard errors over32 trajectories within each phase/case/
schedule/segment remains an approximate fixed-sample screen, not a confidence
sequence, exact coverage theorem or history-wide multiple-testing guarantee.
The two schedules of one latent trajectory are not independent replicates.

- 480 non-harm gates against each control0..4 for every phase/case/schedule/
  full256 or late128 segment: upper Brier harm<=.01.
- 20 generic gains: full parity3/4/complement4 and late two switches in both
  phases and schedules, mean>=.005 and lower>0.
- 6 conservative gains: delayed late parity4 and both switches, both phases.
- 4 gains each against full carry, arrival and Brier-neutral: delayed late
  two switches, both phases. All gains require mean>=.005 and lower>0.

A candidate passes only if all its518 gates and integrity checks pass. Study
success requires one complete candidate pass, not combining favorable cells
or substituting improved log score for failed Brier. Preserve every gate and
all earlier failures.

## Integrity and performance

Require literal model-path and lifecycle/race tests, exact five-control
compatibility, all-run replay, source hashes, independent Brier/log-score and
as-of origin/journal reconstruction. No oracle gate or age change. Statistical
null guarantees are not established by deterministic replay alone.

Benchmark the bounded fixed-model journal and whole seven-arm fixture only
after quality, replay and verification processes are terminal. Keep all repeats,
including cold ones. No production calls, serving-tail claim, new private data,
commits, pushes or whitepaper promotion. All seven roadmap directions remain
open beyond this finite screen, including real tasks and independent generators.
