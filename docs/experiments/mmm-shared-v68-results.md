# V68 Shared-Context Results

**Implementation/inference/audit and allocation PASS; scientific rescue FAIL.**
Sharing improves on the V66 local challenger but does not beat Full or Adaptive.
Several core-loop gates fail. No adoption, confirmation, publication or production
change. All seven WHOLE goals remain OPEN and the thread goal stays ACTIVE.

[Protocol](mmm-shared-v68-protocol.md),
[frozen inputs](../../research/shared-v68-diagnostic/freeze.json),
[raw data](../../research/shared-v68-diagnostic/diagnostic.jsonl),
[readback](../../research/shared-v68-diagnostic/readback.json),
[cost accounting](../../research/shared-v68-diagnostic/cost-audit.json),
[terminal commands](../../research/shared-v68-diagnostic/completed.json).

## Experiment

All 40 CONSUMED paired worlds: two generators, 20 regimes, three delay schedules,
14 arms per cell, 1,680 arms total. There are 96,000 underlying Y outcomes, not
4,032,000 independent arm outcomes. Exactly400 W2 requests in each paired arm,
including missing evidence. This is one trajectory per generator/regime/delay,
not untouched confirmation or population-confidence evidence. No sealed agent
outcomes or reserved future seeds opened.

Four coherent model modes separate local learning, shared noise, shared context
and their evidence-weighted mixture. All issue laws and acquisition values are
from the same joint model. W2 measures the original Y and replaces its likelihood
at the original position; shared predictive value covers ALL affected targets.
The 48-state field is a declared bounded hypothesis, not a Gaussian process,
external-truth model or Anti-Pigeon sharing authority.

## Quality And Cost

Lower Brier is better. Utility is expected usefulness of the final top10, not
observed agent answer accuracy. Maximum core time covers the complete2,400-frame
lifecycle, NOT one served request or loaded durable latency.

| Mode | Issued Brier | Terminal Brier | Top10 utility | Maximum core ms | Cells over400ms /120 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Full | .224699096 | .208151654 | .727150389 | 63.524 | 0 |
| Adaptive | .214747855 | .181598705 | .815037852 | 223.710 | 0 |
| Hybrid no pair | .251807215 | .231895100 | .622761882 | 323.777 | 0 |
| Hybrid random | .250143838 | .232598363 | .625639871 | 431.770 | 48 |
| Hybrid uncertainty | .250088691 | .232033075 | .623639871 | 713.049 | 91 |
| Hybrid information | .249979403 | .232068987 | .624639871 | 724.235 | 91 |
| Hybrid falsification | .249851398 | .231995998 | .624639871 | 717.886 | 100 |
| Hybrid predictive value | .250044799 | .232479486 | .624139871 | 2038.630 | 120 |
| Local no pair | .269705599 | .217094257 | .787380705 | 192.836 | 0 |
| Local uncertainty | .264831503 | .208304871 | .792780334 | 247.848 | 0 |
| Shared noise no pair | .291882828 | .234268194 | .755230148 | 180.142 | 0 |
| Shared noise uncertainty | .277368100 | .211769474 | .782570774 | 255.076 | 0 |
| Shared context no pair | .255239658 | .251537403 | .536863671 | 152.763 | 0 |
| Shared context uncertainty | .254365038 | .253902276 | .532863671 | 488.039 | 40 |

Hybrid uncertainty improves issued Brier .014742812 over V66 uncertainty, but is
.025389595 worse than Full and+.035340836 worse than Adaptive. Its terminal risk
and top10 utility also worsen versus V66. All six hybrid policies have110/120
cells above the .01 Adaptive-harm limit. Hybrid paired recovery is2.35 rounds
slower than Adaptive on the unchanged recovery definition. Every new mode fails
all three scientific quality/harm/recovery gates; none is a complete rescue.

