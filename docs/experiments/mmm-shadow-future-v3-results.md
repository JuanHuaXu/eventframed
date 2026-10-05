# Future-arrival ablation: zero read errors, latency screen FAILED

All1536 reads and384 writes succeeded across three off/on pairs. All64 writes
per arm overlapped active readers. Unlike v2, new records were available strictly
after the frozen query cutoff. No acceptance/retry rule was changed.

| Trial | Off p99 | On p99 | Completed shadow jobs | Stale shadow jobs |
| --- | --- | --- | --- | --- |
| 0 | 29.291 ms | 29.379 ms | 95 | 160 |
| 1 | 28.264 ms | 31.985 ms | 63 | 192 |
| 2 | 29.424 ms | 31.064 ms | 39 | 217 |

Trial1 exceeds the frozen10% p99 allowance (about13.2% increase), so overall
FAILED. No rerun was used to erase this outcome. Cancellation accounting holds
after Close. The short timing sample does not establish a population tail bound.

## Deterministic cause isolation

New tests interpose ingestion after forecast construction and before journal
commit on both memory and embedded LibraVDB backends. Future-available writes
allow first-attempt commit;5 consecutive relevant backfills produce5 rejected
attempts;3 backfills followed by quiet permit attempt4. Returned packets contain
no future evidence. These tests pass under the race detector.

Thus v2's workload was sustained backfill into the query's visible history, not
ordinary future arrivals. The guard correctly rejects changed evidence. This
does not remove the availability requirement for legitimate backfill workloads;
simply accepting stale journals would break the invariant. A lifecycle-aware
admission or consistent-read mechanism needs separate design and testing.

Future-only writes still stale the diagnostic worker because its guard compares
entire snapshots, unlike journal commit's audited temporal-compatibility rule.
This explains unnecessary rejection opportunities, not proof every discarded
job was usable. A follow-up may reuse the existing temporal compatibility
contract with explicit as-of and complete version accounting, while retaining
strict rejection for posterior, graph, policy or unknown changes. Do not infer
that all version changes are harmless.

The workload holds visible candidates at50, whereas v2 grows them through
backfill. Absolute latency comparisons between v2/v3 are confounded; only
within-workload paired comparisons were screened. No production OpenClaw,
remote backend, generation model or true research learner was involved.
