# V69 Class Observation Results

2026-10-04 local date. **Technical audits PASS; both new class-observation
policies FAIL the frozen quality, stationary-harm, recovery and loop-time gates.**
All seven whole research goals remain open. No production or whitepaper changes,
new confirmation outcomes, private data access, commit or push.

[Protocol](mmm-class-v69-protocol.md),
[full readback](../../research/class-v69-diagnostic/readback.json),
[terminal command records](../../research/class-v69-diagnostic/completed.json),
[phase costs](../../research/class-v69-diagnostic/cost-audit.json).

## Scope And Change

40 previously consumed worlds, two geometries and20 regimes, three delays,
ten paired arms/world-delay:1,200 arms,96,000 distinct underlying Y outcomes.
The repeated arms/delays are not1,200 independent worlds. One world per
geometry/regime is a diagnostic, not fresh confirmation or population coverage.
All paired policies request400 second measurements,25 per round. The no-pair,
Full and Adaptive controls use fewer observations and remain resource ablations.

New packages preserve V68's model priors, likelihood, issue clocks, horizon,
forecast law and lifecycle. Only nomination changes: model_class scores the
three model alternatives; noise_class scores the nine(model,noise) classes.
In the local alternative its noise variable belongs to the queried member,
not a single common global target. Rate/field atoms are collapsed before Gini
concentration scoring. These structural classes are not verified truth,
operational decision classes or Anti-Pigeon authority.

