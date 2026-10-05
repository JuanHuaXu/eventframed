# V44 trace reader repair

Confirmed audit-tool failure, 2026-10-04: all sixteen loaded trials completed,
but `readFileSync(..., 'utf8')` attempted to create a 584 MiB string, exceeding
Node's `0x1fffffe8` single-string limit. Original command, exact source freeze,
raw trace and failure remain in `research/eager-load-v44`. No trial is rerun,
removed or resized to avoid this failure.

Candidate causes considered: malformed JSON, a single oversized trial, or the
whole-file string conversion. The exception originates at whole-file conversion
before parsing. Incremental reading of the exact existing trace is the narrow
repair; any subsequent scientific-checker failure must still be investigated.

The new envelope reader checks header/footer/count, rejects extra/truncated/
blank/malformed records, visits every trial and hashes every original byte.
One positive byte-hash control and eleven negative controls cover parser and
scientific-checker rejection. All per-trial scientific checks and thresholds
remain unchanged. Memory is bounded by the largest row rather than the entire
trace; this is not a serving-performance optimization or an RSS claim.

A supplementary freeze must bind the repaired reader/auditor and the original
raw hash. The original measured code is not described as changed or rerun.
Latency logs are preliminary until the independent checker accepts the trace.
