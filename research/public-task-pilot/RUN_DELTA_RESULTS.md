# Buffered delta over immutable runs

Status: query-state prerequisite passed, not durable publication.

`RunDeltaSearch` captures a copied bounded delta as the newest virtual run,
ahead of all immutable graph manifests. Tombstones and downward corrections
shadow every older version, even when a corrected value is outside the emitted
delta top k. Exact delta scores use the same float64 cosine as graph rescoring.
Older graph prefixes compensate for all shadowed live IDs, not only nominated
delta candidates.

Bounds: at most15 graph runs plus the virtual delta, delta capacity supplied by
the caller (validated1..128), k<=200 and per-graph prefix<=200. The tests use64
as the pending-delta cap; nothing raises the load protocol's global budget.
Insufficient prefix capacity fails before any publication. Caller vectors and
run slices are copied; earlier snapshots remain independent.

## Evidence

`go test -race ./internal/researchindex -run TestRunDelta -count=3`
passed in1.728s. Real graph fixtures verify:

- A delta deletion suppresses matching versions in multiple older runs.
- A demoted delta record cannot cause an older high score to reappear.
- Historical and current views retain distinct correct answers after input edits.
- A flush of an earlier delta followed by a newer overwrite yields identical
  query results to the latest undrained delta over the original runs.
- Bounds, duplicate IDs, zero vectors, empty state and cancellation are checked.

The full bulk/candidate-only researchindex suite also passed three race-enabled
repetitions in17.057s using `research-candidate-only.mod`,
`candidate-only-overlay-v1/overlay-local.json` and tag `research_candidate_only`.

The before/after-flush comparison manually constructs both snapshots. It proves
the query-state equivalence for this fixture, NOT race-safe drain/publication.
No semantic revision, authoritative commit or lost-ack recovery is supplied by
this object. Existing graph locks prevent use-after-close but do not replace
cross-run leases.

## Cost and next implementation

Plan construction currently scans/copies all manifests; it is not a constant-time
per-write operation. A durable writer must preflight that cost before commit,
or use a proven incrementally maintained immutable ownership plan. Do not put
fallible planning after a durable acknowledgement. Delta scoring costs O(B*d),
where B is bounded buffered records and d the vector dimension; graph fanout and
prefix costs remain additional.

Next implement prepared write state and one semantic revision, durable callback
quarantine, and flush publication that removes only captured delta versions,
retaining every newer overwrite. Then consolidate runs with reader leases and
test sustained offered load through repeated compaction. Existing static speed
results do not cover these costs. All seven whole goals remain open.
