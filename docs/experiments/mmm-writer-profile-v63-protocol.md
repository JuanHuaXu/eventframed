# Four-writer arm attribution v63

Frozen before profiling. Preserve the v62 four-slot native-off and active
resolved-source fixtures, three trials each at5ms reads/10ms writes. Run each
arm in a separate process with CPU, block(rate1), mutex(fraction1) profiles and
an exclusive JSONL source snapshot. No two-slot or mixed-arm stacks in these
profiles. Default twelve-cell v62 entry point remains unchanged.

Keep all duration/identity/phase/count checks. This is diagnostic, not a fresh
non-harm confirmation: profiling perturbs timing, and cumulative blocked time
can exceed wall time across goroutines. CPU and mutex totals include setup,
teardown and source capture; mutex release stacks are not lock hold durations.
Do not add nested inclusive totals. Compare publication-gate waiting, native
journal barrier/WAL waiting, event ingestion and retrieval lock paths. Retain
v62's failed screen regardless of these timings. No runtime/default changes.

Record commands, artifact/profile hashes, runtime, completion and limitations.
Run focused harness race/accounting and vet before profiling. No concurrent
task-started tests/benchmarks, production access or module-cache changes.
