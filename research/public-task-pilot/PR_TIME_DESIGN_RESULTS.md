# Puerto Rico UTC/local-time design: no measured headroom

The frozen source is an official [USGS earthquake catalog](https://earthquake.usgs.gov/fdsnws/event/1/)
response for January 7-8, 2020. Twenty-four unique-second events form the
corpus. The first twelve were queried in design; the next twelve remain
unrun confirmation. Each design event had a UTC-ISO magnitude question and
an equivalent Puerto Rico local-time (AST, UTC-4) question. Two additional
events supplied absent-from-corpus design controls. The source, corpus,
questions and oracle preceded all service/model output.

| Design arm | UTC top1 / survival@10 | Local-time top1 / survival@10 |
| --- | ---: | ---: |
| Focus | 1/12 / 6/12 | 1/12 / 4/12 |
| Priority | 12/12 / 12/12 | 12/12 / 12/12 |

The frozen headroom criterion **FAILED**: priority had zero of the required
four local-time top1 misses. All 12 local targets were in its ranker frontier.
No new normalizer is justified by this screen, and the confirmation split
remains untouched. The two absent controls still packed unsupported records.

This does **not** show that the overlay converted local time to UTC. Raw
priority explanations identify `what-lexical-v2`, and the task plan left
candidate states `unknown`. The frozen lexical code tokenizes numerals but
has no timezone conversion. Local and UTC representations retain common
minute/second tokens; those are a plausible sufficient explanation for the
observed ordering. A separate discriminating test would need records with
conflicting minute/second cues, independently frozen before its output.
The present positive result is a narrow retrieval observation, not time-zone
semantics or agent answer validation.

Every design frontier contained the 24 distinct corpus records; the numeric
ranker left candidates unchanged; focus and priority full-frontier laws and
original numeric scores matched; all packet explanations matched journals.
The independent scorer found no invariant errors. Sequential isolated Recall
p95 was 22.04 ms focus and 23.99 ms priority across 26 calls per arm, not a
concurrent or end-to-end latency measure. Race-toolchain compilation and vet
passed for the isolated runner.

Source SHA-256:
`db64a8be09c55e9a330a86452c4e14abc9b3d76ee5e818c0b19203320e6c8e65`.
Raw design SHA-256:
`82c7339d92c6da7072c88b81b6a498198fd3d6548fe156e6b013bec1ff30a1e5`.
Oracle SHA-256:
`91e71c9f82a775ad1d7aaf86d37d7dcee21bfb7dd21ec3cce1de67330370998f`.
All artifacts are under `pr-time-v1/`. The candidate overlay is unchanged,
research-only, and not promoted. Goal 5 remains open.
