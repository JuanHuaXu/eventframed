# Untouched NASA mission-date retrieval transfer v1

Frozen 2026-10-01 before any model run on `nasa-transfer-v1`. This is a
public, outcome-labeled retrieval screen, not an agent-answer experiment.
The corpus contains four dated events each from NASA's New Horizons, Dawn,
and OSIRIS-REx mission timelines. Each event has one literal and one
paraphrased question. The twelve fact identities, twenty-four queries, and
oracle dates/support IDs are frozen before serving.

Use the unchanged `sparse-go-golden.json` weights, fitted on earlier consumed
data. No NASA-transfer labels may train, tune, select a threshold, or change
the model. Compare three frozen arms: service baseline, direct lexical score,
and direct frozen sparse score. Use the local Nomic embedder only if its digest
is `0a109f422b47e3a30ba2b10eca18548e944e8a23073ee3f3e947efcf3c45e59f`.
Each arm gets the same 12 captured facts in a fresh in-memory service per
query, recall 50, pack 10, and the same timestamp. The research ranker sees
the nominated frontier before packing. Persist raw packets and code/data
hashes to a new file before opening the oracle in a separate scorer.

Primary screen: on the twelve paraphrases, sparse top-1 support must exceed
baseline by at least two questions, with no lost literal top-1 success and no
lost support at pack 10. Exact date extraction from the top support record is
a deterministic packet proxy, not an LLM answer. Also report per-mission
results, frontier coverage, runtime, score/law agreement, and lexical control.
Mission clusters, not candidate rows or paired wordings, are the independent
units. Three clusters cannot establish population improvement. A failure is
retained as a transfer failure; do not repair it using this set as confirmation.

Sources: [New Horizons](https://science.nasa.gov/mission/new-horizons/),
[Dawn quick facts](https://science.nasa.gov/mission/dawn/toolkit/quick-facts/),
and [OSIRIS-REx in depth](https://science.nasa.gov/mission/osiris-rex/in-depth/).
