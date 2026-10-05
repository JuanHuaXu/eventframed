# Semantic embedding ablation: preserve semantic ordering

42 actual CaptureTurn/Recall query runs on consumed public tasks. The existing
local nomic-embed-text model used the daemon's OpenAI-compatible embedding
adapter,768 dimensions and task prefixes. Store remains in-memory; this is not
the LibraVDB retrieval/ranking backend or OpenClaw agent execution.

| Arm | Literal top1 | Paraphrase top1 | Positive support survival |
| --- | --- | --- | --- |
| Semantic baseline | 6/6 | 5/6 | 12/12 |
| Semantic nomination + lexical direct order | 6/6 | 3/6 | 12/12 |
| Semantic nomination + sparse direct order | 6/6 | 2/6 | 12/12 |

The semantic-baseline lead PASSED: hash baseline paraphrase top1 was2/6.
Both experimental overrides FAILED non-harm against the semantic baseline.
All callbacks saw19 candidates. Learned lexical features do not replace
semantic similarity: direct ordering discards information already available in
the backend score. No production policy was enabled or modified by the runs.

This also changes interpretation of prior rescues. Their benefit was established
against a weak hash-embedding baseline, not a real semantic baseline. Preserve
the old evidence, but do not generalize it to a semantic production deployment.
Neither lexical nor learned direct ranking is recommended for adoption here.

Observed Recall durations across42 requests were approximately14.5-25.9ms,
including local query embedding and memory-store processing, excluding corpus
ingestion. Arm blocks were sequential, not randomized; timings do not estimate
causal overhead, concurrent p99 or persistent-server latency. No generation
model was called; only sequential local public-text embedding inference.

Hash and metric audits passed. Embedding outputs are not claimed deterministically
replayable from source hashes alone; installed model digest was recorded in the
protocol and raw rankings retained. This is a controlled ablation on consumed
tasks, not independent generalization evidence.

Next: preserve semantic retrieval as incumbent, test learned residuals only
with support evidence that semantic ranking lacks, and use genuinely new tasks
for any hybrid tuning. Required controls include semantic baseline, unmodified
rank order, and explicit unsupported/ambiguous cases. Actual agent evaluation
and persistent concurrent load remain missing for roadmap5/6.
