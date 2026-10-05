# Noise-channel challenger V53: diagnostic failure

## Verdict

The forward measurement model is internally coherent and passes independent
math/replay audits, but does NOT rescue the challenger. Do not promote it or
open its reserved confirmation cohort. All seven research goals remain OPEN.

[Frozen protocol](mmm-noise-v53-protocol.md),
[raw outputs](../../research/noise-v53-diagnostic/diagnostic.jsonl),
[independent readback](../../research/noise-v53-diagnostic/readback.json),
[source freeze](../../research/noise-v53-diagnostic/freeze.json).

36 worlds:18 regimes x2 geometries,3 arrival schedules,7 matched arms.756
arms/86400 DISTINCT labels, not1814400 independent observations. All2400
nominations per arm retained. Diagnostic n1 per cell; no confidence interval,
confirmation, production adoption or whole-goal completion inferred. Base
2026105307 consumed; design2026105309/confirmation2026105311 UNEVALUATED.
Seed-separation tests generate seed metadata, not forecasts or evaluations on
those reserved cohorts. Correlated round flips deliberately violate independent
label-noise assumptions and remain in the report.

## Quality

Equal-cell means across all108 geometry/regime/schedule cells. Smaller Brier
loss is better; usefulness is the terminal top10's expected clean outcome rate,
not chatbot-answer accuracy. No ground-truth rate enters the learner.

| Arm | Issued Brier | Priority Brier | Terminal Brier | Top10 Usefulness |
| --- | ---: | ---: | ---: | ---: |
| Full | .226502 | .226073 | .209783 | .711747 |
| Adaptive | .216029 | .215658 | .180963 | .816639 |
| Original rich moment2 | .219490 | .218656 | .189521 | .811499 |
| eta0 | .219490 | .218656 | .189521 | .811499 |
| Fixed eta10 | .220875 | .220032 | .187792 | .807170 |
| Fixed eta20 | .223410 | .222606 | .190061 | .796149 |
| Learned eta | .220159 | .219360 | .188225 | .808836 |

eta0 is EXACT original non-cost law/receipt/snapshot/metric parity in ALL108
paired cells. Its faster construction is the already-existing memo constructor,
not a scientific noise-correction gain.

Fixed eta10 wins39/loses69 issued-risk cells against the original; eta20 wins38/
loses70; learned eta wins34/loses74. Adaptive harm exceeds.01 in18 original/
eta0 cells,25 eta10,36 eta20 and24 learned cells. On the original84 cells alone,
learned eta's mean loss increase is.001176924; the added stresses do not hide it.
All candidates miss the.01 Full-relative mean gain even before the missing
confidence requirement. No gate weakened.

Across54 cells with declared changes: eta10 improves recovery3/worsens7 against
original; eta20 improves0/worsens16; learned improves3/worsens0, a mean gain
of.055556 rounds. Learned wins one round in wide/partial_noise10/immediate and
tight/unrelated_noise10/{immediate,uniform299}. It STILL averages.703704 rounds
slower than Adaptive; mean per-cell relative recovery gain is NEGATIVE12.9644%.
The proposed >=10% Adaptive recovery gate is not met. Terminal utility improves
6/worsens6 learned cells; no broad downstream usefulness rescue.

## Model Identification Warning

Post-run inspection, not a new tuned protocol: learned eta favors about.10 on
CLEAN aligned cases and essentially.20 on CLEAN recurring cases. In recurring
cases terminal usefulness falls.06 against original in every geometry/delay
cell. For stationary IIDnoise10/20, mean issued-risk gains are.002339899/
.003818796, but partial shifts regress. Thus marginal-evidence confidence is
NOT evidence that actual corruption was identified. The declared latent-rate/
family/transition model can explain clean structural mismatch as noise.

This changes the next action: noise-rate model averaging alone is not a broad
rescue. A further noise-aware lead needs independently justified measurement
information, or a prospectively tested joint state/measurement identification
restriction, with acquisition costs and missing-source cases retained. Do not
feed true generator noise rates, case identities or future outcomes to a gate.

## Cost And Audit

Maximum complete2400-nomination/drain loop: Full65.972ms,Adaptive208.226ms,
original121.342ms,eta0 84.769ms,eta10 85.038ms,eta20 85.929ms,learned276.417ms.
All meet the unchanged400ms component cap. This is offline, serialized,
fixed-order measurement, NOT loaded serving/freshness or a100ms e2e guarantee.
Construction retains tickets; maximum allocated bytes: eta0 1789232,eta10
1789248,eta20 1789216,learned5561288. All within8MiB; not resident-memory bounds.
Observed-law/weight instrumentation is included in the candidate loop cost.

Six model test roots plus two fixture roots pass with the race detector.
Independent full-history reference,6 explicit21^3 latent-path cases,768 delayed
scalar resolutions, learned joint evidence and observed-law mixture checks,
ownership/clock/epoch/caps/cancellation and partial Issue/Resolve fencing pass.
12 nonvacuous future forks/84 semantic corruption controls pass. Full artifact
audit independently reconstructs ALL36 populations and all756 arms, including
clean and observed laws, weights, original receipts, snapshots and metrics.
11520 actual seeds across declared old/new namespaces are distinct.

Seven frozen runner commands terminate0; separate readback recomputes all losses,
priority, ordering, usefulness and recovery. Before collection,56 actual
repository compiler/test inputs plus8 supporting files were frozen/copied;
the calibration dependency omitted in V52 is included this time. Go1.27.1
darwin/arm64, compile/link binary hashes, command logs and source copies recorded.
The closure digest in freeze.json describes the compact Go JSON stream; the
saved pretty closure file has its own artifact digest in completed.json.

[Preflight repairs](../../research/noise-v53-preflight.md) preserve the audit's
initial accumulating-score mistake and the new wrapper's ticket-integrity
repair. No model/seed/gate retuning. All14 pre-existing tracked edits retain
their preceding checkpoint hashes. Production, private data, whitepaper and
existing confirmation outputs untouched; no commit/push/install/deploy.

Supplemental [closure readback](../../research/noise-v53-diagnostic/closure-readback.json)
also verifies the56 saved compiler inputs and compact JSON digest. Its first
inline version falsely included3nonrepository generated test mains; the runner
already filtered them. [Audit note](../../research/noise-v53-post-audit.md)
preserves this local repair. No prospective freeze or outcome record rewritten.

[Next identification lead](../../research/noise-v54-direction.md) has18parameter/
68stagewise likelihood identity checks, not a learned or acquisition experiment.
Conditional independent measurements of the SAME outcome can separate noise
and clean-rate ambiguity; source independence/extra cost remain requirements,
not a capability or completed goal established by V53.

## Research Source

Patrini et al., [CVPR2017 paper, section4.2](https://arxiv.org/pdf/1609.03683):
forward correction applies a declared noise transition to clean class
probabilities before comparison with noisy labels. Its theorem requires a
known nonsingular transition; its estimator has additional identification
assumptions. Our finite Bayesian adaptation and noise-prior model comparison
do not inherit its neural-training guarantee. Full PDF accessed via arXiv;
the CVF endpoint returned403 during this continuation.
