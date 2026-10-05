# Prepared admission SQL attribution v68

Frozen before profiling. Run the existing v65 admission-phase experiment with
CPU and block profiling, leaving its12 cells and original-record checks intact:
50/200 batches,cold/trained64labels,3 trials,32 fresh/retry/terminal cycles.
The prepared resolved-source path is selected, not v66 or v67 candidates.

Use exclusive JSONL and unique local profile paths. Record source and profile
hashes. No competing task-started tests, production or dependency changes.
No code changes are needed for this diagnostic; prior race/vet verification
applies to the unchanged measured path.

Inspect SQL preparation/compiler, VM execution, record/index operations and
commit/filesystem call stacks. Profile totals include fresh, retry, cleanup,
setup and replay: distinguish cumulative CPU or Go blocking from wall time,
and do not claim a fresh-only fsync percentage from aggregate samples. Named
v65 phases still provide fresh/retry outer attribution. Do not add nested stacks.
This is not a new speed screen or non-harm confirmation. If profiling cannot
separate compilation from commit cost reliably, state that limitation instead
of using syscall names as proof of a specific bottleneck.
