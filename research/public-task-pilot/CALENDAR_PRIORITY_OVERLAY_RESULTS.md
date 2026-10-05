# Temporal priority: opt-in full-service integration

## Result and scope

The post-correction priority now reaches the actual Recall packing boundary in
an opt-in Go build overlay. It is not enabled in the ordinary daemon. The base
service/model sources were not edited; the overlay manifest records their hashes
and the generated variants. No whitepaper or production configuration changed.

The consumed-data retrieval screen passes its accuracy, score preservation,
full-frontier law preservation, and journal-explanation checks:

| Public fixture | Focus top1 | Priority top1 | Rescued | Lost |
| --- | ---: | ---: | ---: | ---: |
| W3C | 4/8 | 7/8 | 3 | 0 |
| ESA written dates | 6/8 | 8/8 | 2 | 0 |
| ESA ISO dates | 4/8 | 7/8 | 3 | 0 |

Sixty calls used fresh memory stores and the local embedding-only model. All30
priority explanations matched the stored journal. Full-frontier proper laws
matched the control exactly by identity; packed scores matched original numeric
ranker scores. Both arms used diversity packing. There were no learned cache
records in this public replay; opposing cached deltas remain separately tested
at2/50/200 candidates, not covered by these public end-to-end calls.

All six absent-answer cases still packed a candidate. This is NOT an abstention
rescue. Data was previously consumed; the ESA ISO cases repeat the written facts.
There is no new generalization evidence, generated answer, or continuous-learning
validation. All seven whole research directions remain open.

## Boundary and tests

The priority is computed after numeric rank corrections and before journal
insertion. Its original-query digest, typed predicate decisions, snapshot, as-of
time and packed IDs enter a JSON explanation string in both packet and journal.
The full journal hash binds that explanation. The existing snapshot compatibility
check still rejects stale publication. The enabled policy enters the policy
fingerprint; omitempty preserves the disabled fingerprint encoding.

Packet calibration is explicitly `not_evaluated`. The legacy numeric certainty
is0 as a sentinel, not an estimated zero probability. Earlier delta modulation
still sees the original score-based certainty. No artificial priority scores
escape the selection-local packing implementation.

Direct full-service tests with race instrumentation passed three repetitions:

- Deliberately dominant contradicted candidate loses the final packing slot,
  even with diversity enabled; winner retains original score0.1.
- Opposite exclusions sharing a shortened retrieval query produce different
  original-query digests, distinct journal IDs and appropriate winners.
- Returned explanation matches the journal, snapshot and packed IDs.
- Policy motion immediately before journal insertion causes exactly one retry;
  the returned explanation binds the new snapshot.
- Disabled control carries neither priority explanation nor calibration label.

The combined calendar/priority test selection also passed under the overlay with
`-race`, and under the ordinary build without the overlay. These are targeted
tests, not the entire repository suite. Snapshot equality is directly asserted
in the service tests; the public artifact stores the explanation and journal
equality result, not a separate copy of each packet snapshot.

## Reproduce

The existing overlay is rooted at this checkout's absolute path. For a relocated
checkout, generate a NEW directory with build-priority-overlay.mjs, then pass
that overlay explicitly. Do not overwrite old artifact paths.

```sh
go test -race -tags researchpriority -overlay research/public-task-pilot/priority-overlay-v1/overlay.json ./internal/researchcalendar -run TestPriorityOverlay -count=3 -v
go test -race -tags researchpriority -overlay research/public-task-pilot/priority-overlay-v1/overlay.json ./internal/researchcalendar ./internal/service -run 'TestPriority|TestCalendar' -count=1
go test ./internal/researchcalendar ./internal/service -run 'TestPriority|TestCalendar' -count=1
node research/public-task-pilot/check-calendar-priority-overlay.mjs
```

The public runner is cmd/public-calendar-priority, requiring the same tag and
overlay, then a fixture name and NEW JSON output path. Artifacts are the three
fixture directories' priority-overlay-results.json. The verifier checks their
source hashes, original/overlay hashes, original-query digests, law equality,
scores, journal equality, accuracy, and retained absent-answer failures.

Next: broader lifecycle/failure handling for this exact hook, component and loaded
service cost, and then untouched tasks with explicit absent/ambiguous outcomes.
The larger goal remains actual agent-usefulness improvement with reliable
learning, not calendar parsing alone. No serving-latency claim is made here.
