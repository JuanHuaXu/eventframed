# Calendar predicates do not establish exclusion authority

## New evidence

Five executable counterexamples confirm that the existing substring parser can
mark relevant evidence contradicted:

| Request role | Query example | Relevant evidence |
| --- | --- | --- |
| Yes/no question | Was Rosetta's arrival before 2010? | Arrival in2014 answers no |
| Falsification | Which record disproves the claim that Rosetta arrived before 2010? | Arrival in2014 refutes it |
| Contracted negation | Which events aren't before 2010? | Arrival in2014 qualifies |
| Disjunction | Which events occurred before 2000 or involved Rosetta? | Rosetta launch in2004 qualifies |
| Quotation | Explain why the claim 'Rosetta arrived before 2010' is false. | Arrival in2014 explains why |

All five were classified contradicted. Adjacent controls confirm that ordinary
positive selection before2010 is correctly contradicted by2014, while explicit
"did not" yields unknown under the existing negation guard. The flaw is not date
arithmetic: the parser does not bind a date phrase to its logical role in the
request. A hard filter built on this state would remove necessary evidence.
Soft demotion can also harm these tasks; this finding limits the existing ranker,
not only a hypothetical future hard filter.

The counterexample tests PASS when reproducing this unsafe classification. That
is a confirmed negative result, not a passed semantic-correctness gate. No runtime
or frozen parser was changed and no exclusion authority was added.

```sh
go test -race ./internal/researchcalendar -run TestCalendarPredicateAuthority -v
```

Test source: internal/researchcalendar/semantic_authority_test.go. The public fact
content reuses the previously verified Rosetta dates; the questions are designed
semantic controls, not new natural-user observations.

## Absent-answer diagnosis correction

Inspection of the actual priority-fast-results.json explanation plans shows:

- W3C absent-53 and absent-vote: all4 candidates unknown.
- ESA written absent-2020 and absent-cost: all4 candidates unknown.
- ESA ISO absent-2020 and absent-cost: all4 candidates unknown.

AllContradicted is false in all six cases. Merely abstaining on that flag would
rescue NONE of these failures. The earlier fast-service note attributing some
absent failures to that fallback has been corrected. Unrecognized predicates
and missing requested attributes are distinct from recognized contradictions.

## Revised research direction

Before changing packet membership, represent the task as an explicit plan with
answer type (selection, truth assessment, explanation), target entity/relation,
predicate polarity, conjunction/disjunction and quoted-vs-asserted scope. A
candidate date mention must also be bound to the requested event relation, not
merely found in text. Unsupported interpretations must remain unsupported, not
be relabeled safe. A syntactically valid plan alone does not prove interpretation.

Constrained semantic parsing is a relevant implementation lead: [PICARD](https://aclanthology.org/2021.emnlp-main.779/)
uses incremental parsing to restrict generated formal-language outputs. It is
not a guarantee that the chosen query expresses the user's intent. This is a
methodological lead, not an implemented PICARD model in EventFrame.

Measure BOTH relevant-evidence deletion and irrelevant-evidence retention, plus
answerability/coverage. A future calibrated set selector could build on
[risk-controlling prediction sets](https://arxiv.org/abs/2101.02703), but requires
its own labeled calibration data and appropriate assumptions; the present
consumed fixtures cannot supply a new generalization guarantee. No confidence
threshold should be invented from raw rank gaps or unknown parser states.

Proposed paired tests: hold event facts constant and vary only request role;
then hold the request constant and vary asserted/quoted/negated fact scope.
Include absence of an attribute, absence of an entity, explicit falsification,
and positive selection. Compare ordinary retrieval, current temporal priority,
and a separately frozen task-plan variant at equal frontier and labeling cost.
Report packet-level evidence recall as well as top1 support and abstention.

This changes the next semantic experiment, not the original seven success
criteria. All directions remain open; loaded performance is a separate pending
engineering experiment.
