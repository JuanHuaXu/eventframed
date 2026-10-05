# Adapter comparison v5

Diagnostic protocol frozen before execution. Six trials, rotating off,
ingestion-only publication adapter and general capability-aware adapter. Each
uses the existing frozen workload:192 recalls, four readers,96 future-only
writes, queue16 for enabled arms,50 fixture labels per admitted frontier.
No production settings change. This does not test mixed mutations or accuracy.

Retain v4 gates: no errors, overlapping writes, at least154 admissions, every
admitted label completed, no worker failures, completion-age p95<=250ms and
serving p99<=1.10 times paired off. Store18 arms and embedded source hashes in
exclusive-create JSONL. Report paired admission and latency differences as
descriptive evidence, not a population equivalence test. A pass cannot erase
v4 failures. If both adapters fluctuate around the admission gate, investigate
consumer service rate instead of attributing the failure to general wrapping.
