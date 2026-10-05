# Candidate-only retrieval: equivalence passes, load rescue incomplete

Protocol: CANDIDATE_ONLY_PROTOCOL.md. Artifacts: `candidate-screen-results.json`
and `candidate-load-results.json`, load sidecars and fresh derived/authoritative
stores. Overlay source is in candidate-only-overlay-v1.

## Boundary And Verification

Go rejected the initial overlay against GOMODCACHE before executing tests. The
experiment therefore uses an isolated copy of pinned libravdb1.6.13 plus
research-candidate-only.mod and overlay-local.json. Normal go.mod, the module
cache and production remain unchanged. The original rejected overlay is retained
as diagnostic history, not a runnable configuration.

The new explicit ResearchImmutableCandidates method reuses nomination and public
score normalization, returns only IDs/scores, and rejects non-HNSW/non-cosine
collections or ordinal-only candidates. Public Search/Query keep hydration.
The private immutable adapter checks ownership and uses its own vectors for exact
rescoring. This is not a mutable-record or authorization interface.

Same-graph tests compare128 self/random queries over600 records: IDs and scores
match, ordinary search retains vectors/metadata/version, and candidate payloads
remain absent. Bounds, unsupported-index and closed-collection checks pass. Three
full race-enabled researchindex repetitions pass with the research build tag.

The static screen actually calls the new adapter (not the old diagnostic API).
Both eight-partition runs retain zero self misses and100% recall on128 exact-oracle
probes. The one-base controls still miss one self target and have roughly98.1-98.4%
probe recall. The checker recomputes all oracles, scores, merges and source hashes.

## Unchanged Load Screen

| Initial N | Repeat | Read errors | Write capacity errors | Late responses | Successful-read misses | Acknowledged/reopened |
|---|---|---|---|---|---|---|
|800|0|0|0|0|0|512/512|
|3200|0|0|0|0|0|512/512|
|6400|0|0|11|0|0|501/501|
|800|1|0|0|0|0|512/512|
|3200|1|0|0|0|0|512/512|
|6400|1|9|19|0|0|493/493|

Four of six arms pass; overall FAIL. The nine read errors are lease-admission
busy. All3042 acknowledged writes reopen, with no build/shutdown audit error or
failed-write-present record. Final drain is83-315ms. Largest read85.70ms and
write82.36ms; those maxima do not negate rejected operations.

Mean6400 build durations remain123.47/124.14ms. Capacity failures total30 versus24
in the preceding historical run, so fewer payload copies have not demonstrated
a robust throughput rescue. The4/6 versus3/6 pass count is likewise not proof
of causality from unpaired repetitions. No allocation reduction percentage was
measured in this run, even though the explicit hydration branch is omitted.

## Status

This is a semantically checked candidate interface for the private fixture, not
a production API recommendation or a completed performance goal. Further work
must reduce graph construction service time and explain admission clusters,
with fresh controls; do not raise caps or weaken recall. All seven goals open.

Reproduction uses `-modfile=research-candidate-only.mod` and
`-overlay research/public-task-pilot/candidate-only-overlay-v1/overlay-local.json`.
Tests also use `-tags research_candidate_only -race`. No dependency upgrade or
production install/push was performed.
