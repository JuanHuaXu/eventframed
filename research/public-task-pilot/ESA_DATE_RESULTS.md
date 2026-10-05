# Frozen date constraints: ESA chronology transfer

Completed40 isolated CaptureTurn/Recall runs and20 offline constraint replays.
Four new public event records, eight positive questions plus two absence controls,
tested in two corpus date formats. The same frozen focus and date-rule functions
were used; their identities/hashes were verified against prior artifacts.

| Corpus date format | Baseline top1 | Focus top1 | Constraint top1 | Transfer screen |
| --- | --- | --- | --- | --- |
| Written English | 5/8 | 6/8 | 8/8 | PASS |
| ISO calendar date | 3/8 | 4/8 | 4/8 | FAIL |

No focus-correct loss in either format. Both exclusion members are correct with
written dates, but not with ISO dates. Written-date rescues are the Venus Express
launch-date query and the Rosetta arrival-date exclusion. ISO dates are ALL
unknown to the frozen record parser; the constraint stage leaves every ISO
candidate order unchanged. This is a confirmed coverage limit, not evidence that
the calendar comparisons themselves fail. Earlier/later-than paraphrases have
no extracted rule: they happen to be correct under the written-date baseline,
but the ISO later-than query is wrong. Do not credit the parser for those cases.

These formats represent the SAME four facts, not independent observations.
Eight is a question count, not a statistically adequate population sample. This
is new-task-family transfer from W3C publication chronology to mission events,
not proof of general temporal language understanding. All fixtures are now
consumed. Both absence controls remain unresolved; no abstention is implemented.

The ambiguity checks passed: a true compound Rosetta launch/arrival sentence
does not supply a unique date or a contradiction, and an undated true mission
statement supplies no date. This checks conservative parser behavior only, not
the ability to answer from multi-event text.

Constraint replay preserves saved score/law values by record identity and only
changes their priority ordering. There is no calibration claim. Actual recall
uses memory backend and local nomic embeddings; constraints remain an OFFLINE
text-aware layer, not the daemon's existing feature-only rank callback. No
generation calls, fitting, feedback updates, production changes or latency claim.

## Evidence

- [Written-date recall](esa-date-v1/results.json) and [constraint replay](esa-date-v1/date-constraint-results.json).
- [ISO-date recall](esa-iso-v1/results.json) and [constraint replay](esa-iso-v1/date-constraint-results.json).
- `node research/public-task-pilot/check-esa-constraints.mjs` verifies hashes,
  frozen function identities, expected cases/records/ranks, score/law preservation,
  unchanged unsupported-query/ISO ordering and ambiguity controls.

Primary fact sources: [ESA Mars Express operations](https://www.esa.int/Enabling_Support/Operations/Mars_Express_operations),
[ESA Venus Express operations](https://www.esa.int/Enabling_Support/Operations/Venus_Express_operations),
and [ESA Rosetta mission](https://www.esa.int/Science_Exploration/Space_Science/Rosetta).
Only the four historical event dates were paraphrased into the corpus; source
page publication dates were not used as event or availability dates.

## Decision

Retain the constraint approach as a supported limited component lead, but do not
promote this parser. Normalize declared date formats with ambiguity rejection,
and retain the original requested relation and exclusion through an explicit
runtime input boundary. Test that normalization is semantics-preserving across
valid representations and rejects invalid/mixed ambiguous dates. A richer input
representation can help the learner, but these deterministic comparisons are
not continuous learning and do not complete goal5 or any other whole direction.
