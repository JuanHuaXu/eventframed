# Real HNSW net-state-change audit

Status: useful ownership evidence, not a bounded-update theorem or speed result.

An overlay adds only a diagnostic test to the copied pinned backend. Two initial
graphs (800 and6400 records,768d) each undergo8 insertions and8 deletions. Four
deletions target seed IDs; four select a currently high-backlink-count node.
RepairEnabled=false and no concurrent readers/writers during snapshots. Before
and after every operation, capture all logical nodes: ID, level, vector hash,
outgoing links, backlinks and heuristic flags, plus global entry-point/level state.

## Results

| Graph size | Insert changed records | Delete changed records |
| --- | --- | --- |
| 800 | 66-113 | 31-77 |
| 6400 | 72-111 | 38-106 |

One deletion at6400 changes packed global state as well. Across captured graphs,
maximum outgoing links summed over levels is160, maximum backlinks160, and maximum
combined links313. These are not per-level degree counts. The flat64-edge graph
ownership prototype cannot directly represent this layered/backlink state.

The corrected test passes in2.357s; the independent verifier recomputes all32
changed-record counts from full before/after artifacts. Each sampled operation
fits the prototype's128-record edit count, but that does not prove a universal
bound, nor imply a32-event batch fits128 records. Transient writes that restore
their old value, search/read sets, allocator changes and ID-map internals are not
counted. This measures NET LOGICAL CHANGES, not every touched allocation.

## Harness correction and preserved failure

The first run crashed in JSON string encoding after graph close. The diagnostic
had retained a borrowed ID from `segmentedStringArray.Get`, which returns
`unsafe.String` over index storage. Snapshot IDs now use `strings.Clone`; vectors
were already hashed and link slices copied. The failed source is preserved as
`hnsw-touch-test-failed-v1.go.txt`; its output `hnsw-touch-results.json` is incomplete
and must not be used. This was a harness ownership violation, not evidence of a
backend mutation defect. Project-local lesson: diagnostic snapshots that outlive
an off-heap owner must copy strings as well as slices.

## Artifacts and next action

`hnsw-touch-test.go.txt`, `hnsw-touch-overlay.json`, `hnsw-touch-results-v2.json`,
`check-hnsw-touch.mjs`. Test command runs in the copied backend module:

```sh
RESEARCH_TOUCH_OUTPUT=NEW-absolute-output.json go test -overlay <LOCAL_ROOT>/research/public-task-pilot/hnsw-touch-overlay.json ./internal/index/hnsw -run TestResearchTouchedGraphState -count=1
```

Next model immutable layered adjacency/backlinks and shared vector payloads,
including a versioned global entry point, then replay these logical changes and
verify complete captured state plus old-root retention. That remains distinct
from executing the actual HNSW mutation algorithm privately before durability.
Reject edit-budget overflow without truncating repairs. No normal dependency,
production or whitepaper changes. All seven research goals remain open.
