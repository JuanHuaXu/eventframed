# Public-Fact Text Replay Corpus

This directory contains generated conversations grounded in verifiable public
facts. It exercises the complete EventFrame chat path:

```text
raw user/assistant text -> post-contract 5W1H extraction -> EventFrame corpus
-> retrieval -> EventFrame reranking -> packing
```

The corpus contains 32 sessions, 384 turns, and 128 labeled recall queries.
Sessions are split 24/8 into design and confirmation blocks. Topics cover ten
commonly confused factual distinctions documented by NASA, NOAA, USGS, NIST,
and the Smithsonian.

The dialogue wording is generated, but the substantive facts are not fictional.
Every factual record carries its public source URL. No private transcript text,
personal identifier, invented entity, invented measurement, or unpublished
research content is present.

The case, session, turn, run, agent, and tenant labels are deterministic local
keys required by the EventFrame contract and evaluation oracle. They identify
no person, account, machine, production session, or external system and have no
meaning outside this dataset.

Each JSONL record contains:

- `capture`: the request accepted by `eventframed`;
- `sources`: evaluation-only public provenance;
- `oracle`: evaluation-only relevant and obsolete prior event IDs.

Only the nested `capture` value should be sent to the daemon. Source and oracle
metadata must not be indexed as conversation memory.

Regenerate the checked-in corpus and manifest with:

```sh
go run ./cmd/eventframe-synthetic-text
```

The generator is deterministic. `manifest.json` records counts, source URLs,
and the SHA-256 digest of `corpus.jsonl`.
