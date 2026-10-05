# Native-Cue Ablation

2026-10-04. Frozen before execution. Pairrank and Lambda rescue both failed
cross-fit improvement over BM25. Native nomination/rank coefficients were large,
but that observation alone does not identify the cause of harm. Alternatives
include weak raw overlap features and training/metric mismatch. This experiment
tests one controlled mechanism; no production/native change is authorized.

Two new isolated commands: source-only all-pair learner and source-only Lambda
learner. Their scientific inputs are byte-identical to the respective failed
studies:5183actual stored sources,351previously usedFIT queries, common200frontier,
5whole-family folds, labels, baseBM25, .25tanh,200steps/.2rate/.001ridge/4cap.
Only mask feature indices6and7 to0 AFTER source feature extraction, before both
training and scoring. Preserve all other features/base scores/identities and
ALL200rows. Validation precedes copying/masking; malformed input never becomes
valid merely because it was in a masked channel. Original models stay frozen.

Predictors still receive native ranks to reconstruct the original observation
contract, but masked vectors cannot convey them to the fitter/scorer. No model
training/feature selection on calibration or confirmation. These retrospective
FIT cross-fold comparisons remain exploratory, not untouched validation.

Compare masked models to BOTH BM25 and their original unmasked learners under
query-macro and221source-family means. Primary scientific screen: joint recall10
and NDCG improvement over the stronger common-frontierBM25, without silently
switching to a favorable aggregation. Preserve losses and every no-pair query.
Effect attribution is limited to these frozen models/data, not a causal claim
about real-world source validity or a claim native retrieval is generally bad.

Race3times including original learner roots and mask controls. Test all8feature
values, non-mutating full-frontier copy, invariance to arbitrary VALID native
cues, validation of invalid masked cues, caps/cancellation. Independent JS
reconstructs raw features, verifies EXACT masking, independently refits all6
models per arm and checks351complete rankings before outcome summaries. Cost
includes masking/validation, index/features/training/ranking/processRSS.
Native/source bytes, original experiment costs and failures remain preserved.

Falsifier: removal fails to eliminate harm, or any other channel/frontier/epoch
changes. Confidence/law untouched; source-feature snapshot is replayed actual
stored evidence, NOT live mutation/invalidation or an OpenClaw output test.
Methodological inspiration: [Joachims2002](https://www.cs.cornell.edu/people/tj/publications/joachims_02c.pdf)
and [Burges2010](https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/MSR-TR-2010-82.pdf).
Our exact feature ablation is an implementation experiment, not their theorem.
All seven WHOLE goals OPEN/ACTIVE; no production/private/whitepaper edits.