Pure shared-noise uncertainty is+.012536597 worse than local uncertainty. Shared
context alone improves issued risk relative to local, but final utility falls
to.532863671. This rejects simple noise pooling as sufficient and shows a
quality tradeoff for contextual sharing. It does not uniquely identify the
remaining model/prior/clock or observation-objective failure.

Hybrid falsification gains only.000292440 over random and.000237293 over
uncertainty. Its total core cost is1.579285x random and1.003652x uncertainty.
Predictive value gains.000099039/.000043892, costing4.749947x/3.018642x.
These are consumed-cohort means, not confidence-proven superiority. Equal400
requests are NOT equal TOTAL cost, so goal7 remains open.

Measured phase accounting identifies where time went, not a unique causal
profile: nomination consumes37.11% of falsification core time and79.19% of
predictive core time. First receipt replay consumes26.60% and8.84% respectively.
Hybrid no-pair first replay is54.38%; its issue and snapshot phases together are
44.89%. Repeated evidence-weight calculations and suffix work are legitimate
optimization leads; neither may change scored laws, delayed evidence semantics
or certificates merely to pass a timing gate.

## Component Validation

Ten functional race roots pass; vet passes. Independent component coverage
includes29,628 delayed forecast/model-weight/acquisition comparisons,14,896
cross-checkpoint comparisons,1,486 legacy V66 public-law comparisons,160 static
enumeration comparisons (5,856 latent states per case), and48 reference-mode
comparisons. The zero-noise exclusion, full journal, atomic failure, owner/epoch,
missing second evidence, future fork and nonlocal influence guards pass.

Nine corrupted artifact fields are rejected. A future W2 fork changes predictions
only after evidence becomes visible, and is nonvacuous thereafter. All1,440
new-model arms are independently replayed:3,456,000 issued packets with BOTH
clean and W1 laws (6,912,000 scalar comparisons), observation values/choices,
receipt metadata, requests, missingness, snapshots, metrics and cost sums.
All240 Full/Adaptive controls are bitwise identical excluding cost to V66 AND
V60; all40 populations are identical. Stored data alone is not evidence of model
visibility: lifecycle replay checks the actual forecast boundary.

All eight frozen commands terminate0. Collection667.74s; independent replay
1057.11s. Those are multi-arm experiment durations, not serving latency.
Compiler closure/generated test mains and support files are copied and hashed,
with14 protected dirty files unchanged. No concurrent timed experiment added.

Constructor allocation:1,696,320bytes at150 and2,263,304 at200, below8,388,608
at BOTH sizes; allocation is not RSS. Three serial benchmark repetitions on
AppleM4: prediction15.560-15.776us, predictive query552.861-558.720us, bothzero
heap allocations per operation in the benchmark. This does not transfer to
loaded RPC, persistence, freshness or unlimited streams; cap64/member remains.

## Preserved Failure And Next Lead

[Original numerical failure](../../research/shared-v68-boundary-preflight/command.json)
and its original source/log are preserved. After9,600 agreeing frames, a
supported noise hypothesis underflowed in linear joint mass; an opposing W2
then produced zero field support. The repair keeps conditional field laws and
log noise evidence separate. Static fields rebuild log masses from the journal.
It changes numerical representation, not the model, prior, caps or gates.
The repaired long-journal regression passes under race.

[Class-level preflight](../../research/class-v69-direction.md):1,289 algebraic
checks, maximum defect5.41234e-16,256 declared joint cases, independent branch
enumeration and refinement/boundary guards. Its counterexample reverses
latent-Gini observation ordering by duplicating indistinguishable atoms without
changing the class/outcome joint. Class-level concentration stays unchanged.
That identifies representation sensitivity, not the cause of all V68 failure.
No streaming, runtime, cost-matched, external-authority or quality result follows.

Next test a declared useful class/decision objective with nuisance controls,
using the same joint posterior and charging all nomination costs. Do not import
EC2 competitiveness into this approximate concentration objective. Goal3 still
requires valid useful splits; goal5 untouched outcome-labeled agent utility;
goal6 loaded durable freshness. All earlier negative evidence is retained.
