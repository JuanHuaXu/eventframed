# Untuned cross-subject retrieval screen

Freeze before executing new questions. Keep the NASA-trained sparse weights,
ASCII feature map and all service settings unchanged. No Nobel fitting or model
selection. New corpus adds six Nobel award motivations to13 prior NASA facts.
Two questions per new fact: literal and paraphrased. They are paired views of
six facts, not12 independent outcomes. Two absent-record questions are separate.

Compare baseline, lexical and sparse direct ranking using actual post-contract
CaptureTurn/Recall,19-record corpus, recall50/pack10. Three arms*14 queries.
No LLM calls. The new facts are public and may be in model pretraining; this is
a retrieval test only. No private text or invented factual assertions.

Positive screen: sparse top1 and support survival at least baseline AND lexical
in EACH wording group. Report top-score ties. Sparse Brier is not a daemon-law
metric and is not a success gate here. Absent tasks: report max/mean scores;
do not infer abstention from low means. Record all outputs and failures.

Source provenance: official Nobel Prize motivations from2022/2023 physics,
chemistry and medicine. Direct pages/PDF fetches returned403. Factual checks used
search-indexed excerpts of those primary pages, not successful full-page reads.
Source URLs are preserved with each fact. No claims beyond the motivations.
If this passes, different authors/domains, real embeddings and agent outcomes
remain necessary. If it fails, preserve the untuned result before any rescue.
