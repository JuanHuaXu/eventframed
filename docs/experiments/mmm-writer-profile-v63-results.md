# Four-slot arm attribution v63

## Finding

The remaining active writer delay is not solely native journal contention.
Across the three active cells, ingestion waits1.090s at the publication gate;
native Put itself blocks0.703s versus1.235s in the native-off profile. The
guarded callback occupies1.096s of measured wall duration, including0.838s of
durable admission. This makes shortening guarded admission a concrete next
lead, provided snapshot and original-record validity can be preserved.

These profiles do not rescue v62's failed non-harm screen. No runtime code,
default constructor, durability mode or dependency changed.

## Setup and verification

[Protocol](mmm-writer-profile-v63-protocol.md).
Separate off4 and active4 processes, three unchanged v62 cells each. The same
365 source hashes and the same executable hash appear in both runs; all source
hashes were recomputed against embedded and local content. Focused accounting
race test PASS5.132s; service vet PASS. Profiled experiment package times:
off4 PASS5.335s, active4 PASS4.996s. Active cells accept all576 observations,
verify28800 original records and28800 terminal discards, with no drop, expiry
or errors. Across both arms:1152 recalls and576 event writes.

Artifacts:
- [off4 JSONL](mmm-writer-profile-v63-off4.jsonl):
  `3aa4af9c9c14c7c111e3eb2aa2920cda56e3d01827204aacd27440a543d6b00e`
- [active4 JSONL](mmm-writer-profile-v63-active4.jsonl):
  `54aab9b77dba271deb322e7f6b14b61344c6e6789dd7bd0c0e4ae3801019626f`

Profiles and executables remain under
`/tmp/eventframed-writer-profile-v63.pCtuH5/`; these local diagnostic files are
not durable repository assets. SHA256:

| File | SHA256 |
| --- | --- |
| off4.cpu.pprof | b965845c3461084a67cb24678902bc183d8810aec8d4b468da088072edcb0b6e |
| off4.block.pprof | d73a0d45d07044ed92b2bfc6d85fedd7083b60735a2f476520d6e1d85b870db3 |
| off4.mutex.pprof | 02f6009f2e40b4dbe320bdbfa3e87dba4c3bfd5ee06062e37fcddef6cf136535 |
| active4.cpu.pprof | 516b23ccfd2c5bf76240efbb2811912f11de04d1a60ffd7c99e82463b2e04da1 |
| active4.block.pprof | 56140185fcebecc34c81bb07156f848acb6e1e4922c48b4b4a3d4cda7aac0acd |
| active4.mutex.pprof | 3be984cfa9d8eb5a5bcde12415b977ff4800ed1028b2803b79ae9ee7aeb5b185 |
| off4.test and active4.test | 15db016e7bce74e54da05b2b4f781fe21d28a2f51e831516ba17736deb79fde6 |

Each command used `go test ./internal/service -run
'^TestResearchWriterProfileExperiment$' -count=1 -v` with
`EVENTFRAME_WRITER_PROFILE_ARM=off4` or `active4`, corresponding exclusive
`EVENTFRAME_WRITER_PROFILE_ARTIFACT`, `-cpuprofile`, `-blockprofile`,
`-blockprofilerate=1`, `-mutexprofile`, `-mutexprofilefraction=1`, and `-o` paths
above. No concurrent task-started tests/profilers. Go1.27.1, Apple M4,
darwin/arm64, GOMAXPROCS10.

## Attribution

Cumulative blocked seconds, not per-request latency; inclusive rows overlap.

| Boundary | Native off4 | Active4 |
| --- | --- | --- |
| Native event Put total | 1.235 | 0.703 |
| Put writeMu.Lock | 0.815 | 0.342 |
| Put eventMu.Lock | 0.064 | 0.061 |
| Put WithTx | 0.356 | 0.301 |
| Native journal Put total | 5.586 | 2.921 |
| Journal writeMu.RLock | 2.52 | 0.224 |
| Journal collection Insert | 3.03 | 2.55 |
| Ingestion publication-gate Acquire | n/a | 1.090 |
| Research guard Acquire | n/a | 0.930 |
| Guard callback blocking | n/a | 0.019 |

The callback's small Go-block-profile time does not mean it is cheap. Monotonic
phase instrumentation sums1.096s inside callbacks, of which0.838s is durable
admission; post-guard verified cleanup adds0.974s outside the gate. Filesystem
syscalls and running work are not all Go synchronization blocking. Do not add
these phase sums to inclusive profile totals as if disjoint.

Mutex aggregate delay is10.260s off4 and4.82s active4; release stacks under
native Put account for9.015s and3.58s respectively. Those are attributed waiter
delays, not lock holding durations. Active CPU samples total7.52s over4.62s
wall time and include setup/capture/teardown; they do not establish training
cost or a dominant application CPU bottleneck.

Profiled scheduled write p99 is34.857/45.660/136.018ms off4 and
204.153/186.479/216.162ms active4. Active observation age p95 is
114.995/89.377/84.543ms. Profiling and non-interleaved arm order preclude treating
these as a fresh confirmation comparison; v62 remains the relevant screen.

## Next experiment boundary

Investigate whether immutable admission staging can occur outside the
publication gate, leaving only the validity-sensitive transition inside it.
First audit source identity allocation, actual forecast timing, snapshot
compatibility, uncertain commits, replay and learner publication. Merely moving
a durable write or releasing a snapshot barrier early is not an acceptable
rescue. If that split cannot preserve the current contract, retain the barrier
and investigate bounded scheduling instead. No lead is exhausted by this
profile, and no research direction is complete.
