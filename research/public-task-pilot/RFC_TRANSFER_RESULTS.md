# RFC semantic transfer: top-1 gains, packet-survival failure

The [frozen protocol](RFC_TRANSFER_PROTOCOL.md) ran 64 actual
`CaptureTurn`/`Recall` calls on fresh in-memory services: 15 new official
HTTP fact clusters, paired literal/paraphrased questions, two absent-from-
corpus controls, and 13 fixed NASA distractors. Sources are
[RFC 9110](https://www.rfc-editor.org/rfc/rfc9110.html) and
[RFC 9111](https://www.rfc-editor.org/rfc/rfc9111.html). The existing
task-plus-lexical overlay was applied only through Go's research build
overlay; the three original source hashes still matched the saved overlay.
The local semantic embedding model digest was frozen as
`0a109f422b47e3a30ba2b10eca18548e944e8a23073ee3f3e947efcf3c45e59f`.
No LLM generation or feedback fitting occurred. The runner saved raw output
without opening the oracle; the scorer loaded labels afterward.

| Wording | Baseline top1 | Priority top1 | Baseline in packet | Priority in packet |
| --- | ---: | ---: | ---: | ---: |
| Literal, 15 paired facts | 13/15 | 15/15 | 15/15 | 15/15 |
| Paraphrase, same 15 facts | 5/15 | 7/15 | 14/15 | 13/15 |
| Absent, 2 controls | 0/2 | 0/2 | 0/2 | 0/2 |

**The finite transfer gate FAILED.** Net top1 gain was +4/30 positives,
but paraphrase support survival fell by one. There were five top1 wins and
one top1 loss. HEAD and `no-store` paraphrase support left the ten-item
packet under priority; TRACE paraphrase support entered it. The 204
paraphrase lost top1. Both absent controls still packed ten unrelated
records, so neither arm demonstrated abstention or answer correctness.
Literal and paraphrase wordings share facts and are not 30 independent
outcomes. No population interval follows from these 15 clusters.

The scorer found zero invariant errors: all 28 original numeric candidate
scores and full-frontier forecast laws matched across arms for every query;
the passthrough ranker preserved its input, and every packet explanation
matched its persisted journal. Priority explanations were present in all
32 cases; no query had a tie at the maximum original numeric score.
Independent Node arithmetic reproduced every top1/survival count. The
pre-run corpus/oracle check found unique IDs and no positive question
containing its answer token. All records were available before each query's
as-of time, no label entered the service, and the scorer was separate from
selection.

Sequential, isolated Recall p50/p95 were 20.39/22.23 ms for baseline and
21.61/23.81 ms for priority. These block-ordered samples cannot estimate
causal overhead, concurrent p99, persistent-store cost or OpenClaw latency.
The model is a local public-text embedder, not an agent completion model.

The [raw rankings](rfc-transfer-v1/raw-results.json) have SHA-256
`a746cc99e67c593824128556698d191cc736a444926d63d22783df67762dc191`;
the [summary](rfc-transfer-v1/summary.json) records every case. Focused
research command compilation and vet passed with the saved overlay.

The trace suggests a specific *diagnostic*, not an adopted fix: the task
plan marked all inspected failing candidates `unknown`, and the no-store
paraphrase had zero lexical scores. A fixed-budget incumbent-preserving
packet fusion or evidence-gated fallback could protect support, but this
RFC set is now consumed for that idea. Freeze the rule first and test it on
another untouched public domain with both top1 and packet-survival gates;
do not tune this result into a confirmation claim. Goal 5 remains open.
