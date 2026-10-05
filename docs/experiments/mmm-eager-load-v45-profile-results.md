# V45 complete mixed-load profile

2026-10-04. The original sixteen-trial V44 owner-time workload and independent
52-corruption trace audit completed successfully under unchanged frozen sources.
CPU, blocking, mutex and sampled allocation profiles, the exact test binary,
command logs and full raw trace are retained in `research/eager-load-v45-profile`.
All fourteen recorded commands exit zero. Profiling overhead makes these timings
diagnostic only; they are not new latency-adoption or production measurements.
The host was an Apple M4 with ten logical CPUs and 16 GiB RAM.

## What the stacks establish

CPU samples total 42.46 seconds over a 21.51 second profiled interval. Recall
accounts for 2.97 seconds cumulatively (6.99% of whole-process samples), including
0.83 seconds nomination, 0.67 seconds ranking and 0.60 seconds journal handoff.
These inclusive values overlap and must not be summed as independent phases.
Fixture native-index construction, envelope serialization, teardown and runtime
activity are included in the whole process; their CPU fractions are not serving
cost. Sparse sampling and C/native attribution also limit interpretation.

Blocking profiles sum waiting across goroutines, not CPU or request wall time:

| Cumulative stack | Aggregate blocked seconds |
| --- | ---: |
| Admission `Scheduler.Acquire` | 52.82 |
| Recall service | 37.81 |
| Recall journal handoff including eager refresh | 36.39 |
| Projection/warm journal delegation | 32.99 |
| Archived journal delegation | 19.68 |
| Joined journal delegation | 13.27 |
| Eager refresh | 3.40 |
| Nomination | 1.05 |

Admission includes all operation kinds; it is not a Recall-only wait total.
Journal delegation dominates the service's sampled blocked stack. It includes
commit acknowledgement and contention; it does not isolate physical fsync time.
Mutex profiles attribute 8.734 seconds of summed waiter delay to archived batch
apply (60.48% of 14.441 seconds), versus 1.575 seconds to joined apply. Unlock
attribution measures other goroutines' waiting, not the lock's hold duration.
No claim that neural/prediction math or one particular syscall causes the tail
is justified by these profiles.

## Invariant and next falsifiable experiment

The Recall read lease intentionally survives through durable journal handoff.
Cancellation cannot return a packet or orphan an accepted job; no writer may
advance the authority while the journal is pending. Releasing that lease or
acknowledging early would test a different validity/durability contract, not fix
the current one. Both joined and archived paths preserve full original wire
bytes, readback, witness/marker commit, and publish before acknowledgement.

The present commit worker collects at most four entries while admission grants
cohorts up to eight. It is plausible, but unproven, that multiple durable commits
per cohort amplify the wait. V46 prospectively forks only the test implementations
to compare caps four and eight under the complete offered workload, keeping the
one millisecond collection timer and every original gate. It checks whether
larger batches actually form and whether queue-inclusive tails improve across
both storage paths and eager arms. An eight-entry cap alone is not a rescue.

The original negative V44 latency result remains authoritative. All seven whole
goals remain open. Production, private corpora and whitepaper were untouched.
