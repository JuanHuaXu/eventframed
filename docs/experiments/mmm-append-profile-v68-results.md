# Prepared admission SQL attribution v68

## Finding

The profile does not support SQL compilation as the dominant next target.
Commit and shared-memory locking have substantial sampled activity. An opt-in
SQLite exclusive-locking experiment is a better-supported next lead for this
already exclusively owned research ledger. No locking or durability setting
was changed in v68, and the failed speed/non-harm screens remain failed.

## Evidence

[Protocol](mmm-append-profile-v68-protocol.md),
[profiled phase artifact](mmm-append-profile-v68.jsonl).
Artifact SHA256:
`a0855a82171e069d4149d61e3b0d48974af03114046dbe09f2538183d8663dd8`.
All34 captured source hashes were checked against embedded and local files.
The protocol is not embedded by the existing runner; its separate SHA256 is
`cd5f6c3495944c057d83030e42bf649f0b824b7d2d40e637cba2ea34acff7dd3`.
Twelve unchanged v65 cells,48000 originals with exact retries and48000 terminals,
384 synthetic setup labels; original/lifecycle/reopen and timing checks pass.
Experiment PASS4.235s package time. No code changed for this profile, so prior
race/vet verification of the prepared measured path remains applicable.

Profiles under `/tmp/eventframed-append-profile-v68.F9HDu6/` are local diagnostics,
not durable repository assets:

| File | SHA256 |
| --- | --- |
| cpu.pprof | da61fc4222852c74dbd29dc428b78de390242c42ef4bdf48b85bc7b580ad7947 |
| block.pprof | 696ffeac6246c0e56cb67c25fc052f1b4d8ea66893f439a96dadcfaaa6679420 |
| memory.test | 1bba55f98a872730bf512fdff99e3329cd035963dfd0bb979f4d1b7cf454b36f |

Command used `EVENTFRAME_ADMISSION_PHASES_ARTIFACT` pointing at the artifact
above, `go test ./internal/researchmemory -run
'^TestResearchAdmissionPhasesExperiment$' -count=1 -v`, with `-cpuprofile`,
`-blockprofile`, `-blockprofilerate=1` and `-o` pointing at those files.
Go1.27.1,Apple M4,darwin/arm64,GOMAXPROCS10; no competing task-started tests.

## Attribution and limits

CPU profile contains4.79s samples over3.91s wall time. Aggregate transaction
Commit stacks account for0.95s; nested shared-memory locking/FcntlFlock stacks
in the inspected commit-related paths account for0.67s. Prepared append paths
account for0.61s cumulative sampled time. These are overlapping, different
scopes, not values to add or divide into a fresh-only fsync percentage.

The filter `sqlite3Prepare|sqlite3_prepare|sqlite3LockAndPrepare` matched no
samples. That means no observed samples in this short profile, not zero
compilation cost. Raw syscall names alone do not prove fsync dominance.
Block profile totals13.29s of cumulative waiting, largely select/channel waits
in SQL connection/transaction helpers and the background learner. Much of that
is idle infrastructure, not latency to eliminate.

Profiles include fresh, retry, terminal cleanup, setup and replay. They support
a locking lead but cannot establish how much fresh loaded admission time an
exclusive mode would save. v65 remains the isolated fresh/retry phase analysis;
v62/v64 remain the failed mixed-service screens.

## Next experiment contract

SQLite documents reduced filesystem calls with exclusive locking, and supports
WAL without shared-memory operations if exclusive mode precedes initial WAL
access. It also prevents access by other database connections. Sources:
[locking_mode](https://www.sqlite.org/pragma.html#pragma_locking_mode),
[WAL without shared memory](https://www.sqlite.org/wal.html#use_of_wal_without_shared_memory).

The research ledger already holds an exclusive application ownership lock and
limits its SQL pool to one connection. An explicit experimental constructor can
therefore test exclusive SQLite locking without weakening FULL WAL or changing
record semantics, but it must explicitly exclude external readers and preserve
reopen/recovery behavior. Do not apply it to unrelated or production databases.

Connection replacement is a critical boundary: startup-only PRAGMAs do not
prove replacement connections keep the contract. The pinned modernc.org/sqlite
v1.57.0 applies DSN _pragma values on each connection, but sorts them. Merely
listing locking_mode before journal_mode in a URL does not preserve that order.
Use and test an ordering supported by the pinned driver, verify locking_mode,
journal_mode and synchronous readback, then force connection replacement and
process-exit recovery. Keep ordinary Open unchanged. Only after those checks
compare unchanged prepared SQL under normal versus exclusive locking.

All seven directions remain open; no deployment or default promotion.
