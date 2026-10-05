# Snapshot-read profile v37

Diagnostic only: repeated the twelve v36 cells with CPU/allocation profiling.
All embedded experiment source hashes exactly match v36. The run passed its
accounting assertions in 21.49s. Snapshot reads admitted 137/140/138 observations
(415/576); per-key reads admitted 133/137/130 (400/576). The remaining admission
deficit persists. Instrumented results are not an independent latency confirmation.

Independent parsing verified twelve unique cells, source-hash integrity/parity,
192 reads and 96 writes per cell, outcome/group conservation, zero errors and
50 original admissions plus 50 typed discards per accepted durable observation.

## Profile interpretation

Across both durable modes, `researchPersistGroupBatch` accounts for 1.55 sampled
CPU-seconds out of 28.72 overall. Its nested cumulative paths include AppendBatch
0.96s, SQL transaction Commit 0.63s, FcntlFlock 0.56s and Pwrite 0.44s. These
overlap and must not be added. They establish prominent storage paths, not an
isolated wall-clock disk-latency estimate or a snapshot-mode-only attribution.
Serialization alone is not established as the leading cost.

Allocation sampling estimates 683.43MB under the grouped persistence helper,
including 196.02MB under AppendBatch and 135.67MB under GetBatch, out of
6926.16MB for the whole experiment. Again these are overlapping cumulative
allocations, not live/peak RAM. More precise per-mode attribution requires
separate diagnostic arms or profile labels.

## Next discriminating experiment

Inspect the authority scope of typed discard. Admission and original validation
must remain inside the service guard; discard only retires an already admitted
pending record and supplies no label, model fit or evidence clock advancement.
Test whether that terminal write can happen after releasing the service guard,
while retaining the same FULL durable write, readback and completion accounting.
This is a hypothesis, not an implemented or established rescue.

Before a load comparison, require interleaved publication and writer tests,
commit-before/after-error recovery, bounded pending-state behavior and proof that
no feedback/fit authority escapes with the discard. A failed discard must never
be silently counted as completed. Do not generalize a discard-specific result to
labeled feedback, historical admission or learned-history validity. Keep the
current inside-guard mode as a same-run control. Other possible leads remain
SQL statement reuse and checkpoint/retention, but this profile alone does not
establish their benefit.

## Artifacts

Ran `TestResearchSnapshotReadsLoadExperiment` with `-cpuprofile`, `-memprofile`
and a temporary `-o` binary. Local profile directory:
`/tmp/eventframed-profile-v37.D90AWs` (may expire; not published).
Analysis used `go tool pprof -top -cum -focus=researchPersistGroupBatch` and
the corresponding `-alloc_space` view.

- JSONL: [mmm-snapshot-reads-profile-v37.jsonl](mmm-snapshot-reads-profile-v37.jsonl)
- JSONL SHA-256: `a325e3199160abd954f23285b630d8763f55a00656759f4c3101e45da8848382`
- CPU SHA-256: `a4a435d07de4f606b33189811225d0b58ebcb5b3c136488f03359e6ff8446f9f`
- Allocation SHA-256: `f700e5dabef22e8da74e4cee06bcf1a72f65ae09060ad23322c5eda28e7518f6`

No production configuration, paper publication, deployment or push occurred.
All seven research directions remain open.
