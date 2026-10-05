# Preserved V2 mapping failure

2026-10-04. `node research/public-task-pilot/scifact-provenance-v2.mjs` exited1.
Its validator incorrectly equated BEIR qrels with upstream evidence keys.
First actual mismatch: cited document31715818 is a positive BEIR relation,
while the original claim's evidence map is empty. Downloaded archive/license/
download-script and the original V2 producer remain; no prepared data was changed.

Inspection of ALL rows finds exact unique-cited-document mapping:809train and
300dev claims. One dev claim has a repeated citation ID, so list equality without
deduplication would falsely reject it. Original evidence-only matches480train,
175dev;304train and112dev claims have empty evidence maps. Thus the correct
repair is a NEW provenance audit, not relabeling qrels or choosing easier queries.
V3 must verify exact text and unique citation sets, report annotation categories
separately, and distinguish BEIR test from upstream labeled dev. No inference
that citation relevance establishes support, truth or causal identification.
