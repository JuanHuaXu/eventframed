# Public context-discrimination pilot

New hand-designed exploratory corpus from four opened primary-source pages:
NASA Pluto facts, NASA Dawn Ceres overview, IAU proposed2006 draft and adopted
resolution. Corpus records paraphrase their stated historical/proposed/adopted
claims, not external truth labels. Ceres asteroid/dwarf labels are not asserted
to be mutually exclusive under every taxonomy. No invented exact historical
transition date is supplied. Dataset preserves proposal modality explicitly.

Eight positive context-specific questions and two unsupported questions. Read
oracle only after all retrievals. Eight records retained, recall50/pack20.
Run semantic baseline and the previously frozen bounded sparse correction,
without fitting, new thresholds or oracle-informed context filters. All record
timestamps represent present ingestion; historical time is text, not falsely
backdated availability. All records are ingested before all questions.

Success screen: bounded preserves all baseline correct cases and corrects at
least one additional positive; full law equality and delta<=.005. Report each
context separately; failure on a historical/proposed query is not excused by
returning a factually valid statement from a different context. Unsupported
questions have no retained support; no abstention accuracy claim without a
declared abstention mechanism. This is not a learning/feedback experiment.

Existing local nomic embedding only, no generation calls. Actual CaptureTurn/
Recall, memory store, sequential runs, no production/private data. Sources are
linked per corpus record; shared sources and eight tasks are not independent
population evidence. Do not tune this set and then call it untouched confirmation.
