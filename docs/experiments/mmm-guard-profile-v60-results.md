# Guard path profile v60

The unchanged v59 fixture completes with all integrity checks. Captured source
hash maps are exactly equal to v59. Profiling identifies publication-gate and
native journal durability waits as leads; it does not validate a latency rescue
or justify removing validation/locking.

[Protocol](mmm-guard-profile-v60-protocol.md),
[JSONL](mmm-guard-profile-v60.jsonl), SHA-256:
`3213022c3c1f3e1d9c20ae43f9658265fb7acea0c92b084fc083460f0f8bfe66`.

## Execution

```sh
EVENTFRAME_GUARD_ONLY_ARTIFACT=<LOCAL_ROOT>/docs/experiments/mmm-guard-profile-v60.jsonl go test ./internal/service -run '^TestResearchGuardOnlyExperiment$' -count=1 -cpuprofile=/tmp/eventframed-guard-profile-v60.utN4MV/cpu.pprof -blockprofile=/tmp/eventframed-guard-profile-v60.utN4MV/block.pprof -blockprofilerate=1 -mutexprofile=/tmp/eventframed-guard-profile-v60.utN4MV/mutex.pprof -mutexprofilefraction=1 -o /tmp/eventframed-guard-profile-v60.utN4MV/service.test
```

Go PASS30.194s. All18 cells retained:86400 durable originals/terminals and115200
validated candidates including28800 validation-only candidates. Zero drops,
expiry or unexpected errors. Source/hash equality was independently checked
against both v59 and local files. No runtime code changed or competing test run
was started. No new race-suite claim: v59's full race/vet checks cover the same
code; this run adds profiling overhead and is not a timing confirmation.

Profiles and matching binary remain in the local directory shown above:

| File | SHA-256 |
| --- | --- |
| cpu.pprof | 85cb0e10536ab3363ea72fe89653b6de307a2f37238949b0a85c65a9343a8893 |
| block.pprof | 730654a843c866b2d44c8a0cb0e929284273ce47d65b37bb313695955a31eb2d |
| mutex.pprof | a4d0854a1823a1561707f8446923c370a418a38841d74b2c35dd585aff361d98 |
| service.test | 007246c9e6686977cae9e14bef174710f2d7548a99c8e7a7f5578142935bf2a0 |

## Findings

- Blocking under WithResearchAsOfSnapshotWait totals4.930s, of which4.663s is
  semaphore acquisition and0.267s is inside the callback. The latter is almost
  entirely native collection read-lock waiting while retrieving journals.
  Native event retrieval contributes about0.00081s of measured blocking there.
- Across all arms, native PutBayesianJournal accumulates31.53s blocked:9.60s
  at its store read lock,1.31s in existing-record lookup,20.27s inside insertion,
  and0.349s at the journal stripe lock. These are approximate profile values.
- Among wrapper-routed journal stacks,13.74s is specifically under the native
  WAL-flush wait. That is a subset of insertion-related waiting, not another
  duration to add. Journals hold the native store read lock through compatibility
  checking and durable insertion; event Put needs its write lock.
- Native event Put accumulates5.07s waiting for writeMu and1.93s blocked within
  its transaction, with0.273s at eventMu. Mutex profiling attributes37.33s of
  aggregate contention to releasing stacks under Put. Releaser attribution is
  not a per-call hold duration or proof that Put alone caused every delay.
- CPU sampling totals42.05s over29.74s wall duration. Only0.08s is sampled under
  union validation; wrapped event Put has2.13s cumulative CPU, including1.53s
  under index building. This cold, no-fit fixture does not show expensive
  validation computation as the main bottleneck. It says nothing about training
  or particle/MCMC cost, which are absent here.

Total blocking-profile delay is250.82s and mutex-profile delay51.59s, summed
across concurrent goroutines. They are not request latency or wall-clock time.
Profiles aggregate all six arms without arm labels; do not attribute these
totals specifically to guard-only or resolved source. Nested cumulative numbers
must not be added. Source-line listings identify wait sites but do not by
themselves prove a safe alternative ownership protocol.

## Reproduction

With the matching binary and profiles above:

```sh
go tool pprof -list='WithResearchAsOfSnapshotWait' /tmp/eventframed-guard-profile-v60.utN4MV/service.test /tmp/eventframed-guard-profile-v60.utN4MV/block.pprof
go tool pprof -list='libravdbstore.*Put$|libravdbstore.*PutBayesianJournal' /tmp/eventframed-guard-profile-v60.utN4MV/service.test /tmp/eventframed-guard-profile-v60.utN4MV/block.pprof
go tool pprof -top -cum -nodefraction=0 -focus='validateResearchAdmissionGroups' /tmp/eventframed-guard-profile-v60.utN4MV/service.test /tmp/eventframed-guard-profile-v60.utN4MV/cpu.pprof
```

## Next lead

Inspect native WAL batching configuration and durable receipt boundaries before
changing the research guard. Test an explicit bounded flush/batching policy if
the existing backend supports it, retaining acknowledgment only after durability.
Do not simply release the journal read lock early: it currently protects the
snapshot-compatibility check through insertion, and any replacement needs an
equivalent validity/commit contract plus conflict and crash tests.

v59's guarded package shifts read/write balance, but the profiles point to a
broader interaction with ordinary serving journals and event transactions.
Keep v55/v57/v58/v59 failed non-harm screens unchanged. All seven directions,
real outcome validation and full feedback/history authority remain open.
Nothing pushed, deployed or changed in the whitepaper's accuracy claims.
