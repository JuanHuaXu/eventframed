# Frozen date-constraint transfer

Freeze before predictions. New ESA mission chronology, no W3C/astronomy
classification records reused. Eight positive queries and two absence controls;
four factual records in each of two corpus formats. ISO and written dates encode
the same facts, so they are paired stress conditions, not independent samples.
Primary ESA pages were opened and dates checked before corpus creation.

No changes to focusQuery or date-constraints.mjs. Actual isolated CaptureTurn/Recall
baseline/focus with fresh memory service and local nomic embeddings. Then replay
focus records through frozen constraints using ORIGINAL questions and corpus
text. Oracle access only after ranking. Preserve laws/scores by record identity
in the constraint stage; focus itself can change them. No agent generation,
learning updates, production access or hot-path promotion.

Transfer criterion for EACH format: strictly improved positive top1 versus focus,
no loss of focus-correct cases, both Rosetta exclusion members correct. Report
baseline/focus/constraint separately. Six positive queries use supported grammar;
two use earlier/later-than paraphrases deliberately outside the grammar. No
exclusion of failures from denominators. Written and ISO scores must not be
pooled to conceal unsupported input. ISO record dates are expected to be unknown
to the frozen parser, so improvements from constraints may disappear.

Separate ambiguity checks use a true compound statement containing Rosetta's
launch and arrival dates and a source-derived statement with no date. Neither
must supply a single-date contradiction. These are parser safety checks, not
successful question answering or corpus-size evidence. No new phrasing patches
after observing results. All fixtures become consumed after this test.

Sources: ESA Mars Express operations, Venus Express operations and Rosetta
mission overview, linked in corpus.json. Ingestion timestamps are common research
availability; dates inside text are historical event dates, not publication or
availability dates. Dates reformatted as ISO retain the same calendar day.
