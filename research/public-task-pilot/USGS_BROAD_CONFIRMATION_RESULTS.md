# Broad-location USGS transfer: finite retrieval pass

The [predeclared current-contract protocol](USGS_BROAD_CONFIRMATION_PROTOCOL.md)
**PASSED** on an untouched official
[USGS catalog API](https://earthquake.usgs.gov/fdsnws/event/1/) snapshot.
The source returned 219 eligible events for September 15–23, 2026; the first
24 by time and ID were selected without looking at retrieval outcomes. They
have 24 different source place strings. Each event has literal UTC, natural
UTC, and explicitly offset `UTC+11` questions. The answer is the source
snapshot's reported magnitude, with event ID as the retrieval target.

| Arm | Literal top-1 | Natural UTC top-1 | UTC+11 top-1 | Total top-1 | Total support@10 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Passthrough focus | 3/24 | 23/24 | 22/24 | 48/72 | 63/72 |
| Task-lexical ranker | 24/24 | 24/24 | 24/24 | 72/72 | 72/72 |

There are 24 paired top-1 gains affecting 21 distinct events, zero top-1
losses and zero pack-survival losses. The literal wording accounts for 21 of
the 24 gains; the other wordings account for one and two. Every call had all
24 target records in its nominated frontier. The numeric ranker input had
the same IDs and pre-rank scores between arms; the new arm returned the
declared `1-rank/(n+1)` scores. All full-frontier scored laws matched exactly,
all packet journals matched, and the independent scorer reported zero
invariant errors. Sequential isolated Recall p95 was 22.53 ms for focus and
23.57 ms for task-lexical, below the frozen 100-ms gate.

This is useful but narrow. The 24 places are unique in this corpus, so the
natural and offset questions may be answered by place identity without
actually converting time. The strong literal result supports timestamp-aware
retrieval on these public records, not generalized temporal reasoning. The
72 questions share only 24 events and one source snapshot; zero observed
losses do not bound rare harm tightly. The rank scores are ordering devices,
not calibrated probabilities. No language model produced an answer and no
live agent, persistent store or concurrent service was tested. This result
advances Goal 5 retrieval evidence but does not complete the goal.

Reproducibility: [source snapshot](usgs-broad-confirmation-v1/source.geojson)
SHA-256 `ed13585365f5f9dc8ee4b400f8044ce7a9e9c0971234d4aa5eba9407f4089a9a`;
[raw trace](usgs-broad-confirmation-v1/current-contract-raw.json)
`f435d8492ef14d34446241e09e751a932c7073ca06156e3b5480bc6c812a6b42`;
[scored summary](usgs-broad-confirmation-v1/current-contract-summary.json)
`473b2213b166f39cf88ddac70830c51b1069c9118dd61e594094e33ff002bd6a`.
The oracle was opened only by the scorer after the raw trace existed. The
runner passed focused `-race` tests and vet. Production, whitepaper and
remotes remain unchanged; all seven whole research goals remain open.
