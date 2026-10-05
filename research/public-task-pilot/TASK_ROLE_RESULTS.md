# Task-aware candidate: design rescue

## Result

The separately frozen task-role-v1 candidate passes the combined consumed-data
screen over84 actual service calls:

| Fixture | Ordinary retrieval | Previous priority | Task-aware candidate |
| --- | ---: | ---: | ---: |
| Semantic role pairs | 6/12 | 3/12 | 12/12 |
| W3C | 4/8 | 7/8 | 7/8 |
| ESA written dates | 6/8 | 8/8 | 8/8 |
| ESA ISO dates | 4/8 | 7/8 | 7/8 |

No correct ordinary-retrieval case was lost. The new candidate rescues all six
ordinary failures in the role fixture, not merely the four failures introduced
by the earlier priority rule. This is a real finite design rescue, not evidence
that general semantic interpretation or continuous learning has been solved.

## Mechanism

An explicit task plan separates selection from lookup/assessment, names a target
launch/arrival relation when exactly one is recognized, and records selection
negation. Selection applies date constraints; assessment retrieves the target
relation regardless of whether its date confirms or refutes the proposition.
Recognized negated selection complements a single predicate. Unsupported
alternatives are marked unsupported and do not apply this ordering.

Candidate what-fields supply relation mentions and dates. The plan retains
per-candidate reasons in its journaled explanation. It changes ordering only:
numeric ranker output and full-frontier forecast laws exactly match control in
the replay; calibration remains not_evaluated. There is still no hard exclusion,
abstention, factual authentication or semantic-truth certificate.

The implementation is handwritten, bounded vocabulary/role interpretation.
It does not read case IDs, fixture labels, or oracle outcomes during ranking.
It does not learn these rules. Multi-relation mentions, arbitrary entities,
quoted/negated event assertions, general Boolean scope and missing attributes
remain unresolved. In particular, an event mentioning a relation is not thereby
proved to assert it. A finite pass grants no exclusion authority.

All six earlier absent cases still return a candidate. Lookup of an unrecorded
cost is not solved by identifying launch as its target relation.

## Verification

Task-role unit tests pass with race instrumentation. Existing actual service
journal, concurrency, stale retry, injected failure and adaptive tests also pass
under task-overlay-v1. These are targeted tests, not a complete repository audit.
Original parser/ranker/overlays and their failed results remain unchanged.

```sh
go test -race ./internal/researchcalendar -run TestTaskRolePlan -v
go test -race -tags researchpriority -overlay research/public-task-pilot/task-overlay-v1/overlay.json ./internal/researchcalendar ./internal/service -run 'TestTaskRole|TestPriorityOverlay' -count=1
node research/public-task-pilot/check-task-role.mjs
```

The runner is cmd/public-task-role, using the tag/overlay above, a fixture name,
and NEW output JSON path. All four fixture directories retain task-results.json.
TASK_ROLE_PROTOCOL.md was written before dispatch. The verifier checks source
hashes, paired numeric inputs/laws, support membership, journal equality and the
combined accuracy/no-regression criteria. Oracle joins occur after predictions.

Next: freeze this candidate, test new paraphrase/role families and event types,
including ambiguous and absent answers; inspect entity/attribute binding failures
before adding authority. Measure this variant's cost separately. All seven whole
research goals remain open; production and whitepaper are unchanged.
