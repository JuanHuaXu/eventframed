# Preserved Readback Invocation Failure

2026-10-04 UTC. After the diagnostic runner and independent Go audit completed,
`node research/moment-v41-readback.mjs diagnostic` exited 1 because the caller
omitted the required `EVENTFRAME_WEEKLY_USAGE` environment variable.

```text
AssertionError [ERR_ASSERTION]: explicit current weekly usage required
at research/moment-v41-readback.mjs:173:1
actual: false
expected: true
Node.js v26.8.1
```

The readback reached its final usage assertion. It did not write readback.json,
rerun/resample learners, alter the frozen sources or overwrite any study data.
Correct the invocation using the observed weekly usage (13%), retaining exactly
the same completed raw artifacts and unchanged checker. This is a command
repair, not evidence for a model bug or a durable policy/instruction edit.
