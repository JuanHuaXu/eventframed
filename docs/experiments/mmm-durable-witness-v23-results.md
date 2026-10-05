# Durable As-of Witness V23 Results

2026-10-03. **Overall adoption FAIL in all eight normal trials. All seven
whole research goals OPEN.** The new conditional transport component restores
some served learning across future-only insertions, but it does not meet the
unchanged background outcome-latency requirement. Production untouched.

## Scope And Frozen Protocol

[Protocol](mmm-durable-witness-v23-protocol.md). Both arms durably log the same
kind of mutation/query/source witnesses; only the enabled arm may return
request-local epoch-aligned copies. Stored posterior/certificate epochs remain
unchanged. The complete v20 dependency guard, actual request/vector/frontier
binding, native LSN/marker gate and as-of source checks remain load-bearing.
No visible-insert score/graph bound is supplied, so visible insertions stay
Unknown. Synthetic external-coverage certificates are assumed, not proven.

Eight normal trials: two repetitions x future/visible writes x control/enabled.
Each starts with200 eligible normalized256D vectors and17 future records.
128 event writes and128 full Recalls are independently offered every4ms;
eight Recall workers. Event batches<=16/16ms dwell; journal batches<=4/1ms.
16 distinct full-stream outcomes alternate useful/not-useful and are offered
every16ms to one background worker. Serving AsOf=offer time. This is an
isolated libravdb/SQLite test binary, not production, HTTP/OpenClaw, real agent
answer quality or a corpus-scaling benchmark. Actual offer gaps are retained.

Normal workload completes1024 writes,1024 measured Recalls plus8 primes,
and128 outcomes.153600 measured candidate forecast decisions are retained.
Each trial commits157 witness publications accounting for144 runtime mutations
(128 inserts/16 outcomes), plus journal bindings with no runtime motion.
1122 source/protocol/collector/checker files were frozen before collection.
Go runner PASS means collection completed, not that adoption gates passed.

## Normal Measurements

All p99 values below are milliseconds. Every normal trial FAILS because outcome
offered->durably-published p99/max exceeds100/250ms. Read/write/view gates pass.

| Repetition | Writes | Transport | Learned decisions | Cross-epoch uses | Recall call p99 | Recall offer p99 | Write offer p99 | Outcome offer p99/max | Result |
|---|---|---|---:|---:|---:|---:|---:|---:|---|
|1|future|off|0|0|53.490|82.476|118.247|490.931|FAIL|
|1|future|on|310|270|54.781|84.581|110.144|501.863|FAIL|
|1|visible|off|0|0|50.250|71.417|104.038|478.650|FAIL|
|1|visible|on|0|0|53.562|78.266|108.212|483.806|FAIL|
|2|future|off|0|0|46.707|97.887|111.632|508.271|FAIL|
|2|future|on|299|260|51.286|91.189|116.359|502.637|FAIL|
|2|visible|off|0|0|50.285|96.945|112.377|504.434|FAIL|
|2|visible|on|0|0|51.615|90.359|110.892|497.824|FAIL|

Maximum observed published-view age across all normal trials39.162ms, below
250ms. These are two scheduling repetitions, not independent-world quality CIs.
Neither one recorded outcome per member nor this synthetic workload establishes
calibration, risk improvement, robust recovery or statistical error control.

## Learning And Censoring

Enabled future-only mode serves609 learned decisions in total;530 genuinely
cross the stored source epoch. The ordinary single-trial Beta posterior
predictives match the served belief law exactly. Both disabled future controls
serve zero learned decisions. Both enabled visible controls transport zero:
changing the eligible frontier or an unbounded visible mutation stays blocked.
No forecast uses a source available after its request's as-of time.

Only5 of16 outcome sources reach a later observed scored forecast in each
enabled future trial. Label availability->first observed scored use delays:

| Source order | Repetition1 ms | Repetition2 ms |
|---|---:|---:|
|0|97.585|105.873|
|1|190.630|190.332|
|2|276.436|283.038|
|3|352.365|364.319|
|4|446.244|457.737|

The remaining11 per trial are explicitly censored within the128-Recall window,
not called instantaneous or dropped from the denominator. They were ultimately
acknowledged but no later serving request was collected for them. Thus restored
law wiring is real while freshness remains inadequate.

