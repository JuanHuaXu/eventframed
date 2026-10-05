# Finite member-rate shape V35: results

2026-10-03. **Overall component FAIL**: 4/32 design cells and 3/32
confirmation cells fail the unchanged frozen quality screen. All seven
whole research goals remain OPEN. No production or whitepaper changes.

## Evidence

Fresh bases 2026103503/2026103504 produce 2,048 worlds, 61,440 snapshots
and 4,915,200 conditionally independent trial observations. Each of six
learners receives the same tape; learner copies are not extra independent
observations. There are 32 independent worlds per geometry/regime cell.
Intervals are mean +/- 3.5 SE, not time-uniform or simultaneous AP bounds.

Both full audits replay all original issued laws, rates, outcomes,
nomination probabilities, weights, packet rankings and metrics. Separate
batch Beta integrals and direct two-atom products verify the final laws.
The independent arithmetic is within Go, not a separate-language auditor.
Eight precollection source hashes, separately hashed postcollection auditor
and benchmark artifact are recorded in the audit JSONs. Six corrupted-tape
controls and a changed-source control are rejected. Future-outcome flips,
ownership, cache isolation, underflow, replay/cap, atomic normalization,
four-module race tests and vet pass.

Raw SHA256:

- design: `de7b24c11acbe38e498c0d4f722795b3e18fa86fd575d0c00e5f784406b9d728`
- confirmation: `f553375d2e38edd3de61ed1182d82488ceb638727258a5e1aa93a8a1af45c369`

See [protocol](mmm-shape-v35-protocol.md),
[design audit](mmm-shape-v35-design-audit.json),
[confirmation audit](mmm-shape-v35-confirmation-audit.json) and
[benchmarks](mmm-shape-v35-benchmarks.txt). Raw JSONL files accompany them.

## Component Repair and Limits

At 2,400 genuine trials (16/member), confirmation whole-frontier Brier:

| Regime | V34 adaptive | Shape | Shape packed usefulness | Shape signed packed bias |
| --- | ---: | ---: | ---: | ---: |
| Tight independent (.2/.8) | .168895 | .160862 | .800000 | -.000000443 |
| Wide independent (.2/.8) | .168845 | .161024 | .800000 | -.000001451 |
| Wide off-grid (.27/.73) | .208583 | .201852 | .730000 | -.000237 |
| Wide asymmetric (.1/.8) | .144530 | .139704 | .800000 | .042854 |
| Wide triple (.15/.5/.85) | .177867 | .177867 | .850000 | .060604 |
| Wide bounded uniform | .220249 | .220918 | .773650 | -.028907 |

Both independent geometries pass the stricter .05 packed-bias bound,
positive whole/priority gains over adaptive, and all control protections in
BOTH splits. Confirmation tight gain .008033 has lower .007453; tight
bias magnitude upper .000002074. Learned width-.30 posterior mass is
.999989 tight/.999961 wide. This repairs the measured two-rate shape
failure, not every possible member distribution. Off-grid/asymmetric cells
also pass both splits. Triple-rate data largely retain Beta2; continuous
uniform data do not show a consistent new advantage.

At one trial/member, shape weights stay at their frozen prior. Earlier
predictions can differ because the declared prior family changed; they are
not evidence of learned dispersion. Shape posterior mass is not an
authenticated group, causal conclusion, or Anti-Pigeon certificate.

## Remaining Failures

Design fails tight calibrated, tight mean_shared, tight curved and wide
bounded_uniform. Confirmation fails the same first three cells:

- Tight calibrated: confirmation whole gain .004703 and priority .004545
  fall below the declared .005 mean improvement.
- Tight mean_shared: whole gain .002465 and priority .002304 fall below
  .005, despite positive lower bounds.
- Tight curved: usefulness gain versus local is -.005109, lower -.012824,
  outside the -.01 protection tolerance. Design lower is -.011890.
- Wide bounded_uniform: design usefulness lower -.010205 narrowly misses
  -.01; confirmation passes. Keep this failed design evidence.

No gate, prior or collected source was retuned.

## Post-Hoc Oracle Headroom

[Headroom audit](mmm-shape-v35-headroom.json) is a postcollection diagnostic,
not a selection rule. It verifies raw hashes and uses known generator
rates ONLY in the evaluator. The irreducible future Bernoulli Brier floor
is mean p(1-p), with the same priority weights. Control risk minus this
floor is an upper bound on any forecast's realized-cohort mean gain.

In BOTH tight calibrated/mean_shared cells and splits, even this oracle's
whole and priority mean gain is below .005. Confirmation calibrated oracle
gain is .004711 whole/.004553 priority; mean_shared is .002485/.002323.
Thus those fixed .005 mean gates are unattainable on these realized cohorts.
This is not proof about every future population, and does NOT turn the
original failed screen into a pass. The useful remaining quality lead is
nonlinear/packet protection, not forcing a gain beyond available headroom.

## Cost and Scope

Apple M4, Go1.27.1: 150 predictions 13.907-14.727 us, update 9.650-9.687 us,
zero allocations; construction 66.393-66.812 us and 500993-500994 bytes
(six allocations). Update benchmark includes state restoration. Update
is about 6.5 times V34's ~1.48 us. Cache storage is O(N*378), not constant
in frontier size.

Maximum per-model phase sum is 30.702 ms design/39.987 ms confirmation,
within this candidate's predeclared 50 ms component cap. That is NOT V34's
old 25 ms cap, loaded serving latency, acquisition cost or durable freshness.
Stationary components do not prove shifts, delayed evidence, useful valid
AP splits, untouched real-agent improvement, or equal-total-cost observation.

Next: test a bounded delayed-evidence identity/arrival adapter without
modifying these sealed sources, then pursue shifted and loaded integration.