The mathematical refinement counterexample survives runtime integration, but
does not prove that rate/field uncertainty is irrelevant to useful prediction.
This score is not EC2 and inherits no competitiveness guarantee from
[Golovin, Krause and Ray](https://arxiv.org/pdf/1010.3091).

## Scientific Results

Lower issued Brier is better. Utility is expected usefulness of the final
top10 packet under the synthetic truth, not actual agent-answer accuracy.

| Mode | Issued Brier | Final Brier | Final utility | Worst core ms | Cells over400ms /120 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Full | .224699096 | .208151654 | .727150389 | 98.010 | 0 |
| Adaptive | .214747855 | .181598705 | .815037852 | 239.628 | 0 |
| Hybrid no pair | .251807215 | .231895100 | .622761882 | 342.325 | 0 |
| Hybrid random | .250143838 | .232598363 | .625639871 | 429.961 | 47 |
| Hybrid uncertainty | .250088691 | .232033075 | .623639871 | 716.883 | 92 |
| Hybrid information | .249979403 | .232068987 | .624639871 | 752.908 | 94 |
| Hybrid latent falsification | .249851398 | .231995998 | .624639871 | 720.461 | 98 |
| Hybrid predictive value | .250044799 | .232479486 | .624139871 | 2061.143 | 120 |
| Hybrid model class | .249960460 | .232160063 | .626639871 | 729.209 | 102 |
| Hybrid noise class | .249754661 | .231930812 | .626639871 | 720.720 | 95 |

Model class is worse than latent falsification by.000109062 issued Brier.
Noise class improves it by.000096738, and improves random/uncertainty by
.000389177/.000334030. Versus random this is about.156% relative Brier
reduction, not a corresponding percentage-point accuracy gain.

Both new policies still lose to Full by about.025 and Adaptive by about.035.
Each has110/120 cells exceeding.01 Adaptive harm, and averages2.35 rounds
slower restricted recovery in shifted cells. Their stationary harm is
.037959598/.037653913; worst stationary harm.103595522/.099601081.
No failed arm, old control or gate was removed or relaxed. All eight new-model
policies fail the quality/harm/recovery conjunction. Complete readback retains
every cell and regime, not just favorable noisy cases.

| Candidate vs control | Issued-risk gain | Total core cost ratio |
| --- | ---: | ---: |
| Model class vs random | .000183377 | 1.584761 |
| Model class vs uncertainty | .000128231 | 1.007632 |
| Model class vs latent | -.000109062 | 1.002266 |
| Noise class vs random | .000389177 | 1.580934 |
| Noise class vs uncertainty | .000334030 | 1.005198 |
| Noise class vs latent | .000096738 | .999845 |

These are the contemporaneous serial core measurements. Equal400 requests are
NOT equal TOTAL cost. Tiny internal gains at58% more core time than random do
not establish Goal7, and no fresh confirmation or equal-cost win is claimed.
The field named v68LatentGain in readback compares the two NEW modes with the
old latent policy; for old modes it checks their own V68 counterpart instead.

## Runtime Components

Apple M4, Go1.27.1, darwin/arm64, three serial100ms benchmark repetitions.
Single Predict in a150-member model:16.017-16.242us, zero allocations.
Single model/noise-class query:83.699-84.401/83.811-84.800us, zero allocations.
Single all-target predictive-value query:563.391-566.912us, zero allocations.
These are repeated public API calls on a prepopulated900-frame model, not
150-query batches, a loaded daemon, persistence or full request tails.

Separate constructor allocation check:1,696,320bytes at150 members and
2,263,304 at200, below8,388,608 in both. This is measured allocated memory,
not RSS. Construction costs about1.632-1.641/2.182-2.189ms respectively.
Benchmark B/op and the separate TotalAlloc test have distinct measurement
boundaries; neither closes Goal6.

Postcollection phase accounting covers all960 model arms and verifies every
phase sum. Noise-class nomination is36.91% of core elapsed, first-observation
replay26.59%, issue14.51%. Predictive-value nomination is79.20%. These are
recorded wall-time fractions, not a CPU profile or unique causal diagnosis.
The faster class-score arithmetic has not made the complete loop meet400ms.

Collection655.925s; independent replay1065.741s; final readback9.712s.
Those offline experiment/audit durations are not serving latency. No other
timed benchmark ran during collection. Small finite-prior checks ran during
the offline replay, not the timed learner experiment.

## Audit And Future-Data Boundary

Twelve functional race roots pass; allocation is explicitly skipped in that
race run then separately executed. Vet and three fixture future forks pass.
29 fixture corruption controls reject; two specifically alter class values.
New tests compare24 actual posterior class branches, verify latent-refinement/
nuisance/boundary cases and preserve original-state nonpublication.
30,636 delayed forecast/weight/acquisition comparisons pass, alongside14,896
checkpoint,1,486 legacy,48 mode-specific and160 static-enumeration comparisons.

ALL960 model arms independently replayed:2,304,000 issued packets with clean
AND measured W1 laws,4,608,000 scalar comparisons, plus selected values,
choices, requests, receipts, missingness, snapshots, scores and costs.
All960 original-mode controls are bitwise unchanged excluding timing versus
V68; all40 underlying populations are identical. The class score cannot read
unarrived W2 or synthetic true probabilities; those are evaluator-only.

116 compiler inputs,14 support files and two generated mains were frozen and
copied. All eight required command records are terminal0; candidate sources,
referenced V68 data and14 protected tracked files remained unchanged.
Raw data SHA256:
`8732574de45d442ec007fd93211a4f20c011bd7a606c40bbc3b9aa734e6462f5`.

## Oracle Feasibility

The separate [postcollection oracle audit](../../research/class-v69-diagnostic/recovery-headroom.json)
checks expected-risk floors and maximum top10 utility at every round. All24
declared change phases across20 shifted worlds can satisfy the two-consecutive-
round recovery rule under a perfect forecast. None is impossible on this
realized cohort. Therefore an unattainable recovery threshold does not explain
the failure here; the original failed verdict remains intact.

The perfect-forecast issued Brier floor averages.160094456, leaving a loose
.064604641 gain ceiling against Full. Oracle recovery leaves4.416667 rounds
of average possible improvement over Adaptive in shifted cells. These are
truth-only evaluator upper bounds, not deployable policies, available sample-
budget gains or external-population impossibility claims. No learner used them.

## Next Boundary

Retain the small noise-class component gain, but do not repeat observation-
score changes as though they have fixed the larger learner gap. The
[V70 prior preflight](mmm-rate-prior-v70-results.md) verifies that unconditional
widening raises stationary excess risk. The next
[dynamic dispersion direction](../../research/dynamic-dispersion-v71-direction.md)
revisits V34/V35's useful, still-incomplete hierarchy components with explicit
joint dynamic/noise and original-position evidence contracts.

Goal3 still requires useful valid splits, Goal5 untouched outcome-labeled agent
tasks, and Goal6 loaded durable freshness including visible mutation/recovery.
The score, oracle and memory components are not substitutes for those goals.
All seven whole goals remain OPEN/ACTIVE; all negative evidence is retained.
