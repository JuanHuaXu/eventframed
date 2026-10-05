# As-of validation profile v24

Diagnostic only. Repeated the unchanged nine-arm v23 experiment with Go CPU and
allocation profiling enabled. Source-bearing JSONL is
`mmm-asof-profile-v24.jsonl`; SHA-256
`7c6d4ac94d4de90794fe8113ee3d0696ab94b679704f70179dc03fd5ee30100e`.

Run command: `EVENTFRAME_ASOF_LOAD_ARTIFACT=... go test ./internal/service
-run '^TestResearchAsOfLoadExperiment$' -count=1 -v -cpuprofile ... -memprofile ...
-o ...`. Profiles and test binary were written to the isolated local directory
`/tmp/eventframed-profile-v24.gy5K6e`, not the source tree or production.
CPU SHA-256: `1a4f73bbaa96c97569af65b06be5180ec87d078e9c1c5740ab8c3bd4ca213817`.
Allocation SHA-256: `cdfc7d4abc80db0a83371db9521797b0abf8670d86dcac5d124769dd16ce076b`.
Raw temporary profiles may expire; the JSONL preserves the measured experiment's
selected sources/configuration, not the profiler samples.

`go tool pprof -top -cum -focus=ValidateResearchAdmission` reported 0.86 seconds
of sampled CPU under validation, including 0.74 seconds under journal reading
and JSON decoding: about 86% of validation CPU. This is only 4.29% of the whole
experiment's 20.04 CPU-seconds; it does NOT explain all elapsed time or waiting.

The corresponding `-alloc_space` profile estimated 1190.94MB allocated under
validation, including 1075.50MB under journal reads: about 90%. These are sampled
cumulative allocated bytes, not resident or peak RAM. Total experiment allocation
was 5366.39MB. CPU/heap sampling and instrumentation introduce uncertainty.

As-of accepted 75/74/74 observations, dropping the other 353 of 576; exact queued
controls again accepted zero. There were no execution errors. This replicates
the bottleneck shape, not an independent uninstrumented latency confirmation.

Source inspection matches the profile: each candidate's single-record validator
reads and decodes the entire same journal, recomputes the same query embedding
and digest, scans journal decisions and retrieves its event. For frontier N and
journal length J this repeats shared O(J) work N times. A per-call batch can
reduce it to O(J+N), apart from backend/event/feature costs, while retaining all
record checks and final dependency validation. No cache survives the call.

Competing contributors remain writer waiting, event decoding, feature extraction,
serving GC and queue scheduling. The v25 same-run batch ablation tests whether
removing the measured redundant work actually improves admitted throughput and
age. If not, this profile alone cannot justify calling the queue issue rescued.
