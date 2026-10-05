# Transactional source reads v52

PASS for the frozen primitive-level correctness/performance screen. The source
owner and loaded service still use point reads. v50/v51's loaded age failure is
not resolved by this isolated result, and all research directions remain open.

## Reasoning and implementation

v51 sampled repeated SQLite read-lock operations under point source lookup.
Plausible costs included transaction lock churn, query setup and decoding; the
profile did not establish statement parsing as dominant. No production data
repair or upstream bug claim was involved: this is a local experimental read API.

`GetServiceAdmissions` opens one read-only transaction and prepares the same
five-field indexed lookup for repeated execution. It returns caller-order results
bound to each requested Source. Misses have Found=false and zero Entry; any late
read, validation, cancellation or commit failure returns no result slice.
Duplicate keys are allowed and consume count/byte budgets independently.

The batch accepts at most512 keys; encoded request bytes and returned payload
bytes each cap at8MiB, individual payloads at1MiB, and each identity envelope at
128KiB. SQL checks payload size against the remaining budget before materializing
it. Point and batch reads share the existing source/envelope validation helper.
The original point API, unique index, writes, source-owner defaults and service
authority boundaries are unchanged. Canonical forecast decoding is still the
consumer's separate responsibility; these storage fixtures are not forecasts.

A batch snapshot intentionally differs from a sequence of independent point
reads under a concurrent writer: the batch cannot mix before/after states. The
test inserts a row after the first read and checks it appears only in a subsequent
transaction. This is snapshot consistency, not fresh service authorization.

## Verification

- Source identity and batch-read race tests passed three repetitions.
- All five key components, complete indexed SEARCH, point/batch parity, ordered
  misses, detached duplicate payloads, and concurrent snapshot boundaries passed.
- Empty/oversized/invalid keys, 512-key boundary, encoded request cap, exact8MiB
  payload cap, zero-budget miss, and aggregate overflow passed.
- Cancellation/panic cleanup, oversized/text/binding/identity corruption, missing
  index and ambiguous results reject without a partial success slice.
- Full ledger, learner and service race suites passed; their `go vet` passed.

## Measurement

Twenty-four fresh-ledger cells, each with1,000 bound 1,024-byte JSON storage
fixtures and32 lookup groups. All48,000 hits and48,000 misses match expected
identity/payload or explicit absence. No excluded samples or trials. Test took
2.12s on Go1.27.1 darwin/arm64, Apple M4, 10CPUs/GOMAXPROCS. Full source hashes,
group counts and byte accounting were independently recomputed from JSONL.

| Size | Case | Trial | Point mean ms | Batch mean ms | Reduction | Point/batch p95 ms |
|---|---|---|---|---|---|---|
| 200 | Hit | 0 | 5.916 | 1.370 | 76.84% | 6.181 / 1.567 |
| 200 | Hit | 1 | 5.921 | 1.305 | 77.96% | 6.282 / 1.591 |
| 200 | Hit | 2 | 5.908 | 1.350 | 77.15% | 6.210 / 1.590 |
| 200 | Miss | 0 | 4.980 | 0.583 | 88.29% | 5.224 / 0.815 |
| 200 | Miss | 1 | 4.976 | 0.595 | 88.05% | 5.288 / 0.781 |
| 200 | Miss | 2 | 5.001 | 0.576 | 88.49% | 5.268 / 0.788 |
| 50 | Hit | 0 | 1.487 | 0.326 | 78.07% | 1.854 / 0.462 |
| 50 | Hit | 1 | 1.420 | 0.324 | 77.17% | 1.540 / 0.439 |
| 50 | Hit | 2 | 1.419 | 0.323 | 77.26% | 1.559 / 0.422 |
| 50 | Miss | 0 | 1.172 | 0.166 | 85.80% | 1.261 / 0.337 |
| 50 | Miss | 1 | 1.184 | 0.167 | 85.85% | 1.341 / 0.283 |
| 50 | Miss | 2 | 1.188 | 0.169 | 85.79% | 1.292 / 0.300 |

Every size200 trial/case exceeds the frozen20% mean improvement criterion.
This bundles transaction amortization and statement reuse; it does not separately
identify each contribution or imply a77-88% whole-service speedup. No billion-row
scaling, warm posterior activity, real-agent benefit or statistical population
guarantee is established.

Next integrate through an opt-in source-owner variant for admission resolution,
original readback and discard resolution. Keep canonical model/seed/epoch checks,
source retry identity, stopped-on-uncertainty behavior and the point-read control.
Then repeat the loaded fixture under the unchanged250ms age and serving-tail
screens. Do not skip the owner/lifecycle layer because the ledger primitive passes.

Command: `EVENTFRAME_SERVICE_READ_COST_ARTIFACT=.../mmm-source-reads-v52.jsonl go test ./internal/researchledger -run '^TestServiceReadCostStudy$' -count=1 -v`.
Artifact SHA-256:
`1a7c84388b9b69fb45250782b096218abbadb23a2f38c25058fe333392444ed3`.
