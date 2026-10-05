# V63 bank-owned journal results

Storage/ownership component PASS; no whole research-goal completion.
The immutable V62 checkpoint was verified before cloning. Production/private/
paper/publication and14preexisting dirty paths remain unchanged.

## Correctness

26functional race roots pass, including child-level mutation rejection for all
five standalone mutators, canonical BANK-issued metadata, shared-journal fault
rejection, all-or-none child/selector publication, epoch/foreign replay fences,
expired-factor protection and nonvacuous future forks. Standalone Model APIs
remain independently mutable. Raw journal and sequence index are bank-owned;
each posterior/member count/support/window stays independent.

26,394public comparisons are BITWISE identical to V62 across depth0/2/7 and
delay0/3/11, including issued/query/request/receipt laws, metadata, cancellation,
pending counts and epoch reset. This intentionally does not demand unchanged
private child forecast metadata: the canonical journal now stores the bank law.
25,758independent tree/selector forecast comparisons,2,000explicit selector path
checks,4,656legacy forecasts and120prediction values also pass. No concurrent
API or durable publication guarantee is established by this serial race run.

## Allocation And Cost

Constructor TotalAlloc (NOTRSS), unchanged8,388,608-byte gate:

| Members | V62 shared | V63 journal | Unshared control | V63 gate |
| --- | ---: | ---: | ---: | --- |
| 150 | 7,951,176 | 6,389,264 | 10,001,168 | PASS |
| 200 | 9,988,752 | 7,924,368 | 12,735,632 | PASS |

The200-member cap is preserved; no shrinking of the frontier or memory gate.
Three constructor benchmark repetitions:150members2.845-2.858ms,200members
3.582-3.606ms,22allocations. Mixed Predict792.6-815.6ns,zero allocations.
The2,400-frame/400fixed-paired-packet core loop takes162.146-166.895ms and
6,389,264allocated bytes,22allocations. This includes3windows,selector,
expiry/updates and16all-member snapshots, but NOT adaptive nomination, delay
queues, persistence/RPC/backend/concurrent serving. Do NOT call it the full
400ms quality gate or goal6proof.

All4frozen commands (race/vet/allocation/benchmark) terminal0. Compiler closure
and generated-main copies are under `research/retention-v63-integration`.
V62's old200-member failure remains preserved. Next: the complete actual-mixture
quality ablation, independent replay and full nomination/delay cost; calibrated
observer mathematics, certified useful splits, untouched tasks and loaded
freshness remain required under ALLseven original goals.
