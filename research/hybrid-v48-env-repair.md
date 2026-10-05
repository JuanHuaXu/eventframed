# V48 diagnostic harness repair

The original `hybrid-v48-study-diagnostic` is terminal FAIL, not a live job.
Race/vet and actual seed separation passed; cost screen allocated7891552 bytes
but two uniform299 loops exceeded400ms (407.224 and405.483ms). These are valid
negative computational results, not discarded because the harness later failed.

Fixture generation was SKIPPED: the mechanically copied helper read
EVENTFRAME_SPECIALIST_V48_FIXTURE/SPLIT while the runner supplied the HYBRID
names. The downstream collector then failed opening the missing JSONL file.
No quality data were produced. Frozen copies, command logs and failure.json
retain the exact failed state. This is a confirmed integration defect, not
evidence that the predictive formulation failed. The corrected helper uses
the same HYBRID contract as the runner; the new exclusive env-repair runner
requires real nonempty output artifacts before advancing. Original learner,
priors, hazards, seeds, gates and workloads are unchanged. No production fix.