## Audit And Collector Corrections

Six lifecycle subtests pass under race: future/visible controls, bypass and
backend-before-witness interruption, corrupted binding and corrupted chain.
Actual database close/reopen preserves accepted transport and rejects gaps.
Changed query, normalized vector and selection settings block reuse; old-as-of
requests never see later outcome evidence. Stored epochs remain unchanged.
Core guard race tests and package vet pass. Controls cover injected boundaries,
not power loss, an external malicious writer or general hierarchical closure.

The sealed collector reports128 `durable journal mismatch` flags per trial.
**Confirmed collector defect, not a demonstrated journal storage failure:**
`reflect.DeepEqual` includes private `EvidenceGroupKey`, which is deliberately
`json:"-"`; decoded journals omit it. A separate post-hoc race-tested falsifier
reproduces the flag while all150 wire decisions and snapshots match exactly.
The frozen native journal gate already requires byte-equivalent persisted
readback before ack. Original flags, checker, gates and tape are unchanged.

The original checker also assumes UTC `Z` for published-view timestamps, while
Go's `time.Now()` view uses RFC3339 `-04:00`. The supplement normalizes only that
representation in an analysis-local copy, preserving the fractional timestamp
and instant; raw data remain unchanged. The supplement separates only the
known private-field flags and keeps every original failed adoption outcome.
It does not pretend to reopen/replay the discarded normal-cohort databases.

[Supplement](../../research/durable-witness-v23/supplement.json) independently
checks source hashes, complete commit/state checksums, runtime/epoch motion,
exact top150 nomination oracle, request pins, as-of eligibility, ordinary Beta
arithmetic, served/packed laws, stored-source identity, freshness censoring,
actual timings and all gates. Eight corrupted-tape controls reject missing
writes, false future nominees, changed pins, future labels, damaged chain,
retagged sources, invented timing and fabricated passes. Original checker
failure is retained separately. These are technical conditional checks, not
new coverage certification.

## Performance Attribution

A separately labeled instrumented replay records CPU/mutex/block profiles;
its timing is **not** a new confirmation or normal gate measurement. It repeats
the workload to attribute blocking, preserving separate artifacts. Go pprof
reports aggregate concurrent wait, not additive wall time or CPU work.

`ObserveBayesianOutcome` has4.37s aggregate blocking;4.35s occurs at
`witnessStoreV23.ApplyBayesianOutcome`'s admission write-lock. Recall admission
read-lock waits sum12.76s. Native journal queue/ack waits sum12.05s, followed by
4.17s owner-lock waiting while publishing per-journal witnesses. CPU also shows
native HNSW work and substantial runtime thread/lock activity. A cached source
journal alone would therefore not fix the measured outcome bottleneck.

Next prospective rescue: explicit bounded mutation/read admission scheduling
that services waiting outcomes fairly relative to ingestion, with read cohorts
and no admission bypass. Also investigate integrating witness commits into the
existing journal batch owner, avoiding a second serialized publication queue.
Both require new freeze/tests/cohorts; neither is yet implemented or validated.
Preserve durable-before-ack, source/as-of bindings, visible blocking, original
offer cadences, all work counts and every100/250ms gate. Do not drop inconvenient
labels, throttle the offered load or release the lease without a complete
immutable law/dependency snapshot.

## Evidence

Artifacts: `research/durable-witness-v23/` contains source freeze, run metadata,
normal raw NDJSON/log, original checker failure, supplement, lifecycle tests,
separate profile raw/binary/CPU/mutex/block/top lists and boundary attribution.
Collector/core/checker remain sealed; only a new diagnostic and supplemental
analysis were added after normal collection.

Normal raw SHA256:
`957ba698f2141c0f02d57cafcb1b01c96750efbd3f247accd83681bc9133a37c`.
Supplement SHA256:
`300dd0939ff9bbecc93eb6abe1de5b8a2c321fa205d2bd53f6f22853ca7e10a6`.
Profile boundary artifact SHA256:
`1496037a5b9d5922bd04874fdc5c18d8b2fa8e43e658a6ae170d6f69bf91cb59`.
No existing frozen study, runtime source, whitepaper, installed tooling or
private data changed; no commit or push. Full Goal6 and all other goals OPEN.
