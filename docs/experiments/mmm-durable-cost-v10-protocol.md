# Durable component cost v10

Frozen diagnostic:12 arms,256/1024 labels, three alternating in-memory/durable
pairs. Wait for completion every64 labels; same feature/outcome sequence and
seed, no real-user data. Measure per-admission/feedback duration, total time
through completion, closed database bytes and full-log OpenDurable replay time.
Verify completed counts and all512 pre/post-restart predictions per durable arm.

No speed success threshold is invented here: this identifies persistence and
replay costs before choosing a realistic serving integration budget. It is not
concurrent retrieval, independent task accuracy or an end-to-end latency claim.
Mode scheduling can change issued forecasts; compare restart to its own issued
model, not a different mode's model. Preserve source/hash header and raw arrays
in exclusive-create JSONL. Report all arms; do not overwrite slow results.
