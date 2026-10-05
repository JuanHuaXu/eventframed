# NASA mission-date retrieval transfer v1

**Decision: FAIL the frozen sparse-promotion screen.** Run 2026-10-01 on an
isolated in-memory EventFrame service with local Nomic embeddings. This is
public structured QA retrieval, not an LLM or agent-answer trial. The twelve
fact identities and twenty-four questions were frozen before the run; the
runner did not read the oracle. Raw results were written before scoring.

| Arm | Top-1 all | Top-1 literal | Top-1 paraphrase | Support in pack 10 |
| --- | ---: | ---: | ---: | ---: |
| Service baseline | 15/24 | 12/12 | 3/12 | 23/24 |
| Direct lexical | 16/24 | 12/12 | 4/12 | 23/24 |
| Frozen sparse | 16/24 | 11/12 | 5/12 | 23/24 |

Sparse gained two paraphrase top-1 supports (`dawn-ceres` and
`osiris-return`) over baseline, meeting the +2 paraphrase clause, but lost
the literal `dawn-vesta-departure` top-1. Its net top-1 gain is one, tied
with the simpler lexical control, and the no-literal-loss clause fails.
`osiris-launch-paraphrase` had its support in all three 12-record nominated
frontiers but outside all three ten-record packets. That is a ranking/packing
miss, not nomination failure. Every arm returned 24/24 stored Bayesian
journals. Full-frontier corrected-law maps and nominated identities were
structurally equal by case across all arms; these rank-only interventions did
not change the scored forecast law.

Local 24-query arm timings, measured only around Recall after in-memory
capture, were: baseline p50 17.01 ms / nearest-rank p95 18.03 ms; lexical
16.81 / 19.62 ms; sparse 17.28 / 18.48 ms. These small sequential samples
cannot establish population tail latency or an overhead bound.

The independent unit is the mission cluster: New Horizons, Dawn, or
OSIRIS-REx, only three units. Corpus sentences and questions are constructed
from NASA primary-source dates, not natural agent conversations. Support
top-1 is a deterministic date-answer proxy because each fixture has one date;
there was no provider-visible answer, abstention, delayed feedback, or
production backend. The set is now consumed for tuning. Do not promote the
sparse override or count this as Goal 5 completion.

Reproduction from the project root:

```sh
go test ./cmd/public-nasa-transfer ./internal/researchsparse ./internal/service
go test -race ./internal/service ./internal/researchsparse -run 'TestSparseDirectRankingBeforePacking|TestResearchRank' -count=1
go run ./cmd/public-nasa-transfer NEW-RAW.json
jq -n --slurpfile raw NEW-RAW.json --slurpfile oracle research/public-task-pilot/nasa-transfer-v1/oracle.json -f research/public-task-pilot/score_nasa_transfer.jq
```

The third command requires the declared local Nomic digest, writes only a new
raw file, and never opens `oracle.json`. Score its `results` against the
separate oracle only after the run. Frozen SHA-256: corpus
`dcac3e4a2180de2b519dd795f647591a087ee47bfcc0cb3d3c00664e931dfdf6`,
queries `d28badf58723efba95998e670fc9283df00419bdf31cd19cb9b86d2fe201b2b1`,
oracle `59f4ac5d193924023162a47b494fcb295d9ccd8464b2ff63bd19aa6f47bfcfa5`,
weights `33fc2b941b92374a3bcba12f1f1f6885042134cc4fd151b121817ccdba897f07`,
runner `d9428913f691b11af87ad97b27e07972d679a073201b2634977213c2f6ba35b2`,
and raw output `d55c28f59798f85b37b907efda88a35d1884620d8152e38062942b3a88ec34b9`.

Primary sources: [New Horizons](https://science.nasa.gov/mission/new-horizons/),
[Dawn quick facts](https://science.nasa.gov/mission/dawn/toolkit/quick-facts/),
[OSIRIS-REx in depth](https://science.nasa.gov/mission/osiris-rex/in-depth/).
