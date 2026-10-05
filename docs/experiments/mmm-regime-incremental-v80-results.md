# V80 equivalent computation rescue

All seven WHOLE goals remain OPEN. No daemon adoption, new empirical accuracy,
recovery/non-harm, useful-split, agent-utility or equal-TOTAL-cost success claim.
V77 approximation error/vacuous bounds and V79 excluded noisy explanations are
unchanged. The change is computational, not an observation-selection or law change.

## Implemented invariant

Separate `internal/researchregimeincremental` fork; V79 remains frozen.
Issue advances the current inference state once. Four immutable cached prefixes,
stride64 original issues, accelerate nearby delayed updates/queries. A changed
row uses only a prefix that EXCLUDES it. Later prefixes are rebuilt on successful
updates; refresh rebuilds all caches. Query/failed update cannot mutate them.
Outside the retained cache, old evidence falls back to equivalent full replay;
history is not truncated, silently discarded or given a different law. Retained
support, same-Y likelihood, arithmetic order, log normalizers, envelope/epoch/
receipt guards and configured history limits stay intact. Two replay component
buffers and support-index maps are reused instead of allocated on every row.

The cached prefix may be up to63rows before the changed row, but every subsequent
row still propagates to the present. This is not a blanket64-row cost bound:
old evidence still requires O(T) replay, and four recent prefixes cover only a
recent region. Full-history publication encoding/hash and row copies remain.

## Audits

Frozen attempt `research/regime-incremental-v80-initial`: unit/race/vet/benchmark
all terminal exit0, closure/source/protected tracked hashes unchanged during run.

- Original108 configurations, 2,474 inspected publications and2,341 pending
  query checks retained; all independent dense/original-clock tests pass.
- 143,038 total scalar checks, maximum dense defect1.2435e-14 (tol2e-11).
- 10,364 V79 reference scalar checks: maximum disagreement ZERO.
- 214 independently cold-checked cached prefixes, including boundaries63/64/65,
  multiple reverse/old/latest reveals and explicit refresh.
- Cached/replayed normalizers, discarded mass, envelopes, component counts,
  forecasts and small joint laws checked, not point means alone.
- Full150/200-member16-round histories, old and recent pending branches checked
  against frozen V79; query leaves cached arrays unchanged.
- Unknown/future forks, stale/phase/value/clock errors, ownership, unsupported
  same-Y pairs and atomic uncapped64-row history limit preserved.

Maximum tower defect3.1087e-15 and issued-receipt defect1.3323e-15. Race maximum
dense defect1.2435e-14; independent-reference maximum remains zero. A single-owner
API passing race tests is not a concurrent-serving guarantee.

One mechanical patch initially mismatched gofmt field spacing and applied no
changes. The context was re-read and repaired before freeze/tests. No failed
scientific attempts were deleted or criteria changed.

## Matched cost

One serial Go benchmark command ran V79 reference then V80 candidate, both
prepared under equivalent as-of laws. Two repetitions, AppleM4; no randomized
paired trial, confidence interval, peak/RSS or loaded latency claim. Readback
verifies32 reference groups,34 candidate groups,132 records against raw output.

| Workload/operation | V79 | V80 |
|---|---:|---:|
| M150/T2400 latest first query |146.61-147.58ms|.849-.854ms|
| M200/T3200 latest first query |186.13-187.09ms|1.395-1.405ms|
| M150/T2400 accepted first reveal |74.98-75.01ms|5.99-6.06ms|
| M200/T3200 accepted first reveal |100.95-101.28ms|8.18-8.20ms|
| M150/T2400 issue |72.37-72.61ms|5.84-5.86ms|
| M200/T3200 issue |98.70-99.80ms|7.80-7.85ms|

Latest-query allocation falls from765.67MB to506,532-506,533bytes at150 and
1,020.80MB to542,368-542,369bytes at200, TOTAL per query, not peak/live/RSS.
Accepted publication operations still allocate about3.18-4.62MB: whole-history
JSON fingerprinting remains a cost. A future fingerprint rescue must retain
complete state/epoch binding, not merely omit it to make a benchmark pass.

Candidate OLDEST first queries take30.97-31.27ms at150 and41.18-41.47ms at200;
allocation506,528/542,368bytes. No separately timed old reference is supplied,
so report these as absolute costs, not an old-query matched speedup. Older
evidence is processed, not sacrificed for the recent-case result.

Full setup uses excluded cold support fitting plus cache reconstruction, not
full growing-history ingestion.32-row actual-Issue setup equivalence is checked.
These are public query/update/issue operations, but they do not include corpus
retrieval, acquisition policy, all round operations, queueing, persistence,
serialization outside the ledger or real agent serving.

## Memory and limits

AtM150/T2400, retained Key payload1,381,968bytes; atM200/T3200,1,842,768bytes.
Four cached component payloads total233,856bytes. These type-size counts exclude
slice headers, base/current vectors, replay scratch, temporary serialization,
heap/GC and RSS. Empty constructor allocation is still not the whole resource
bound. Maximum allowed cap256/T12800 support payload can still reach50MiB;
prefix caching does not fix that. No whole8MiB or allowed-maximum resource pass.

Replay now indexes support in expected O(K) per row rather than O(K squared)
linear matching for every key, and allocates reusable O(K*MaxMembers) buffers.
Copies/filters still cost O(L*K*MaxMembers) for suffix lengthL. Current-state
Issue has L=1; a query outside cached range has L=T. Prefix storage bounded by
four component snapshots. Publication still costs O(T*K+T+M) and temporary
serialization bytes, so Issue/Reveal are NOT asymptotically constant-time APIs.

## Decision

Keep the rescue as a computational research component; no change to scientific
claims. Recent and oldest SERIAL queries meet100ms in these fixtures, not a
loaded-serving guarantee or whole400ms-core gate. Original repeated-policy
work, quality and equal-TOTAL-cost comparisons remain required.

Next: remove whole-history fingerprint work with fully bound state identity,
measure whole growing-history/round costs and full retained/scratch resources;
separately protect plausible noisy explanations or abstain explicitly. Then
run the original varied generators/noise/shift/delay and frozen control gates.
Valid useful splits, untouched labeled agent tasks, durable loaded freshness
and equal-TOTAL-cost observation remain open. Production/private/sealed/reserved
data, dirty tracked files, whitepaper/publication stayed untouched; no pushes.
