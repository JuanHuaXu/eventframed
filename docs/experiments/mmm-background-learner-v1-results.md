# Actual background learner: lifecycle study v1

2026-09-12. Research-only, not wired into service or served ranking. This closes
part of roadmap 6's diagnostic-versus-learner gap, not its full integration or
loaded-latency success criterion. No new accuracy claim or research rescue.

## Implemented boundary

`internal/researchmemory/background.go` owns the existing retained count/forest
adapter in one worker. Prediction uses an immutable snapshot, journals its
original experts, and never takes the fitting mutex. Verified feedback enters a
bounded queue; the worker refits and atomically publishes a new research model.
Mixture updates use the ORIGINAL pre-feedback experts, not forecasts recomputed
after intervening labels. Batches need not wait for each feedback update.

Pending journals cap at 256; queue capacity is 1..256. Saturation rejects without
destroying retryable evidence. Availability times must be nondecreasing for
accepted feedback; duplicate, early and wrong-epoch labels reject. Explicit
discard handles missing labels without learning a negative. A fresh epoch needs
a new worker. Close abandons remaining work, prevents subsequent publication,
and waits for the current bounded fit. It is not a hard-deadline cancellation API.

The old synchronous adapter and service diagnostic queue are unchanged. This
new worker has no text input, source-authentication authority, persistence,
Anti-Pigeon sharing authority, or automatic feedback inference. Unpublished
research snapshots are not permissions to change served forecasts.

## Verification

- Batched synchronous parity over 128 explicit labels: identical prediction
  journals, then identical scores for all 512 feature vectors after each batch.
- Old snapshots remain immutable across refits.
- Wrong epoch, early/repeated labels, bounded pending state, missing-label
  discard, invalid capacity, prediction after close and idempotent close.
- Deterministic queue pressure while the worker fit lock is held: rejection
  retains the original journal and retry successfully learns it once.
- Snapshot readers execute concurrently with actual refits under the race
  detector. No fitting mutex is used by those readers.

Commands: `go test -race ./internal/researchmemory -count=3` and
`go vet ./internal/researchmemory`. These are module/lifecycle tests, not an
isolated OpenClaw end-to-end run. No private data or production access was used.

## Isolated timing

Apple M4, darwin/arm64; three repetitions of
`go test ./internal/researchmemory -run '^$' -bench BenchmarkBackgroundFrozenScore -benchmem -count=3`:

| Repetition | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| 1 | 34.41 | 0 | 0 |
| 2 | 34.48 | 0 | 0 |
| 3 | 34.49 | 0 | 0 |

This measures snapshot load plus one score against a trained immutable model.
It excludes prediction journaling, feature extraction, embedding, queue admission,
refits, persistence and concurrent serving load. It does not establish a p99
budget or compare loaded off/on service latency.

## Still required

The current nine lexical features previously failed transfer against semantic
retrieval; asynchronous execution does not repair that. A defensible real-field
representation, verified feedback bridge, dependency/staleness checks at service
handoff, and persistent concurrent load with this actual learner remain open.
Do not substitute this lifecycle pass for those accuracy and integration tests.
