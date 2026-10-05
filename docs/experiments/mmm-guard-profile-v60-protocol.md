# Guard path profile v60

Frozen before execution. Rerun the unchanged v59 six-arm offered-load fixture
with CPU, blocking and mutex profiles. Capture every blocking/mutex event for
diagnosis; profiling overhead means timings are not an independent performance
confirmation. Retain all18 cells, existing integrity checks and exclusive JSONL.
Compare captured source hashes with v59 before interpreting differences.

Use go test CPU profile, blockprofilerate1 and mutexprofilefraction1; retain the
matching test binary and profile hashes. Inspect publication semaphore waiting,
native store read/write mutex sites and guarded validation call paths. Profiles
aggregate all arms and goroutines: cumulative wait seconds can exceed wall time,
nested cumulative values cannot be added, and absence of arm labels prohibits
per-arm attribution. Mutex samples identify releasing stacks, not necessarily
the waiter or complete owner duration. No profiling percentage is a causal proof.

No code/default changes, private data, production, push or deployment. A profile
lead must survive a scoped follow-up test before changing validation/locking.
Earlier failed non-harm screens and all seven open research directions remain.
