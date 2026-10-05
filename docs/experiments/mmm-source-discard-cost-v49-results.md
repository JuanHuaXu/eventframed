# Source-key batch discard v49

PASS for the frozen isolated correctness/performance screen. Not a loaded
service, feedback-authority, learning-quality or whole-direction success.

## Mechanism and boundary

v48 measured cleanup as the largest source-owner phase. Inspection showed an
explicit source lookup and separate durable terminal call for every item; the
raw prepared control already supported an atomic terminal batch. This is an
API/transaction-granularity gap, not proof of faulty SQLite durability or a
reason to drop source checks. The alternative explanation, source resolution
cost alone, remains relevant to residual overhead and is not called solved.

v49 adds `SourceOwner.DiscardBatch` with at most 256 source keys. Under the
private owner lock it resolves every original first, then calls the existing
atomic snapshot-preflight discard batch. No new learner IDs, forecast computation,
labels or evidence clock advances. Sources retain their identity after terminal
cleanup. Missing/duplicate/invalid late members cannot partially discard earlier
ones. The single-item API and production defaults remain unchanged.

The comparison changes only discard granularity; both arms keep the source
index, per-source lookup and identical admission/retry/readback paths. Safety
tests are the falsifier for replacing N independent operations with one atomic
operation: late rejection or uncertain commit must never acknowledge a subset.

## Results

Twelve cells, 384 cycles, 48,000 originals and 48,000 terminals; all retry,
readback and replay checks passed. Go test completed in 10.15s, Go 1.27.1,
darwin/arm64, Apple M4, 10 CPUs/GOMAXPROCS. No excluded samples or trials.

| Size | Trial | Single mean cycle ms | Batch mean cycle ms | Reduction | Batch cycle p95 ms |
|---|---|---|---|---|---|
| 50 | 0 | 11.403 | 8.302 | 27.19% | 10.038 |
| 50 | 1 | 11.054 | 8.040 | 27.27% | 8.192 |
| 50 | 2 | 10.936 | 8.016 | 26.71% | 8.253 |
| 200 | 0 | 43.525 | 33.531 | 22.96% | 36.216 |
| 200 | 1 | 43.625 | 33.843 | 22.42% | 36.247 |
| 200 | 2 | 43.625 | 33.557 | 23.08% | 36.423 |

The predeclared >=20% whole-cycle mean reduction at size 200 passes in every
trial. Size-200 discard means fall from 20.67-20.68ms to 9.94-9.99ms; discard
p95 falls from 21.35-21.54ms to 12.28-12.53ms. Unchanged-phase means are not
bit-identical: batch-arm readback is 6.76-6.88ms versus 6.38-6.44ms, and admission
7.96-8.10ms versus 7.69-7.73ms. Whole-cycle reporting includes those differences.
This experiment does not prove their cause or provide population confidence bounds.

Size-200 closed database size remains identical at 7,979,008 bytes in both arms;
restarts span 97.68-99.72ms single and 98.01-99.69ms batch. Full lifetime replay
and on-disk index growth are unchanged. At size50 both files are 1,990,656 bytes.

## Verification

- `go test -race ./internal/researchmemory -run '^TestSource' -count=3`: PASS.
- New tests cover late missing/duplicate/early/zero/conflicting-terminal requests,
  cancellation and count caps without partial cleanup or stopping on preflight.
- Mixed new/exact terminal retries and reopen retain originals and invent no label.
- Injected acknowledgment failure before/after actual commit stops the owner;
  replay then returns the correct all-new or all-retry terminal result.
- Full ledger, learner and service race suites: PASS. Their `go vet`: PASS.
- Artifact source hashes, cell/sample counts and exact phase sums independently
  recomputed; screen evaluated on all three size-200 pairs.

Command: `EVENTFRAME_SOURCE_DISCARD_COST_ARTIFACT=.../mmm-source-discard-cost-v49.jsonl go test ./internal/researchmemory -run '^TestSourceDiscardCostStudy$' -count=1 -v`.
SHA-256: `423e0ab2724c1fabcb566563c197d7337ec9e93cdcac1c1d8fdee6ec7b9bbfdf`.

## Next evidence

Integrate the opt-in source owner into the existing guarded, durable loaded
fixture and compare with the retained raw prepared control. Preserve unique
journal/event identity, actual originals and typed cleanup; do not relabel a warm
preview to fit the new IDs. Apply the unchanged completion-age and serving-tail
screens. Per-source resolution remains a possible next bottleneck, not a reason
to assume loaded performance from these isolated timings. Feedback/history
authority remains an independent gap and no raw feedback/scoring API is exposed.
