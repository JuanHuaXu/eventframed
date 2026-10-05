# Date normalization rescue and boundary repair

Thirty offline public replay cases completed. No new embedding calls or oracle
access before ranking. The query grammar remains unchanged.

| Set | Before normalization | After strict normalization |
| --- | --- | --- |
| W3C written | 7/8 | 7/8 |
| ESA written | 8/8 | 8/8 |
| ESA ISO | 4/8 | 7/8 |

No correct prior-constraint case lost; both ISO exclusion questions now correct.
The ISO after-year, exact Venus date and arrival-date exclusion are rescued.
The remaining ISO miss is the unsupported later-than paraphrase. Absent-answer
controls remain unresolved. PASS the consumed-data normalization screen; this
does not validate new domains or the whole continuous-learning objective.

## Bug Hunt

The first normalizer wrongly accepted valid-looking prefixes of malformed
strings:2004-03-020,2004-03-02suffix and2004-03-02+01:00. Calendar equivalence
alone did not catch token-boundary failure. The first module and its artifacts
are preserved for audit, but it must not be adopted on its own.

The strict follow-up rejects unsupported token adjacency and time suffixes.
Boundary regressions and a full Gregorian400-year cycle passed against an
independent UTC calendar oracle:146097 valid dates and2703 invalid month/day
combinations. Written/ISO comparison decisions agree over that test domain.
This is not proof for all natural-language date extraction or all calendar eras.
All30 strict public records, including scores/laws/order, equal the initial
normalization records exactly. Unknown/mixed/invalid dates remain conservative.

Commands:

```sh
node --test research/public-task-pilot/date-normalized-strict.test.mjs research/public-task-pilot/date-strict-cycle.test.mjs
node research/public-task-pilot/check-date-normalized.mjs
node research/public-task-pilot/check-date-strict.mjs
```

Artifacts: normalized-strict-results.json in [W3C](w3c-focus-v1/normalized-strict-results.json),
[ESA written](esa-date-v1/normalized-strict-results.json), and
[ESA ISO](esa-iso-v1/normalized-strict-results.json). The earlier normalization
artifacts remain alongside them. No timing/latency conclusion from these tests.

The strict wrapper still depends on the original calendar module but checks
boundaries before it is called. It is a restricted research parser, not a
general temporal reasoner. Production remains unchanged. Next is an explicit
pre-packing typed input contract, not corpus-ID lookup or post-pack reranking
presented as full integration.
