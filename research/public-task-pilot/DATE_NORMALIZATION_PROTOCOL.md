# Date normalization ablation

Frozen after ESA results, before this replay. All current public fixtures are
consumed. Extend ONLY record-date recognition to standalone ISO YYYY-MM-DD;
query relation grammar, ambiguity policy, comparator and stable partition
semantics remain unchanged. No earlier/later synonym patch or status rules.
Keep date-constraints.mjs and previous artifacts intact.

Normalize a single valid ISO token to the existing English-date representation.
Multiple date mentions, mixed written/ISO mentions, invalid calendar dates and
ISO timestamps are unknown; they must not create a contradiction. Unknown is
not approval. This is calendar-date normalization, not timezone conversion,
temporal attachment, interval reasoning or generalized date extraction.

Replay saved focus candidates from W3C, ESA written and ESA ISO through the new
variant. Use original question and text only for scoring; labels read after
ranking. All candidate IDs, original scores and forecast laws remain preserved.
Compare against prior constraint results. Success: preserve written-date results,
improve ISO top1 with no previously correct ISO loss, and correctly distinguish
both ISO exclusion queries. Unsupported query grammar remains in denominators.

Test a full Gregorian400-year cycle with an independent UTC-calendar oracle,
including invalid month-day combinations, and mixed/invalid/ambiguous records.
ISO/written forms of the same valid date must yield identical contradiction
decisions. Any observed gains are consumed-data implementation rescue, not fresh
domain confirmation or evidence of continuous learning. No runtime integration
or performance claim from this replay.
