# RFC semantic transfer screen

Frozen before the first model/service run on 2026-10-01. The existing
task-plus-lexical overlay and nomic-embed-text configuration are unchanged.
The new confirmation set has 15 public HTTP facts from [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110.html)
and [RFC 9111](https://www.rfc-editor.org/rfc/rfc9111.html), paired literal
and paraphrased questions per fact, and two absent-from-corpus controls.
Thirteen previously used NASA facts are fixed distractors. All 15 RFC
fact clusters are new to this research overlay; paired wordings are not
independent observations. Source summaries are paraphrases, not copied
passages. Do not change questions, support labels, coefficients or selection
rules after examining model output.

Use the existing opt-in research overlay on current unmodified service
sources. For every query create a fresh in-memory Service, CaptureTurn all
28 records, and Recall with semantic embeddings, recall50, pack10, diversity
enabled, identical effective question and source order. Compare `focus`
(ordinary packing) with `priority` (frozen task-plus-lexical packing). No
feedback, fitting, real agent conversation or LLM generation. Nomination
must happen before packing. Keep the numeric ranker as passthrough.

Before reading the oracle, retain raw ordered candidates, full-frontier
forecast laws, ranker before/after traces, journal match and timings. Score
top1 and support survival by wording and by 15 fact clusters. A finite
useful-transfer pass requires, separately for literal and paraphrase:

1. Priority top1 and survival no worse than focus.
2. At least one additional correct top1 across all 30 positive queries.
3. Identical full-frontier numeric scores and forecast laws across arms,
   with every priority explanation matching its persisted journal.

Report absent controls as packed/unsupported; do not equate packing an
irrelevant record with answering correctly or claim abstention. Report ties
and failure IDs. No population interval from only 15 paired clusters. The
timings are sequential isolated request durations, not concurrent p99 or
OpenClaw latency. A failure is retained before any design rescue. Do not
touch production OpenClaw or the default runtime.
