# Semantic embedding ablation on consumed Nobel tasks

Use the existing local Ollama nomic-embed-text:latest model, advertised digest
0a109f422b47e3a30ba2b10eca18548e944e8a23073ee3f3e947efcf3c45e59f,768 dimensions,
F16. Existing daemon OpenAI-compatible EmbedDocument/EmbedQuery contract with
search_document: / search_query: prefixes, endpoint127.0.0.1:11434/v1/embeddings.
No model installation, provider change, private data or production OpenClaw access.

Same consumed19-record Nobel/NASA corpus and14 questions, three arms: baseline,
lexical direct, sparse direct. Same frozen design-only weights. Fresh service
per query/arm, real CaptureTurn and Recall, pack10/recall50. All embedding calls
sequential. Local embedding inference is allowed; no generation/LLM agent calls.

Purpose: isolate semantic embeddings from hash embeddings. No new-generalization
claim, tuning or learning on this set. Record actual callback candidate count,
support rank, ties and request time. Request time includes query embedding but
excludes corpus preparation; not a concurrent/persistent deployment benchmark.

Read oracle for evaluation after all requests. Save all results and source/model
metadata. Success lead: semantic baseline paraphrase top1 improves over2/6 while
retaining6/6 literal top1. For each experimental ordering separately require no
loss versus the semantic baseline. Failed overrides must not be enabled simply
because the embedding baseline improves.
