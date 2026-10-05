# Envelope batch-read diagnostic

Read-only candidate test, no daemon changes. Against the existing indexed
GetServiceAdmissions control, compare bounded per-call envelope reuse with a
zero-cache negative control. All paths validate requested source, original key
and payload binding. Preserve order, misses and duplicates; no partial results
on corruption. Envelope cache cap 8MiB, independent returned-byte cap 8MiB;
uncached bodies use bounded slices, not dropped results.

Three rotated trials; 32 stored envelopes of 200 events, each payload contains
Binding and 1024 padding characters. For sizes 50/200 run 32 queries in each of
two layouts: grouped consecutive events, or round-robin across stored batches.
Measure all calls and validate all returned bytes/sequences outside timing.
Capture sources, pins, protocol and raw durations. No threshold tuning, production
access, private data or concurrent task-started benchmarks. Report cache counters
and cache-disabled time to test repeated-blob hypothesis, not just speed ratios.

This is a consumed synthetic read diagnostic. It cannot establish end-to-end
admission, feedback, crash recovery, quality or serving requirements. Preparation,
database creation and fixture insertion are outside read timing. No candidate
is promoted based on this microbenchmark alone.
