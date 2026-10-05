# Temporal contrast diagnostic

Research-only; frozen before this run. This is a diagnostic of existing retrieval,
not a new temporal model, independent domain confirmation, or agent answer test.
The eight public NASA/IAU facts in context-v1 are already consumed. No new facts,
private data, completion model, or production service are used.

Hypothesis: temporal reference selection contributes to the observed Ceres error.
Alternatives include weak subject matching, generic semantic ranking, and missing
temporal information in compressed features. This run can demonstrate relation
insensitivity but cannot identify one unique internal cause.

Five paired questions target different retained records: before/after an anchor,
initial/subsequent classification, and proposed/adopted status. Two controls have
no supporting record. A correct pair requires BOTH top-ranked records correct;
merely changing the top record is insufficient. Report top1, support rank,
pair accuracy, unchanged-top pairs, and absence behavior separately. No threshold
tuning or query reformulation after observing results. All ten positive questions
must pass for this small diagnostic to clear temporal adequacy; that is not the
roadmap's broader success criterion. Absence is diagnostic only because the
existing harness does not implement abstention.

Use the existing baseline and frozen bounded sparse correction, unchanged local
nomic embeddings, fresh memory service for every arm/question, all eight records
retained (recall50, pack20). The oracle is read only after predictions. No learning
updates. Verify saved source hashes and identical forecast laws across arms.
Record-level IDs only join corpus and evaluation; they are not ranking features.

The words before/after refer to the sequence of retained historical classifications,
not a claim that asteroid and dwarf-planet taxonomies are mutually exclusive.
The historical Pluto record answers the pre-decision question; the proposed draft
is a conditional alternative, not an adopted classification.

Methodological motivation: Shang et al. (2022),
[Improving Time Sensitivity for Question Answering over Temporal Knowledge Graphs](https://aclanthology.org/2022.acl-long.552.pdf),
sections1 and3: distinguishing a temporal anchor from its relation, and contrastive
questions differing in relation words. This harness does not implement TSQA's
trained temporal encoder or inherit its benchmark gains.

The current research rank callback exposes compressed features and baseline, not
candidate text. A text/graph-aware rescue must declare a new input boundary; it
must not secretly recover answer labels using fixture IDs.
