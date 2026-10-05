# Task lexical admission: fixed-arrival screen

Frozen before dispatch. No parameter selection from these outcomes.

Use actual Recall/CaptureTurn with task-lexical-overlay-v1 and the reusable
researchadmission gate; admission is at runner call boundaries, not daemon API.
Fresh memory store, hash32 embedder, 200 repeated public landing facts, recall200,
pack10, adaptive/diversity on, four read permits/exclusive writes. Original score,
law and store freshness rules remain unchanged. No LLM calls or private data.

Per arm: two warmups then 32 reads and 16 writes. Read arrivals are independently
scheduled every20ms or5ms (50/s or200/s). Write arrivals occur at odd multiples
of that interval (25/s or100/s). All48 timers are created before execution; no
completion schedules another arrival. Each deadline is scheduled arrival+100ms.
Each writer uses a distinct session so out-of-order admission is not confused
with sequence ordering. A finite48-task burst is not an unbounded queue design.

Two repetitions, future/visible writes, admission off/on, original/experimental
packing, both intervals: 32arms,1024reads,512writes. Reverse admission and packing
order on repeat2. All mutations within these arms participate when admission on.
Future stamps exceed request AsOf; visible stamps equal seed time.

Measure dispatch lag, pre-callback waiting and total duration from scheduled
arrival (including queueing). Record callback entry, all errors and stale journal
rejections. Validate successful packet/journal agreement and frontier availability.
Nominated may be200..216 under visible writes; future mode requires200. Missing
callbacks count as failures if canceled, never silently omitted. No retry-count
or freshness relaxation is allowed. Readback checks are outside service duration
but within reader wall completion; expiry there is a reported validation failure.

Finite admission rescue passes per condition only if all32 reads+16 writes
succeed, no stale rejection or future leak occurs, all returned journals match,
and each operation finishes within100ms of scheduled arrival. Report failed
conditions and writer waiting, not just successful latency quantiles. This strict
screen is descriptive, not a population tail bound. Do not infer sustainable
throughput from32 arrivals. Durable backend and background mutation coverage are
still absent. Preserve raw output and hashes, including unsuccessful arms.
