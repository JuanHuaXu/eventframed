# Unchanged source-owner profile v51

The profiled repeat corroborates the unresolved load deficit, not a rescue.
Source hash maps match v50 exactly. Workload, source/control selection, guards,
identity checks and thresholds were unchanged. Profiles add instrumentation;
this is not an uninstrumented confirmation or a population latency estimate.

## Load result

All nine cells pass integrity/accounting. Source completes 404/576 observations
versus 576/576 for the raw prepared control, with 172 source queue drops and no
unexpected errors. Source age p95 is 554.222, 564.338 and 565.768ms, failing
250ms in every trial. Raw-control ages are 204.767, 195.363 and 189.222ms.

Source read p99 is 24.757, 24.396 and 27.943ms against off 33.910, 33.170 and
33.908ms: all three finite read screens pass. Source write p99 is 28.830,
33.842 and 33.860ms versus off 18.732, 17.895 and 19.012ms, so writer overhead
persists. All 20,200 accepted source and 28,800 raw-control originals received
complete readback and durable unlabeled cleanup. No feedback or learning occurred.

## Profile evidence

CPU profile: 15.76s wall duration, 22.88s aggregate CPU samples across all arms
and threads. The source lookup stack accounts for 500ms cumulative samples;
460ms lies under SQLite step/shared-memory locking through `FcntlFlock`/fcntl.
The source call's `QueryContext` line accounts for 250ms and its final `Rows.Next`
for 240ms. These are nested cumulative measurements, not additive totals.

The profile does NOT establish statement parsing as the dominant cost. It points
more directly to repeatedly entering/exiting SQLite read transactions for point
lookups. Low global CPU percentage does not by itself rule out significance to
the serialized background consumer, and samples do not measure all wall waits.

Allocation profile: 5,820.89MB estimated allocated space over the full run,
316.62MB under source-owner lookup. Of that nested total, GetServiceAdmission
accounts for 162.54MB and canonical record decoding for 143.58MB. These are
sampled cumulative allocations, not retained memory or a leak finding. Decoding
must remain semantically strict; this is not permission to omit verification.

v46's EXPLAIN test already checks a five-field indexed SEARCH rather than a
history scan. Inspection of the current driver also confirms that explicitly
prepared statements can retain native SQLite handles. Together these observations
warrant a bounded transactional source-read batch: one snapshot/transaction with
reused query machinery, explicit missing/error distinction, caller-order results,
per-record and total byte caps, and unchanged canonical ownership checks.

The decisive next test compares that primitive against point reads with identical
results and includes cancellation, late failure, corrupt payloads, concurrent
snapshot boundaries and all identity fields. No expected speedup is claimed in
advance, and loaded integration must still satisfy the original screens.

## Reproduction

`go test ./internal/service -run '^TestResearchSourceLoadExperiment$' -count=1 -v -cpuprofile=.../cpu.pprof -memprofile=.../mem.pprof -o .../service.test`
with `EVENTFRAME_SOURCE_LOAD_ARTIFACT` pointing to the exclusive v51 JSONL.
Test completed in 15.58s; no other tests from this task ran concurrently.
Source snapshots and hashes are in `mmm-source-profile-v51.jsonl`.

Artifact SHA-256:
`13c28579d4a2bcc5d08920f05486fcb3ace49108cac0f27788b9cca94e93e855`.

Temporary profile directory: `/tmp/eventframed-source-profile-v51.UsyXj5`.
Profiles are local diagnostic files, not committed research data; temporary-file
retention is not guaranteed. Their hashes are:

- CPU: `76e990726badd79c09a37480ad8940028051efb73b35b2e422206d3c7fc901ae`
- Allocations: `a4dcf23dba50b13c16de41a766f802754335b0f4157d5c676a0c57cc74ebecb2`
- Test executable: `78c3558cecad325d2a2e87d9b2ce01170c89c7c5f5e74b910899acb8e7615fb9`

Commands used `go tool pprof -top -cum` and `-alloc_space` focused on
`GetServiceAdmission|researchmemory.*SourceOwner.*lookup`, plus line-level
`-list=GetServiceAdmission`. Code, default APIs and all research-direction
completion statuses are unchanged by this profiling run.
