# Exact-guard load v19: zero useful admissions

Nine non-race arms completed with fresh persistent stores: three rotated trials
of off, validation-only and durable admission. Each used192 recalls/four readers
and96 future-dated writes. Runtime: Go1.27.1, darwin/arm64,10 CPUs/GOMAXPROCS10.
The raw JSONL embeds selected sources, protocol and dependency hashes; the
verified summary links its exact SHA256.

## Availability failure

Across all six enabled arms,1,152/1,152 observations were rejected as guard busy.
There were zero accepted observations, zero candidate validations, zero durable
admissions/discards, zero stale rejections and zero queue drops. Reads/writes
completed without errors and overlap was observed in every arm.

This is a substantive failure of the tested immediate-try-lock handoff under
this workload. It does NOT measure ledger throughput, callback critical-section
cost, learning latency or model fitting. Each accepted observation was designed
to validate all50 candidates, but none entered that path. No apparent low
latency in the durable-labeled arm can be attributed to efficient durable work.

## Request latency

| Trial | Off Recall p99 ms | Validation p99 ms | Durable-labeled p99 ms |
| --- | ---: | ---: | ---: |
| 0 | 31.123 | 40.112 | 30.994 |
| 1 | 33.028 | 30.889 | 34.128 |
| 2 | 32.020 | 31.086 | 35.657 |

Validation/off p99 ratios were1.289,0.935,0.971; durable-labeled/off ratios were
0.996,1.033,1.114. No latency pass gate was declared. Enabled write p99 ranged
16.951-22.729ms. Busy-check p95 was0.0005-0.001417ms, measuring rejection only.
Accepted-observation age is unavailable, correctly represented as null rather
than0. These results cannot validate the full learning path or its staleness.

## Verification

The8-recall accounting fixture passed three race repetitions before the measured
run. Service vet passed. The experiment test completed in15.566s and passed its
accounting assertions; that Go PASS is NOT an availability or research-success
verdict. There was intentionally no test assertion redefining rejection as a
successful outcome.

`python3 research/verify_guard_load_v19.py` independently checked nine-arm
coverage, embedded source hashes, read/write counts and overlap, queue/attempt
conservation, candidate/ledger accounting and timing completeness. It wrote
`mmm-guard-load-v19-summary.json` exclusively. Prior v18 controlled tests remain
the evidence that actual guarded callbacks and ledger reopen work when entered.

## Next investigation

Distinguish sustained writer contention, read-triggered mutations and phase
alignment after journal reads before selecting a fix. The adapter holds its
own writer mutex through backend work and publication completion; immediate
TryLock does not promise fairness or eventual admission. A bounded queued guard
may help entry, but exact-snapshot rejection can remain after waiting, as v18
already demonstrated for compatible future ingestion. An as-of-compatible guard
must validate under the held lock, not weaken freshness by checking before it.

No lock, freshness rule, service code or production configuration was changed
to make this run pass. Full feedback authority, history validity and actual
learning under persistent load remain open. Preserve this zero-admission result
as a negative control for any subsequent rescue.
