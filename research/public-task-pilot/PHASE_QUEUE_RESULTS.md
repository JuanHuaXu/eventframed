# Admission-phase classification: partial improvement only

The phase-aware research overlay passed queue/admission race tests ten times,
then completed the same32-arm durable workload. Production and dependencies
remain unchanged. Raw artifact: phase-queue-results.json. Verify with
`node research/public-task-pilot/check-batch-queue.mjs research/public-task-pilot/phase-queue-results.json`.

Only a synchronous trusted local adapter can mark a failure as never entered.
The queue continues after that marker, but still stops after unclassified commit
errors. Neither context errors alone nor matching error strings establish phase.

All16 low-rate arms pass. All16 high-rate arms fail. Each high-rate admission
group contains256 reads and128 writes:

| Admission | Read errors | Write errors | Responses >100ms | Stale rejections |
| --- | --- | --- | --- | --- |
| off |118|0|138|141|
| on |0|74|15|0|

Previous clean batch run had123/0 errors without admission and0/91 with admission.
Thus this run returned17 more acknowledged writes overall. This is a cross-run
description, not a randomized estimate or robust validation. Deadline misses
remain substantial and the whole-service success criterion is still false.

The74 write failures consist of46 proven never-entered admission rejections,
12 index-stage transaction timeout members, and16 later reconciliation stops.
No cancellation-string heuristic was used to continue after a transaction error.
All906 successful journals and438 acknowledged events survived orderly reopen.
Warmups passed and no future records appeared in successful frontiers. Failed
transaction members in this new run have not yet been individually reconciled;
the earlier91-write reconciliation does not cover them automatically.

## Decision

Retain the phase distinction as a tested research refinement, not a throughput
rescue. The remaining transactional index cost needs a scalable remedy; upstream
inspection found the current non-delta rebuild path still present. The next
backend experiment must preserve synchronous event/snapshot publication and
search visibility, not merely move index work behind an acknowledgement.

All seven whole goals remain open. No production, whitepaper, commit or push.
