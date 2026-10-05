# Batched API profile v34

Diagnostic only. Repeated all twelve v33 arms with CPU/allocation profiling;
embedded source hashes exactly match v33. No tuning or source change intervened.
The instrumented batched arms admitted 134/137/133 observations, still dropping
172 of 576 overall. Individual durable arms admitted 77/77/78. These are not new
untouched-task or uninstrumented latency confirmations.

CPU sampling found 0.60 seconds under AdmitBatch/DiscardBatch, of 29.89 total
CPU-seconds for the experiment. AppendBatch accounted for 0.47 seconds, COMMIT
about 0.22, and transaction Exec about 0.20 (nested cumulative values, not additive
independent phases). Locking and page writes were prominent: FcntlFlock and
Pwrite each appeared under about 0.16 seconds. This does NOT isolate wall-clock
disk latency, and small sample totals warrant caution.

Allocation sampling estimated 188.41MB cumulatively allocated under the batch
APIs, including 94.01MB under AppendBatch and 59MB under Get. These cumulative
values overlap and are not resident/peak RAM. Total experiment allocation was
6701.16MB. The profile does not establish statement parsing as the dominant cost.

Source inspection shows per-record preflight and integrity reads still call Get
separately, each using its own database/sql query operation. Thus the next
discriminating candidate is bounded, ordered multi-key reads within one explicit
read transaction, preserving found/missing distinctions and payload integrity.
Do not replace this evidence with an unsupported prepared-statement speedup claim.
Measure that candidate against the current API before adopting it.

Any read-batch design must bound requested identities and aggregate returned
bytes, preserve caller order and exact tenant/stream/contract keys, and reject
cancellation/partial read errors without inventing missing records. It must not
cache service authority across transactions or weaken FULL write durability.
Wrapper ownership remains responsible for excluding conflicting logical writes.

## Reproduction and artifacts

The command reran `TestResearchTransactionsLoadExperiment` with `-cpuprofile`,
`-memprofile` and a temporary `-o` binary. Local profiler files are under
`/tmp/eventframed-profile-v34.kK7i7U`; they can expire and are not published data.
Analysis used `go tool pprof -top -cum '-focus=AdmitBatch|DiscardBatch'` and
the corresponding `-alloc_space` view.

- JSONL: `mmm-transactions-profile-v34.jsonl`
- JSONL SHA-256: `b683f9f1b2edd65785fd880840147968a5982d54bc19817652779f3685416bf6`
- CPU SHA-256: `f9041ff1fbcc56a5908bea3ac60a3e8795c0adf1768146c984cb7443abcdebb3`
- Allocation SHA-256: `b4145a694b9379482d2a13d13a94a70c8b505f1389eb97960d4fbf9d50470e6b`

Independent parsing verified all twelve unique cells, source-hash parity, raw
request/write counts, outcome conservation and actual durable operation counts.
Full service/learner/ledger race tests and vet passed before profiling. No
production configuration, deployment, paper publication or push occurred.
