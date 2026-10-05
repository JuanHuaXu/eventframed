# Continuing Local Rates V39 Results

2026-10-03. Technical audit PASS; quality adoption FAIL for every candidate
on BOTH splits. No production/whitepaper/runtime deployment changes.

## Coverage And Frozen Decision

Two geometries,14regimes,three evidence-delay schedules,16worlds/cell,
seven models:448worlds/9408arms/156800snapshots/1075200distinct outcome
trials per split. Total896worlds/18816arms/313600snapshots/2150400trials;
copies across arms and within trajectories are NOT independent observations.
Two additional regimes flip10% feedback labels independently of true utility.
Original issued probabilities are scored, never hindsight-refitted probabilities.
No hazard/model/gate change between design and untouched confirmation.

| Candidate hazard | Design failed cells /84 | Confirmation failed cells /84 |
| --- | ---: | ---: |
| 0 | 78 | 78 |
| 1/32 | 78 | 78 |
| 1/16 | 78 | 78 |

Only six baseline-matched cells pass for each candidate. All computation and
allocation checks pass; quality/recovery, not runtime cost, reject adoption.
Mean +/-3.5SE over16paired worlds is the declared exploratory screen, NOT a
simultaneous confidence sequence, calibrated target-law certificate or general
valid Anti-Pigeon error control. See frozen protocol for exact multi-metric gates.

## What The Negative Result Shows

Confirmation wide/partial/immediate means:

| Model | Issued expected Brier | Final Brier | Recovery | Top10 utility |
| --- | ---: | ---: | ---: | ---: |
| Full | .228704 | .214287 | 9 | .800000 |
| Adaptive | .205726 | .171452 | 4.875 | .800000 |
| Global orientation | .247627 | .284528 | 9 | .526250 |
| Local0 | .240142 | .215485 | 9 | .800000 |
| Local1/32 | .232108 | .194398 | 8.250 | .800000 |
| Local1/16 | .231276 | .194619 | 8.0625 | .800000 |

Recovery9 here includes the frozen phase-length+1 miss penalty. The local
reset models improve final Brier over Full/global orientation, but lose on
whole-trajectory issued Brier and recover more slowly than Adaptive. Improved
endpoints do not rescue the requested learning objective. On abrupt confirmation,
Local1/16 issued Brier .255144 is worse than Full .250651 and Adaptive .214856.

The exact21-state independent-member filter removes forced shared/global
orientation, but also removes compatible-member evidence borrowing. The
prior is a normalized discrete Beta-shaped mass centered on supplied baseline
geometry, not a calibrated truth probability or exact continuous Beta posterior.
Bad prior calibration, sparse16-label histories, lost borrowing and reset
misspecification are credible explanations; this run does NOT isolate their
causal contributions. Next use a NEW frozen factorial with uncertain/calibrated
priors and properly validated hierarchical borrowing, not tune consumed seeds.

## Correctness And Performance Boundaries

Independent21^3path enumeration, hazard0 batch likelihood, all-member batch
reference, delayed-arrival permutation, future-prefix/private issued-score,
owner/epoch/replay/cancel/cap controls and full race/vet checks pass. Thirteen
non-identity technical corruption/allocation controls pass. Independent
full/window/anchor/orientation control replays retain existing definitions.
Likelihood and next prediction belong to one declared model family; unit
missing emissions assume non-informative delay/cancellation. Fixed/uniform
delays satisfy this experiment, not arbitrary selective evidence in production.
Single-owner code; race PASS does not authorize shared-goroutine ownership.

Three constructor benchmarks:228606-231719ns,2401975-2401981B,14allocations.
Predict150:2553-2577ns,zero allocations. LateReplay64:776.4-777.3ns,zero
allocations, but ONLY the first of64positions has a label; the other63have
unit emissions. This is NOT the fully observed worst-compute suffix benchmark.
Complete collector elapsed maxima per2400labels(all under400ms frozen cap):

| Model | Design ms | Confirmation ms |
| --- | ---: | ---: |
| Full | 65.305 | 61.335 |
| Adaptive | 222.682 | 226.816 |
| Local0 | 1.970 | 1.931 |
| Local1/32 | 1.825 | 1.893 |
| Local1/16 | 1.732 | 1.898 |

Collector elapsed includes nomination/issue/arrival/drain computations but
excludes serialization, independent auditing, acquisition, serving, storage
and embedding. It cannot prove whole Goal6 latency or Goal7 equal TOTAL cost.
Auditor footprint/sample is not learner allocation; offline audit took
1005.21s design/1013.36s confirmation, distinct from these model timings.

## Preserved Failures And Audit Trail

Original preflight fails because its schedule-corruption fixture assigns the
same immediate schedule (a no-op). Archive original auditor and failing log;
before any outcomes, change ONLY that fixture to fixed150 and require every
mutation non-identity. Repaired full race PASS231.599s and vet/benchmarks PASS.
Design collector PASS393.01s; original12minute offline audit times out, with
no recorded arithmetic mismatch. Its `design-audit.json` is command metadata,
not a statistical report. Original runner also had a command/report filename
collision risk; preserve it and separate names in the continuation wrapper.

Separate wrapper freezes45sources, reuses SAME sealed design raw, extends
only offline verification timeout to35minutes, leaves400ms learner gate
unchanged, and collects untouched confirmation once. Both full audits PASS.
All original failures/logs/sources remain; no outcome resampling or adoption
gate relaxation. Raw SHA256:

```
design       2a50abb66cc7cad1f84308c5763df938904303a4b917e38304a12fbe9835ddb1
confirmation ecdf8faaf835a9ae20d9c13ea12e0fbcea939707fc18b73052d7bc02028eddf4
```

Artifacts: `research/continuing-v39-completion/{freeze,completed}.json`, both
statistical `*-audit.json` and separate `*-audit-command.{json,log}`. Prior
attempts under `research/continuing-v39` and `research/continuing-v39-recovery`.
All seven whole goals remain OPEN; this is a rejected research candidate.
