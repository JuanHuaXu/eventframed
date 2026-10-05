# V72 Equivalent Cache Results

**PASS exact-law and runtime components; FAIL scientific quality rescue.**
All seven WHOLE research goals remain open. No production, private/sealed-data,
paper, publication, install, commit or push change.
[Frozen protocol](mmm-dynvarcache-v72-protocol.md).

## Complete Screen, No Removed Failures

All 18 V71 arms across 40 consumed worlds, 20 regimes, two geometries and three
delay schedules: 2,160 arms, including 1,920 model arms. One trajectory per cell
is a development screen, not untouched confirmation. Each arm issues 2,400
first observations; paired policies request 400 second measurements of those
original outcomes. No generator, delay, failed arm or scientific gate changes.

| Arm | Issued Brier | Mean core ms | Worst core ms | Over-400 ms cells |
| --- | ---: | ---: | ---: | ---: |
| Full | .224699096 | 50.307 | 68.356 | 0 |
| Adaptive | .214747855 | 201.944 | 211.952 | 0 |
| baseline_no_pair | .348212362 | 31.191 | 33.986 | 0 |
| current_no_pair | .291882828 | 30.703 | 42.158 | 0 |
| current_uncertainty | .277368100 | 56.416 | 61.082 | 0 |
| free_no_pair | .271358559 | 32.429 | 33.437 | 0 |
| individual_no_pair | .262436406 | 29.588 | 30.583 | 0 |
| individual_uncertainty | .258671153 | 55.059 | 60.465 | 0 |
| local_no_pair | .248644065 | 31.468 | 32.614 | 0 |
| local_uncertainty | .245903739 | 57.210 | 61.836 | 0 |
| learn_no_pair | .271856086 | 31.813 | 33.025 | 0 |
| learn_random | .261542227 | 38.877 | 44.921 | 0 |
| learn_uncertainty | .256949823 | 57.698 | 63.889 | 0 |
| learn_information | .255968083 | 59.452 | 66.364 | 0 |
| learn_falsification | .255240816 | 59.504 | 72.235 | 0 |
| learn_predictive | .257241842 | 280.555 | 327.090 | 0 |
| learn_model_class | .261468372 | 57.827 | 62.435 | 0 |
| learn_noise_class | .258068777 | 57.777 | 62.444 | 0 |

All 16 model arms still FAIL mean-gain, per-cell Adaptive-harm and recovery
gates. The best new quality arm, local_uncertainty, is still .031155884 worse
than Adaptive on issued Brier, with 111/120 harm cells over .01 and 1.75 rounds
slower recovery. Its final Brier .199800776 and top-10 usefulness .807328674 are
unchanged. All earlier quality failures are preserved, not reinterpreted.

Every arm now passes the original <= 400 ms complete-core gate. Across matched
serial runs, local_uncertainty mean core drops 851.827 -> 57.210 ms (14.889x),
with worst core 907.963 -> 61.836 ms. All-target predictive mean drops
797.721 -> 280.555 ms, worst 871.549 -> 327.090 ms. Fully independent uncertainty
drops 450.448 -> 55.059 ms. These are entire synthetic core loops, not per-agent
request latency, loaded serving or durable freshness. Different serial runs
are not a randomized profiling attribution, even with exactly matched outputs.

## Why The Law Remains The Same

Cache normalized next-state vectors, member noise evidence/odds and global
family/noise odds at source-belief publication. Preserve the original arithmetic
order. Publication rebuilds global sums in member order with explicit zero
counts, without subtracting infinite evidence. Issue refreshes rate moments
even without an observed outcome; resolve prepares the factor, belief, cache
and global odds before one serialized publication. Hypothetical queries publish
nothing; cancellation supplies no likelihood; epochs reset derived objects.

The numeric log guard is intentionally narrower than a private-memory-tamper
certificate. Serialized ownership remains required. Race tests do not establish
concurrent-owner safety or a daemon integration guarantee.

## Audit Trail

- Fresh bitwise uncached-V71 audit: 120,204 counted scalar comparisons in 36
  full-journal configurations, all three noise/sharing modes, four family
  choices, hazards 0/1/16/1, all query modes, late pairs, cancellation and epochs.
- Fresh independent dense-reference audit: 108 configurations / 83,808 counted
  scalar checks; joint enumeration / 1,856 checks; priors/emissions / 1,911.
- Fresh external 64-trial old-query audit: 36 configurations / 42,840 checks.
  Lifecycle, fault atomicity, support and future-evidence tests pass under race.
- Three fixture future forks remain nonvacuous; 29 corruptions are rejected.
  Race, vet, allocation, benchmark, experiment, audit and readback terminate 0.
- ALL 2,160 arms match V71 bitwise outside ONLY `Costs` and `Breakdown`: clean
  and noisy issued laws, values, choices, requests, receipts, snapshots and
  metrics. All 40 populations match. This includes all 240 Full/Adaptive arms.
- That exact equivalence is linked to the immutable, hash-verified V71 full
  independent replay of all 1,920 model arms: 4,608,000 issued packets and
  9,216,000 scalar clean/noisy comparisons. This is transitive verification,
  NOT a new full independent replay. The small/full-journal audits above ran
  freshly. Prior sources, completion, commands/logs, raw data and checkpoint
  copies were verified, not assumed from prose.

Microbenchmarks at 150 members / 16 trials on Apple M4: prediction
957.1-959.1 ns, class queries 20.147-20.381 us, all-target predictive query
128.596-138.740 us. All zero allocation/op. Constructor allocation
1,771,200 / 2,353,344 bytes at 150/200, below 8 MiB, but not RSS.
Measured collection 150.53 s is job wall time, not serving latency.

Structured artifacts are in `research/dynvarcache-v72-diagnostic/`, including
raw data, compiler closure, generated mains, source freeze, command/log hashes,
transitive audit, readback and cost audit. Original and log-pointer-optimized
preflights remain immutable in `research/dynvarcache-v72-{initial,optimized}/`.

## Equal Cost And Next Quality Lead

Falsification's Brier gain remains .006301411 versus random and .001709007
versus uncertainty. Measured whole-core costs are now 1.530545774x and
1.031297592x respectively. Equal 400 requests still does NOT satisfy equal
TOTAL cost. Goal 7 remains open; a larger speedup for the random control makes
the relative cost premium bigger, not evidence against exact-law preservation.

[V73 mean/dispersion preflight](mmm-mean-joint-v73-preflight-results.md) retains
the full 27-mean family, verifies its coherent joint law and measures a packed
storage layout. It has no fitted-stream quality result yet. Mean mismatch,
noise, sparse evidence and static hyperparameters remain competing explanations.
Valid useful Anti-Pigeon splits, untouched-agent utility and durable loaded
freshness retain their original requirements. No completed whole goal, adoption
or exhaustion is declared.
