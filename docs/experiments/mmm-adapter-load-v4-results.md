# Capability-aware adapter load v4: FAILED

The frozen 18-arm study used the actual background learner and persistent store
through researchpublicationstore.Wrap. Raw evidence: mmm-adapter-load-v4.jsonl.
The test exited 1 on three failed cells. Source hashes, embedded/current sources,
18 unique arms, request/write counts and label conservation were independently
checked afterward. No thresholds or artifacts were changed.

All arms had zero errors and overlapping reads/writes. All admitted labels
completed (50 per frontier), with no worker failures. All enabled serving-p99
ratios passed the 1.10 paired-control limit. Short 64-request trials passed both
capacities: queue16 admitted 60/57/57; queue64 admitted all64.

Long workload (192 requests), trial order 0/1/2:

| Queue | Admitted | Completion-age p95 ms | Serving p99 / control | Result |
| --- | --- | --- | --- | --- |
| 16 | 149, 157, 148 | 110.720, 106.482, 110.224 | .7741, .8808, .7950 | FAIL admission in 0 and 2 |
| 64 | 192, 192, 192 | 236.586, 264.744, 232.534 | .8748, .8222, .8219 | FAIL age in 1 |

Queue16 requires at least154 admissions to reach80%. Queue64 exceeds the250ms
age ceiling once. Thus neither capacity robustly passes this finite workload.
An earlier ingestion-only pass cannot be carried over to the general adapter.
These separate runs do not isolate adapter overhead from machine/scheduling
variation; no causal performance regression is established by this comparison.

Next useful experiment: compare both adapters within the same randomized/rotated
load study and instrument admission versus refit/wait time before another rescue.
Do not choose a new queue size solely to pass these consumed cells. Recovery,
mixed mutations, persisted feedback and real-task quality remain open. Repeated
fixture labels test workload only, not independent predictive accuracy.
