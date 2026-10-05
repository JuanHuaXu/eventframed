# Explicit date-constraint replacement

Offline replay of saved W3C focus candidates: positive top1 improves4/8 to7/8,
with all four correct focus cases preserved and BOTH necessary-exclusion
questions correct. PASS the consumed-design screen. The separate proposal/final
Recommendation distinction still fails; no status classifier was implemented.

The three rescued questions are exclusion of the September publication, the
pre-2016 publication and the post-2016 publication. The remaining positive
cases are unchanged. Unknown or unsupported syntax supplies no contradiction.
Both absent questions remain unresolved. No answer generation or learning.

The scorer sees original question and candidate text, not fixture IDs, sources
or oracle. It extracts recognized English calendar patterns and demotes
contradicted records. It does not learn temporal roles or infer missing dates.
Explicit calendar units matter: after2016 means a later calendar year, not a
later instant within2016. Date arithmetic and partition invariants are tested.

Original numeric scores and laws are preserved by record identity; the output
is a constraint-prioritized ordering, NOT a newly score-sorted list or calibrated
law. Unknown dates are not endorsed. An all-contradicted pack is flagged and
preserved, not silently declared correct. No all-contradicted case occurred in
this replay; unit coverage tests that branch.

Five test groups pass: invalid/leap dates, strict year comparisons, conjunction
of exact and excluded dates, unsupported/negative syntax, and stable partition
including unknown/all-contradicted records. The artifact verifier independently
checks the expected date-exclusion and before/after partitions as well as hashes,
coverage, ranks and law preservation.

Commands:

```sh
node --test research/public-task-pilot/date-constraints.test.mjs
node research/public-task-pilot/check-date-constraints.mjs
```

Artifact: [date-constraint-results.json](w3c-focus-v1/date-constraint-results.json).
This is offline, not daemon integration: the existing research rank callback
does not supply text to this module. The full W3C set was consumed before this
replacement was frozen. A new-domain test is required before claiming transfer
for this replacement. The source dates are unambiguous single-date statements;
multi-event records, attachment of dates to the wrong subject, implicit anchors,
temporal negation outside the explicit guards and linguistic paraphrases remain
important limitations. The finite pattern extractor is not a general parser.

## Next Step

Freeze this module and test it on a separate public event chronology with
unambiguous dates plus ambiguity/unknown controls. Keep the original question
through the future contract boundary. Do not add status keywords merely to
erase the remaining W3C failure or count these fixtures as continuous learning.
