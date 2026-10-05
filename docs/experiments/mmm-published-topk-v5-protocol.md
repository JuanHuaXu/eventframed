# Published-LSN top-k v5: frozen transfer protocol

Date: 2026-10-02. Test-only, opt-in. No production search or writer change.

## Question

Does the normalized certified-LSN SQL path preserve ordinary `Store.Search`
candidate membership and cosine scores at recall caps 10, 50, and 200,
including when as-of-future records are more similar than eligible records?
The v4 five-result fixture is insufficient to establish top-k behavior.

## Frozen fixture

Use the v4 private four-dimensional, declared-payload collection and
normalization boundary. At `as_of = 2026-10-02T00:00:00.120Z`, retain the
two aligned eligible genesis rows, append 198 eligible rows with angles
`theta_i = 0.005(i+1)` radians for `i=0..197` and raw vectors
`2(cos(theta_i),sin(theta_i),0,0)`, and append 16 future rows at
`as_of+5 ms` with angles `0.001(i+1)` and the same raw-vector rule. The
genesis future sentinel remains. All writes go through the v4 test-only
normalizer and certified publication gate. Query with raw `(2,0,0,0)`.

At the published LSN, run SQL `embedding <=> unit_query` with a declared
payload projection, `available_at_sort <= as_of`, and `LIMIT k`. Decode
each EventFrame and convert distance to `clamp(1-distance,-1,1)`.
Compare against ordinary `Store.Search` on the same committed state and
as-of, at `k in {10,50,200}`. Ties may differ in order but not membership.
Warm each arm, then measure 32 quiet calls per arm per cap in alternating
order; report median and nearest-rank p99. These timings include SQL
lease/query/decode and ordinary Search call, but exclude ingestion,
publication, journal, full Recall, and offered-load scheduling.

## Gates and interpretation

Pass only if each arm returns exactly k unique eligible rows, neither
future sentinels nor other future rows appear, top-k membership agrees
exactly, each matched score differs by at most `1e-5`, SQL scores are
nonincreasing, and all 32 measured calls per arm per cap succeed.
Report quiet timings regardless of pass/fail; there is no quiet-latency
promotion gate. A mismatch is a negative transfer result, not a reason
to retune the vectors or metric on this fixture.

The study does not establish loaded service p99, multi-dimensional
embedding performance, full Recall/journal coherence, robust ANN recall,
or Goal 6 completion. It is deliberately a transfer test from v4's
five-result score component to bounded retrieval candidate sets.
