# Task plus lexical feature: actual service integration

## Result

All108 Recall calls completed under the separate task-lexical-overlay-v1 build.
Every one of the54 experimental top1 winners matches the frozen offline combined
comparator, including absent controls. Positive support remains12/12 semantic
roles,8/8 W3C,8/8 ESA written,7/8 ESA ISO and9/10 landing. No task-only or
ordinary-retrieval positive success is lost in these consumed fixtures.

The verifier checks input/source hashes, paired full-frontier law equality,
numeric ranker input equality, original packed scores, journal explanation
matching, packed-ID matching, and every offline winner. The lexical trace and
task plan are now journaled together with the snapshot and original-query digest.
Calibration remains explicitly not_evaluated; feature scores are not probabilities.

## Selection invariants

Lexical scores exist only in a local candidate copy used for packing. The task
class precedence is retained, with diversity penalties inside that selection.
Returned records are restored in full from original candidates. For pack1 this
matches offline order; for larger packets diversity may intentionally change
later slots, so no full-permutation equivalence is claimed.

Adaptive expansion is evaluated from original-order forecast laws before any
lexical reordering. A dedicated test would expand if the wrong boundary were
used; it stays unexpanded under both diversity settings. The same tests verify
full-record restoration, input immutability, token-budget fallback, evidence
occupancy suppression and rejection of duplicate permutation indices.

Race-enabled lexical and actual service tests pass under the new overlay,
including journal/cancellation-error handling, concurrency, stale retry and
adaptive expansion. These are targeted checks, not a complete repository audit.
No original runtime source, prior overlay, production configuration or whitepaper
was changed. The ordinary binary does not enable this research hook.

## Remaining gaps

All eight absent controls still return candidates. The two remaining positive
failures include unsupported later-than and wasn't phrasing. No new model was
trained and no generated agent answer was evaluated. These54 questions are
dependent, consumed design data, not54 independent population observations.

The earlier76/293 microsecond lexical component measurements exclude this full
hook's packing and persistence. Loaded end-to-end latency remains to be measured.
Next freeze this integrated variant for new relation/attribute and semantic-scope
tests, alongside cost. All seven whole research directions remain open.

## Reproduce

```sh
go test -race -tags researchpriority -overlay research/public-task-pilot/task-lexical-overlay-v1/overlay.json ./internal/researchcalendar -run 'TestLexical|TestPriorityOverlay' -count=1
node research/public-task-pilot/check-task-lexical-service.mjs
```

Runner cmd/public-task-lexical requires the same tag/overlay, fixture name and
NEW result path. Each of the five fixture directories retains
lexical-service-results.json. Protocol: TASK_LEXICAL_SERVICE_PROTOCOL.md.
