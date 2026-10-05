# Fast priority: service replay

The equivalent fast packing helper is now called by the actual Recall path in
priority-overlay-v2. Original runtime sources and priority-overlay-v1 remain
unchanged. The v2 generator changes the hook call only; the policy and explanation
semantics are the same. This is an opt-in research build, not production.

## Checks

- Combined calendar/priority tests passed with the race detector under v2.
- Sixty fresh-memory public replay calls completed under v2.
- The verifier compares EVERY recorded result field against v1, excluding only
  RecallNS: exact equality, including scores, laws, explanations, input/output
  ranker records, packet certainty and journal-match status.
- Top1 remains W3C7/8, ESA written8/8, ESA ISO7/8. No baseline success was lost.
- All30 priority journals matched packet explanations, and original/overlay
  source hashes verify. Absent cases still all pack a candidate.
- An additional actual Recall test, repeated three times under the race
  detector, enables adaptive expansion with MaxPack2 and initial pack1. It
  verifies expansion, original scores, temporal first place, and matching
  two-item journal explanation.

The public replay uses the prior fixed pack1 protocol, without adaptive
expansion. The adaptive case is covered by the separate actual-service test;
do not present the public run as a loaded adaptive benchmark.

## Important interpretation

The adaptive test returns BOTH Rosetta records, including the event explicitly
excluded by the question in the second slot. Priority is an ordering preference,
not a hard filter. Top1 support accuracy therefore does not establish packet
consistency or generated-answer accuracy. The all-contradicted path likewise
falls back to ordinary packing. A subsequent artifact census found that all six
current absent-answer cases are actually all-UNKNOWN, not all-contradicted;
their failure cannot be attributed to the all-contradicted fallback.

The earlier approximately50% adaptive-hook improvement remains component timing,
not a measured full-Recall speedup. Loaded serving and durable-backend costs
remain unmeasured for this variant.

## Reproduce

```sh
go test -race -tags researchpriority -overlay research/public-task-pilot/priority-overlay-v2/overlay.json ./internal/researchcalendar ./internal/service -run 'TestPriority|TestCalendar' -count=1
go test -race -tags researchpriority -overlay research/public-task-pilot/priority-overlay-v2/overlay.json ./internal/researchcalendar -run TestPriorityOverlayAdaptiveExpansion -count=3 -v
node research/public-task-pilot/check-calendar-priority-fast.mjs
```

Runner: cmd/public-calendar-priority-fast, same tag/overlay, fixture name and NEW
output JSON path. Each fixture's priority-fast-results.json is retained alongside
the unchanged earlier artifacts. Reuse the frozen CALENDAR_PRIORITY_OVERLAY_PROTOCOL.md.

Next research steps are loaded cost and explicit contradiction/abstention
contracts, including unknown and ambiguous cases. A detector recognizing some
calendar predicates must not infer absence from unrecognized language or call
an empty packet proof that no event exists. Actual agent usefulness and all
seven whole directions remain open.
