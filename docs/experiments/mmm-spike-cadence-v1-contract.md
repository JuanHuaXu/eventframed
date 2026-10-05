# Frozen arrival-triggered cadence diagnostic

Keep the existing single ascending-order model, pi=1/255, unit slab/intercept,
255 masks,64-label window,1,024 iterations and mixture integration unchanged.
Use the84 index0 source records, predict clocks128 through159. At each clock,
admit only earlier nonmissing events whose arrival is <=clock. Retain the
latest64 by event origin, as in the existing control. Refit from the same
initialization iff that retained origin set changed. An arrival outside the
retained window does not require a refit. Always fit at128. No warm starts,
prior adaptation, order ensemble or evaluator-informed selection.

Emit each query forecast after processing its available earlier evidence and
before reading its outcome. Delay0 is still not permission to use the current
event outcome: require origin<clock. Save every fitted state, its origins,
publication clock, bound/motion trace and per-query fit index/integration
accounting. Report exact fit counts and total collection time. Fail explicitly
on model errors; no skipped records or fallback.

Compare with the audited same-model32-frame frozen pilot and existing generic64,
Boolean64 and Markov controls on identical queries. All four phase/schedule
cells and per-case scores remain visible. Require prefix poisoning tests,
no-new-evidence reuse, exact initial-fit agreement, source/moment audit and
byte-identical replay. This is consumed exploration on one index, NOT new
confirmation or a full32-index gate. Immediate refit publication is an ideal
cadence diagnostic; all its compute is charged, but it is not a working async
serving or equal-compute benchmark. No claim of real-world agent utility.
