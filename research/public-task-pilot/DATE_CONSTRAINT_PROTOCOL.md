# Explicit date constraints: research replacement

Frozen after W3C focus results, before replacement replay. W3C is now consumed
design data. Offline replay uses the saved focus arm for ranking but the ORIGINAL
question for constraints, and the original public corpus text for dates. This is
not wired into the daemon's feature-only research callback. No fixture IDs or
oracle labels enter the constraint scorer: joins recover text, then IDs are only
used for evaluation. Does not claim automatic general temporal understanding.

Parse a restricted English grammar: before/after one four-digit year in the
requested clause; `on DAY MONTH YEAR` there; and a whole explicit date following
`, rather than `. Preserve all supported constraints as a conjunction. Detect
multiple temporal bounds as unsupported. Unsupported or mixed exclusion text
fails closed by producing no constraints. Dates in records require exactly one
valid explicit calendar date. Missing, multiple or invalid dates are unknown,
not contradictions. Comparisons use calendar dates, with before/after YEAR
meaning strictly earlier/later calendar years, not fabricated interval endpoints.

Only DEMOTE candidates with an explicit contradiction below non-contradicted
candidates, preserving original order within each partition. Unknown is not
confirmation. If every candidate contradicts, report the condition and preserve
the original order: no silent fallback claim of correctness. Preserve original
scores/laws as provenance, without asserting the order remains score-sorted or
recalibrating their probabilities. Record demotion separately. Absent-answer
rejection remains unimplemented. This is a logic-filter experiment, not a law
correction or new learned intelligence.

Design screen: improve positive top1 over saved focus, no loss of a correct
focus result, and both essential-exclusion members correct. Report each case,
constraint extraction and contradiction. Test invalid dates, ambiguous bounds,
unknown records, no-rule queries, unsupported exclusions and original-question
retention. No new embedding requests or model fitting required. Future transfer
requires a fresh source/domain and ambiguous/exclusion controls; this cannot
complete roadmap goal5.
