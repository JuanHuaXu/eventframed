# Capability-aware adapter load v4

Frozen before execution. Fork the v3 workload with only the enabled store adapter
replaced by researchpublicationstore.Wrap. Keep v3 sources and results unchanged.
This measures actual background refits alongside persistent ingestion and recall;
it does not exercise every general mutation concurrently or establish accuracy.

Use 64 and 192 recalls, four readers, requests/2 future-only writes, three rotated
trials of off/queue16/queue64. Fifty explicit fixture labels per admitted frontier.
Repeated fixture labels are load, not independent evidence of learning quality.

Retain every v3 screen: no errors, write/read overlap, enabled recall p99 at most
1.10 times paired off, at least 80% frontier admission, every admitted label
completed, no failed updates, and completion-age p95 at most 250ms. Quantiles use
nearest rank. Store raw timings/counts and embedded source/hash header in an
exclusive-create JSONL file; do not overwrite failures or retune gates afterward.

Failure of any cell remains a failure. Passing this finite screen does not prove
population tail bounds, crash recovery, duplicate recovery or mixed-write safety.
