# USGS held-back confirmation gate

Frozen after the [design-only run](USGS_HEADROOM_PROTOCOL.md) and before any
confirmation service call. The 24 confirmation questions, answers, corpus,
source snapshot, model digest and task-plus-lexical overlay were generated and
frozen before the design run. This gate introduces no fitted coefficient,
query rewrite, new record, or threshold in the served algorithm. Design data
showed focus 4/24 and priority 24/24 top1, with all targets in the 24-record
frontier. That is motivation for confirmation, not confirmation itself.

Run focus and the unchanged priority overlay on only the 24 held-back July 6
queries, with fresh in-memory service per query, recall50, pack10, diversity
enabled, unchanged numeric ranker and the same Nomic digest. Capture raw
ordered packets, full-frontier laws/scores, ranker traces, journal agreement
and Recall times before opening the oracle. The confirmation set has no absent
controls; two absent controls were design-only. No LLM generation, feedback,
new training, or production OpenClaw.

Finite confirmation passes only if all hold:

1. Priority has at least eight additional correct top1 records across the 24
   positive questions and at least one gain in each wording.
2. No positive question loses top1 or target survival relative to focus;
   priority survival is also no worse in each wording aggregate.
3. Every ranker frontier has the same 24 unique records; the numeric ranker
   leaves candidates unchanged; full-frontier forecast laws and original
   numeric scores match across arms; all packet explanations match journals.
4. The isolated, sequential nearest-rank p95 Recall duration is below 100 ms
   for each arm. This is a local feasibility screen, not a loaded p99 claim.

This confirms at most in-sequence transfer for timestamp-heavy public
retrieval. Source events share one earthquake sequence and the query templates
are fixed. It cannot establish actual agent answer quality, cross-domain
generalization, or completion of goal 5. Preserve any failure.
