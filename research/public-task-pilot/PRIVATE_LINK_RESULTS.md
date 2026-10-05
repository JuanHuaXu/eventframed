# Private insertion link decision

Implemented `DecidePrivateLink`, a pure bounded serial live-record decision for
slack append, diversity pruning at capacity, and the already-heuristic shortcut.
It returns owned kept/dropped IDs, acceptance, heuristic count and metric work.
Cancellation, invalid metric output and budget exhaustion return no partial
decision. It is not yet a graph mutation: backlink removal/addition, immutable
root composition, persistence and publication remain caller work.

Backend comparison uses 40 deterministic 768-dimensional vectors and 24
artificial target-list cases: counts 0/7/11/12, heuristic flag unset/set and three
new neighbors. Pair distances are backend-captured. These are branch controls,
not naturally occurring insertions, and do not validate graph-wide backlinks.

The first comparison failed at the first full non-heuristic case. The private
prototype conflated maxM with physical capacity. The backend defaults to a 2.25
level-zero multiplier, not 2: M=4 means selector target 9, physical capacity 12.
Both are now explicit inputs. `link-capture.json` retains the original outcomes
but lacks the declared limits; `link-capture-v2.json` records them explicitly.
All 24 captured kept-ID orders, acceptance and heuristic counts now agree;
dropped IDs are checked against the original-minus-kept set. Three race-enabled
repetitions of local and backend checks passed (1.311s).

This discovery also narrowed the prior construction comparison: it tested a
caller-imposed level-zero cap of 8, not the backend default. The new
`construction-capture-v3.json` records actual selector limits (9 at level zero).
`construction-comparison-v4.json` reports zero candidate-set, selected-set or
selected-order mismatches across all 90 cases (0.262s). Earlier capped results
remain valid only at their explicitly tested cap.

The ordinary research package suite passed (4.904s). This is component
correctness evidence, not a sustained-load or whole-goal success. Remaining:
compose real insertion including backlink updates, compare complete graph state,
then integrate with bounded durable publication and rerun the failed load gate.
