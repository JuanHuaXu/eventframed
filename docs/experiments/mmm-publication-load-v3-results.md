# Ingestion-only publication rescue v3

**Queue16 PASSED all six finite cells; queue64 FAILED one age gate.** This is a
test-only ingestion adapter, not completed general store integration.

See [protocol](mmm-publication-load-v3-protocol.md) and
[18-arm raw artifact](mmm-publication-load-v3.jsonl). Same request/write counts,
labels, queue sizes and completion/age/serving gates as v2. The new adapter
announces Put intent and publishes the validated version without making the
research bridge wait first on the ordinary Snapshot read lock.

| Requests | Trial | Queue16 completed | Age p95 ms | Queue64 completed | Age p95 ms |
| --- | ---: | ---: | ---: | ---: | ---: |
| 64 | 0 | 57/64 | 96.027 | 64/64 | 102.189 |
| 64 | 1 | 59/64 | 96.815 | 64/64 | 106.123 |
| 64 | 2 | 61/64 | 91.444 | 64/64 | 114.278 |
| 192 | 0 | 155/192 | 109.353 | 192/192 | 236.046 |
| 192 | 1 | 154/192 | 116.126 | 192/192 | 228.348 |
| 192 | 2 | 155/192 | 110.539 | 192/192 | **252.417** |

All arms had zero reported errors and overlapping writes. Every admitted
frontier's50 labels completed. Queue16 serving-p99 ratios were .699,.732,.758
for short workloads and1.005,1.052,1.036 for longer ones, all within1.10.
Long-workload completion is80.21%-80.73%, a narrow margin above80%, not a robust
population claim. Queue64 completed all requests but its252.417ms age exceeded
250ms; the overall test correctly exits failed. Do not round away that failure.

The comparison with v2 supports snapshot-lock waiting as the architectural
bottleneck in this fixture. It is not a simultaneous paired comparison against
the old wrapper within each trial; scheduling/hardware noise remains possible.
Fresh workload replication is needed before claiming a stable minimum benefit.

## Verification

Before load: three race-test repetitions covered pending future versus visible
writes, successful versus ambiguous outcomes, model quarantine and proof routing
without an ordinary Snapshot call. Standalone publication-model race tests also
passed. After load, source hashes, all18 arm/sample counts, nearest-rank p99/p95,
label/frontier conservation and each gate were recomputed from the artifact.
Only (192 requests, trial2, queue64) fails the recorded finite screen. Vet passes.

## Scope and next requirement

The adapter exists only in a test file and is installed after initial binding
and seeding. Measurement uses Put and non-version-changing journal writes. The
embedded store has other mutation methods that are NOT wrapped, so reusing this
adapter as a general runtime store would be unsafe. No deployment/push occurred.

The bridge's optional coherent-proof hook is inert for ordinary stores. Full
integration still requires every mutation path, recovery, in-progress visibility,
and ambiguous commit outcome to enter the publication protocol. This finite
ingestion success does not complete roadmap6 or validate real-task learning.
Queue16 is a candidate for replication, not a production recommendation.
