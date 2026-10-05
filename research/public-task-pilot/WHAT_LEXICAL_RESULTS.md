# What-field lexical feature: failed first pass and normalization rescue

## Result

Lexical overlap inside the task plan's noncontradicted/contradicted classes
rescues event-relation matching without adding a landing dictionary. This is
an offline top1 computation on actual saved service frontiers, not yet a runtime
packing implementation or fresh transfer test.

| Fixture | Task only | Lexical v2 only | Task + lexical v2 |
| --- | ---: | ---: | ---: |
| Semantic roles | 12/12 | 8/12 | 12/12 |
| W3C | 7/8 | 4/8 | 8/8 |
| ESA written | 8/8 | 5/8 | 8/8 |
| ESA ISO | 7/8 | 3/8 | 7/8 |
| Landing transfer, now consumed | 6/10 | 7/10 | 9/10 |

No task-aware positive success is lost in v2. Pure lexical retrieval is not a
replacement: removing date/intent constraints loses substantial accuracy.
The combined screen requires no regressions per set, not merely net improvement.

## Confirmed normalization bug

V1 removed all numbers. It improved landing6/10->9/10 but regressed W3C's
exclude-november case from the correct HTML5.1 record to HTML5.2. Its combined
screen therefore FAILED despite unchanged W3C aggregate7/8. Generic number
removal erased meaningful version distinctions.

V2 preserves integer/decimal tokens while removing only full ISO or English
day-month-year date spans. No fixture lookup, topic vocabulary or threshold
change was added. Unit tests reproduce the v1 version collision, show v2
distinguishing otherwise identical versions, and check date-format invariance,
input immutability and rejection of unframed text. V1 artifacts remain intact.

This normalization still is not general entity resolution. Alphabetic tokenizing
and simple suffix stripping are lossy English heuristics. Date-span removal is
not parsing of arbitrary timestamps, units, mathematical values or identities.

## Scope

The comparator uses effective query, actual what field, candidate score for ties,
and the existing task plan's state partition. It never reads label IDs before
ordering. The score is binary-term cosine with smoothed frontier IDF, not a
calibrated probability. This changes neither stored numeric scores nor proper
laws in the offline artifact, but does NOT prove unchanged live-service behavior.

All54 queries across the five sets are consumed. There are46 positive cases
and8 absent cases, with repeated facts and dependent questions. The algorithm
always returns an order, so absent-answer handling remains unresolved. The
landing "wasn't" contraction failure is not fixed by lexical matching; it is
still a task-polarity defect. Do not claim learned domain translation from these
handwritten features or treat aggregate totals as independent observations.

## Reproduce and next step

Protocols WHAT_LEXICAL_PROTOCOL.md and WHAT_LEXICAL_V2_PROTOCOL.md were written
before their respective runs. Artifacts what-lexical-results.json and
what-lexical-v2-results.json contain per-query orders/scores, labels joined after
ordering, summaries, and hashes of comparator, runner, protocol and input files.
The failed v1 result remains a negative control.

```sh
node research/public-task-pilot/check-what-lexical-unit.mjs
```

Each run-what-lexical script takes a NEW output path and rejects overwriting.
Next port the frozen feature to the research service hook, verify numerical
agreement with these traces, preserve token/diversity rules, and measure cost.
New relation/attribute and adversarial scope tests are required after integration.
All seven whole goals remain open; production and whitepaper are unchanged.
