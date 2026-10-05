# Frozen transfer to landing events

## Result: bounded transfer screen passes, interpretation remains incomplete

The frozen task-role-v1 candidate improved top1 support from4/10 to6/10 on the
ten positive landing-family questions, with two rescues and no lost correct
ordinary-retrieval case. All24 calls completed. Both absent-information controls
still packed a candidate. No rule, vocabulary or threshold changed during this
transfer run, and the prior design artifact hashes still verify.

The two rescues are positive date selection after2020 and explicitly negated
selection before2020. The candidate did NOT learn the landing relation. Its
TargetRelation is empty on all landing questions. The success establishes some
transfer of the selection logic, not generalization of event-relation grounding.

## Remaining failures

| Case | Selected evidence | Required evidence | Diagnosis |
| --- | --- | --- | --- |
| verify-perseverance | Perseverance launch | Perseverance landing | Right entity, wrong relation |
| refute-curiosity | Curiosity launch | Curiosity landing | Right entity, wrong relation |
| refute-perseverance | Perseverance launch | Perseverance landing | Right entity, wrong relation |
| wasnt-after | Perseverance landing | Curiosity landing | was-not contraction unrecognized |

NegatedSelection is false for wasnt-after. This distinguishes that failure from
date arithmetic. Assessment correctly avoids treating dates as selection
conditions, but without relation grounding it falls back to the earlier ranking.
The missing-cost and return-to-Earth controls still select landing records;
retrieving a related event is not evidence that the requested attribute exists.

Full-frontier forecast laws and numeric ranker inputs are identical across arms;
journal explanations match packets. This is still ordering research, not a
posterior-learning result, generated-answer evaluation, or abstention rescue.

## Sources and verification

The four facts were checked against primary NASA sources before dispatch:
[Curiosity mission overview](https://ares.jsc.nasa.gov/missions/msl/) and
[Perseverance landing release](https://www.nasa.gov/news-release/touchdown-nasas-mars-perseverance-rover-safely-lands-on-red-planet/).
Curiosity uses its2012-08-06 UTC landing date. Dates are historical public facts;
the questions are researcher-designed controls, not sampled production queries.
Two missions and ten dependent questions do not support a population accuracy
claim. This new fixture is now consumed for future design purposes.

Protocol: LANDING_TRANSFER_PROTOCOL.md. Artifacts in landing-transfer-v1 include
corpus, queries, oracle and results JSON. The runner reads oracle only after
all predictions. The verifier checks hashes, matched numeric inputs, laws,
journals, support labels and the predeclared improvement/no-regression screen.

```sh
node research/public-task-pilot/check-landing-transfer.mjs
```

## Next lead

Do not merely append landing to a growing hand-written dictionary and claim
domain translation. Compare relation-aware retrieval using the existing what
field and a declared relation representation against lexical and current
whole-frame controls. It must distinguish entity, event relation, requested
attribute and assertion scope, with calibration/unknown handling tested on
subsequently new relations. Morphological negation normalization is separately
testable but insufficient to repair the three relation failures or absent data.

The current candidate remains frozen and unpromoted. All seven whole research
directions remain open; production and whitepaper are unchanged.
