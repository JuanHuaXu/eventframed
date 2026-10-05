# Returning-pattern V38 frozen protocol

2026-10-03. Freeze BEFORE outcome collection. Seeds2026103803 design and
2026103804 confirmation, distinct from all previous cohorts. No outcome,
source, prior, gate or parameter may be changed after collection begins.

## Population and Controls

Two cosine baseline geometries,12 regimes,16 independent worlds/cell:
384 worlds/split,768 total. Each world has150 members and16 genuine trials
per member,2400 labels. Pair four modes(full64, adaptive V37 bank, frozen
anchor, two-state orientation) under immediate and fixed150-tick feedback.
All modes see identical nomination/outcome volumes. There are8 arms/world;
copies are not independent trials. Frozen old collectors define as-of order:
issue before same-tick arrivals, ascending trial ties, drain after issuance.

First eight regimes retain V37 definitions on new seeds: aligned,
independent two-rate, curved, Beta baseline-matched, abrupt round8, late
round12, recurring round4/8/12, gradual rounds4..12. Four additional
two-rate regimes use a separately seeded permutation and fresh Bernoulli
draws: partial randomly selected75 members reverse at round8; unrelated
rates independently permute at round8; asynchronous each member reverses
at a independently drawn round4..12; early every member reverses at round2,
before the fixed anchor completes. Regime/true rates remain evaluator-only.
The generator's discrete times are controls, not discovered causal events.

All templates, original issue forecasts, private receipt forecasts, snapshot
laws/counts, readiness boundary and accounting times are serialized. The
anchor modes' Weights[0] snapshot field records next-state reversal mass;
other three entries are0, not predictive expert weights. ExpertIssued rows
are all0 for those modes. Full/adaptive retain V37's meanings.

## Metrics and Frozen Screen

Original issued expected Brier and first10 priority-weighted Brier use the
contemporaneous true rate. Round-end and final-drain whole/priority Brier,
top10 expected usefulness and packed bias use the current declared rate.
All are evaluated only AFTER emission; truth never influences training.

Per geometry/regime/schedule, all16 paired worlds contribute. Intervals are
mean +/-3.5SE across worlds, not confidence sequences or AP coverage.
Same frozen .01 improvement/protection tolerance as V37; no relaxed gates.

For orientation versus FULL control:

- stationary first four regimes: issued whole/priority and final whole/
  priority/usefulness paired gain lower>=-.01;
- shifted remaining eight: issued whole AND priority gain mean>=.01 and
  paired lower>0; final whole/priority/usefulness lower>=-.01;
- all regimes: issued whole/priority and final whole/priority/usefulness
  noninferiority versus ADAPTIVE control, gain lower>=-.01;
- discrete changed phases(abrupt/late/recurring/partial/unrelated/early):
  recover after TWO consecutive in-phase round ends with Brier<=.20 and
  usefulness>=.75. Miss penalty=phase length+1, average phases WITHIN world.
  Relative to full: mean recovery gain>=10% of control mean and lower>0.
  Gradual/asynchronous have no discrete global phase recovery criterion;
- every arm's summed setup+issue+resolve+snapshot phase<=400ms/2400 labels;
  orientation construction plus separate freeze-plan allocation<=8MiB;
  zero learner errors, correct identity/future-prefix/source/audit checks.

Overall adoption requires EVERY48 cells in BOTH splits to pass. Model and
width controls are never selected after outcomes. Endpoint improvement or
one successful reversal regime cannot satisfy the whole objective. A failed
CI protection test is inconclusive unless its interval actually shows harm.

## Reproduction and Auditing

Collector TestExperimentV38 uses explicit EVENTFRAME_ORIENTATION_V38_OUT
and _SPLIT, exclusive-create,0600, flush/fsync. The manifest freezes the18
old V37 sources plus5 new candidate/test/collector/preflight/protocol files.
Full independent arithmetic audits are written separately after collection,
hashed with benchmarks, and cannot change sealed data or the screen.

Audit exact generator replay; private original laws/receipt identities;
separate batch anchor integrals and matrix/path filtering; control-window
batch laws; priority/packet/risk/recovery calculations; source/shape/counts,
time accounting and corruption rejection. Keep design and confirmation
failures. No production, install, whitepaper, commit or push permission is
inferred. Keep all seven whole goals OPEN until their full evidence exists.
