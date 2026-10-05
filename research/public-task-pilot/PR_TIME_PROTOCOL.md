# Puerto Rico UTC/local-time retrieval transfer

Frozen 2026-10-01 before any model output on this set. Source is the official
[USGS earthquake catalog API](https://earthquake.usgs.gov/fdsnws/event/1/)
for M>=4 events around Puerto Rico during January 7-8, 2020. Freeze the
response bytes, SHA-256, and deterministic selected IDs. The corpus takes the
first 24 events sorted by time; first 12 form design, next 12 are held-back
confirmation. Two remaining events form absent-from-corpus design controls.
Each event's source record gives UTC time and preferred magnitude. Each query
asks for that magnitude in two paired forms: exact UTC ISO time, and the
equivalent Puerto Rico local time (AST, UTC-4) written in natural language.
The latter does not repeat the corpus timestamp string. Unique seconds are
required for all 24 selected events. Answers and target IDs stay in an
oracle file outside the service process.

First run focus and the unchanged task-plus-lexical priority overlay on design
only. Use the same Nomic digest, fresh service per query, recall50, pack10,
diversity enabled, and passthrough numeric ranker as the USGS Ridgecrest
screen. Save raw packets, ranker traces, full-frontier laws/scores, journal
agreement and timings before opening the oracle. No learner fitting, feedback,
LLM generation or production OpenClaw. Do not run confirmation until a
specific normalization rule and gate are frozen.

The design set justifies a normalization experiment if at least four of 12
local-time questions miss top1 under the priority overlay while all 12
targets remain in the ranker frontier. Also report UTC controls, survival@10,
per-case changes and absent controls. If there is no local-time deficit, do
not invent a rescue. If targets are absent from the frontier, the problem is
nomination, not merely rank/packing. Confirmation remains untouched at this
stage.

This is a structured public QA retrieval probe, not a real agent-outcome
trial. A design-only result is diagnostic. A later passing same-sequence
confirmation cannot by itself satisfy goal 5's broader agent-task criterion.
