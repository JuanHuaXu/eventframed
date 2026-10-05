# Admission phase diagnostic v65

## Finding

Fresh resolved admission spends71.84-74.96% of measured mean time in atomic
append,18.91-20.75% in source resolution plus durable preflight, and the rest
mostly in actual forecast/record staging and encoding. Moving pure preparation
alone cannot remove the dominant cost. Source/preflight totals are upper bounds
on potentially movable work, not proof of purity or attainable speedup: they
read owner-protected state and plan IDs. Durability and service guards remain.

Retries differ:67.23-71.28% is source/preflight and21.61-24.05% append. Do not
use retry proportions to explain the fresh-admission service workload.

## Evidence

[Protocol](mmm-admission-phases-v65-protocol.md),
[raw artifact](mmm-admission-phases-v65.jsonl).
SHA256 `907cf6ca30913a9a22c21b8712aa2fd9a357846890e7b811e24993138bfdace7`.
All31 captured source hashes were checked against embedded and local files.
Twelve cells,32 cycles each:48000 originals with exact retries and48000 verified
terminals, plus384 synthetic setup labels across warm cells. Reopen checks
validate lifecycle counts, endpoint originals and retained terminal retries.
Each warm setup waits for every training label to be processed before continuing.

Full researchmemory race suite PASS12.902s, vet PASS. Timing-enabled versus
disabled parity passes for cold and trained cases, with identical training and
original hashes. All768 measured fresh/retry phase tuples have positive durations
whose sum is bounded by corresponding outer call time. Experiment PASS4.045s
package time. Go1.27.1, Apple M4, darwin/arm64, GOMAXPROCS10; no competing
task-started tests or profiling during measurement.

Command: `EVENTFRAME_ADMISSION_PHASES_ARTIFACT=<LOCAL_ROOT>/docs/experiments/mmm-admission-phases-v65.jsonl go test ./internal/researchmemory -run '^TestResearchAdmissionPhasesExperiment$' -count=1 -v`.

## Phase ranges

Ranges across three trial means, in ms. Preflight column combines source
resolution and durable validation. Phase endpoints are not synchronized service
profiles and must not be added to v63 totals.

| Batch | State | Operation | Total | Preflight | Stage/encode | Atomic append |
| --- | --- | --- | --- | --- | --- | --- |
| 50 | cold | fresh | 1.042-1.232 | 0.197-0.234 | 0.063-0.082 | 0.778-0.915 |
| 50 | trained | fresh | 0.986-0.993 | 0.190-0.197 | 0.063-0.065 | 0.733-0.736 |
| 200 | cold | fresh | 3.569-3.646 | 0.691-0.737 | 0.265-0.305 | 2.597-2.632 |
| 200 | trained | fresh | 3.437-3.454 | 0.712-0.714 | 0.239-0.246 | 2.479-2.495 |
| 50 | cold | retry | 0.819-0.869 | 0.575-0.602 | 0.054-0.058 | 0.186-0.209 |
| 50 | trained | retry | 0.813-0.825 | 0.580-0.581 | 0.054-0.058 | 0.178-0.189 |
| 200 | cold | retry | 3.140-3.229 | 2.131-2.217 | 0.251-0.329 | 0.678-0.760 |
| 200 | trained | retry | 3.073-3.095 | 2.177-2.200 | 0.204-0.222 | 0.676-0.690 |

Timing is enabled by a private pointer installed only before exclusive test
owner calls. Constructors leave it nil; normal calls add conditional checks,
not clock calls or public instrumentation state. Clock and return overhead plus
owner-lock acquisition lie outside phase sums. Source resolution is measured
under the source mutex, durable validation/staging/append under the durable
mutex. Timing does not relax these locks or alter stopped-on-uncertainty behavior.
The retry Stage column includes re-encoding stored records, not new forecasting.

## Next lead

Investigate append statement count before designing a historical-certificate
contract. Current prepared append does an identity lookup followed by insertion
for every fresh row. An insert-first path with conflict handling scoped exactly
to the existing identity/kind key might remove one statement per new row while
retaining atomic rollback, exact retry bytes, source-identity uniqueness and
commit-before-ack. This is a hypothesis, not an implemented optimization or a
claim that SQL dominates fsync. Test late conflicts, feedback-before-admission,
mixed retries, uncertain commits and reopen before comparing fresh/retry cost.
Do not suppress unrelated uniqueness failures with broad INSERT OR IGNORE.

All seven research directions remain incomplete; no non-harm rescue, default
promotion, production change or publication follows from this diagnostic.
