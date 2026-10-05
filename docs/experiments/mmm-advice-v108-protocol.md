# v108: neutral competition and origin-age ablation

Freeze before fresh quality outcomes. Preserve v106's full design: 12 cases,
two phase-disjoint rule pools, 32 trajectories per phase/case, each under
immediate and jitter0..31/missing0.2 schedules. There are 768 latent streams,
1,536 schedule-runs, 16 initial labels, 256 scored predictions, switches at128,
and as-of model publication every32 frames from long64/short32 available labels.
The six-coordinate adaptive observation policy and nine-coordinate training
audits remain unchanged. All models see identical arrived training evidence.

Fresh seed base2120110800 + phase*1000000 + case*10000 + index*10, roles0..4
for rules, inputs, outcomes, delays, missingness. Effective allocations must
avoid earlier v90-v106 learner/null allocations. No interim quality look,
replacement seeds, parameter sweep, optional sample enlargement or selective
case removal. V107 reused old data and introduced no allocation.

## Arms and isolation

0 generic64; 1 conservative journal; 2 origin-prefix role carry; 3 arrival-time
role carry. All four controls must match frozen v106 on consumed compatibility
trajectories before generation.

4 neutral-only: explicit fifth .5 forecast with prior .05, raw priors scaled
by .95, eta=.5 and .001 per-arrived-label fixed share.
5 age-only: original raw prior, no direct neutral candidate, origin-discounted
losses with32-frame half-life and eta=.5; no additional fixed share.
6 combined: the neutral prior of4 and discounted selector of5.

The [component contract](../../research/aged-advice-component-contract.md)
defines exact updates. Age-only is a replacement of fixed share by discounted
loss accumulation, not an isolated change to the exponent of one new loss.
All three retain the same evidence gate, thresholds, expiry and version scopes.
New direct neutral mass never restores rejected raw models. The composed law
controls acquisition and the issued forecast. No oracle change detector or
oracle mixture enters any arm. The invariant raw-role identities, not fitted
model identities, support cross-publication selector losses.

At clock t advance advice time, deliver earlier due outcomes, expire unresolved
age>=32 records, publish if due, then acquire and forecast all seven arms.
Sample the current outcome afterward; deliver zero-delay labels only then.
Flush clocks256..287. Model training lists remain ordered by origin and
restricted to arrived evidence. No missing outcome is labeled negative.

## Frozen gates

Report every candidate separately. Paired mean +/-3.5 standard errors over32
trajectories is a fixed-sample approximate screen, not a confidence sequence,
exact coverage theorem or history-wide multiple-testing guarantee. Do not treat
the two schedules of one latent stream as independent samples.

Each candidate has418 gates:

- 384 non-harm: Brier harm upper<=.01 against each of controls0..3, every phase,
  case, schedule and full/late segment.
- 20 generic gains: full parity3/4/complement4 and late two switches, both
  phases and schedules; mean>=.005 and lower>0.
- 6 conservative gains: delayed late parity4 and two switches, both phases;
  mean>=.005 and lower>0.
- 4 full-carry and4 arrival gains: delayed late two switches, both phases;
  mean>=.005 and lower>0.

Candidate PASS requires every one of its418 gates and integrity checks.
Study success requires at least one complete candidate pass, not collecting
passing cells from different candidates. Preserve all1,254 gates and all
previous failures. Descriptive diagnostics cannot change this decision.

## Integrity and performance

Require exact replay of all1,536 runs, source hashes, independent JS score,
as-of model-origin and journal-counter reconstruction, paired tapes, original
control compatibility, literal discounted batch tests, rejection-authority and
atomicity tests with race detector. Normalize combined neutral mass to avoid
an all-rejected 1+ULP roundoff error; the regression was found on consumed
compatibility data before quality generation.

Measure component and entire seven-arm fixture performance only after quality
and replay stop. Do not report a whole-fixture mean as query latency, and do
not promote research-only behavior into production. No production changes,
commits, pushes or new private data use are authorized by this experiment.
Real-agent, independent-generator, member-split and serving-tail criteria
remain separate requirements, even if a candidate passes this finite screen.
