# USGS near-duplicate retrieval headroom protocol

Frozen 2026-10-01 before service/model output. Source is the official
[USGS earthquake catalog API](https://earthquake.usgs.gov/fdsnws/event/1/)
for M>=4 events near Ridgecrest, California, from 2019-07-04 through
2019-07-07 UTC. The exact query URL, response SHA-256, and selected event IDs
are stored with the derived fixture. The historical catalog is a public
factual source, not synthetic truth. A new API revision does not replace the
frozen response.

Selection: sort events by time ascending. Use the first 12 on July 4 for
design and the first 12 on July 6 for untouched confirmation. Both groups
are present in the 24-record corpus; only design questions run at this stage.
Two absent controls come from July 5 and July 7 events excluded from the
corpus. Each included event has two paired magnitude questions: literal UTC
ISO timestamp, and a natural-language UTC date/time. Questions omit the
magnitude. Target event ID, magnitude answer, and split are in an oracle file
not opened by the service runner. Paired wordings share an event cluster.

Run the existing semantic Nomic model and frozen task-plus-lexical research
overlay in isolated in-memory services, focus versus priority, recall50,
pack10, diversity enabled, unchanged numeric ranker. Capture raw ordered
packet IDs, ranker inputs/outputs, full-frontier laws and scores, journals,
and isolated Recall timings before scoring. No learning, feedback, LLM
generation or production OpenClaw. No method is fitted or tuned on the
confirmation questions before a new rule and gate are frozen.

Design headroom qualifies for a subsequent rescue experiment only if focus
top1 is at most 18/24 positive questions and at least four target IDs are
present in the ranker frontier but not top1. If the baseline saturates or
targets are missing from the frontier, do not tune a top10 promotion rule on
this set. Report survival@10 and top1 separately for each wording, the two
absent controls, and all changed choices. The confirmation split remains
unrun until a specific rescue rule and non-harm gate are frozen.

These templated public-fact questions are retrieval probes, not actual agent
outcomes. Even a successful confirmation would not alone satisfy goal 5's
prospective real-task criterion. Sequential isolated timings are not p99 or
full-agent latency claims.
