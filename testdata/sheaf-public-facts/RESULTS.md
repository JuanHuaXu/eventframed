# Signed-Snap Rescue Results

The rescue protocol in `RESCUE_PROTOCOL.md` was frozen before these aggregates
were generated. Each case and topology uses a fresh in-memory EventFrame runtime.
Turns 1-9 are captured, a fixed topology is installed, and recall is evaluated
before turns 10 and 12 are captured. Oracle and source annotations are never
ingested, and no case can consume another case's events.

The benchmark repair replaces the exact repeated turn-10 query with a frozen
paraphrase and correctly labels turn 5's prior answer as relevant. The graph
rescue adds bounded directional `supports` and `supersedes` effects. A
`supersedes` relation maps high source support to a low target graph feature; it
is predictive rank control, not a causal edge. The bounded frontier remains
available throughout.

## Primary confirmation

The primary run uses graph weight 0.05, a 16-event frontier, a three-event pack,
and the deterministic 256-dimensional development hash embedder. The untouched
confirmation split contains eight sessions and 16 queries.

| Topology | Relevant hit | Obsolete hit | MRR | Graph applied |
| --- | ---: | ---: | ---: | ---: |
| current | 100.0% | 37.5% | 0.927 | 72.9% |
| unchanged | 100.0% | 37.5% | 0.927 | 72.9% |
| no graph | 100.0% | 37.5% | 0.927 | 0.0% |
| split only | 100.0% | 37.5% | 0.927 | 60.4% |
| split and rewire | 100.0% | 0.0% | 0.938 | 72.9% |
| harmful overmerge | 100.0% | 37.5% | 0.927 | 72.9% |

`split_and_rewire` passed every frozen condition. Relative to `current`, its
relevant-hit delta was 0.000, obsolete-hit delta was -0.375, and MRR delta was
+0.0104. Six query pairs included the obsolete claim under `current`; none did
under signed split-and-rewire. No query lost a relevant hit.

At the query level, all six discordant obsolete-hit pairs favored the rescue
(two-sided exact binomial p=0.03125). They occurred in five of eight sessions;
at that clustered level the two-sided sign result is p=0.0625. The 95% Wilson
interval is [18.5%, 61.4%] for current's 6/16 obsolete rate and [0%, 19.4%] for
the rescue's 0/16 rate. The frozen pass is therefore real for this fixture but
not broad statistical validation.

Relevant retrieval was already saturated at 100% for every primary topology.
This experiment validates synthetic obsolete-item suppression without measured
recall loss; it cannot establish that signed snapping improves promotion or
general semantic retrieval.

## Weight sensitivity

Graph weight 0.25 was run only after the primary verdict and cannot rescue it.

| Topology | Relevant hit | Obsolete hit | MRR |
| --- | ---: | ---: | ---: |
| current | 87.5% | 12.5% | 0.583 |
| unchanged | 87.5% | 12.5% | 0.583 |
| no graph | 100.0% | 37.5% | 0.927 |
| split only | 68.8% | 62.5% | 0.500 |
| split and rewire | 93.8% | 0.0% | 0.750 |
| harmful overmerge | 100.0% | 25.0% | 0.740 |

The stress run also passed the frozen comparison, but it confirms that graph
weight strongly changes the tradeoff. Production should not increase this
weight from these synthetic results alone.

## Runtime cost

On the Apple M4 test host, the existing bounded 200-node, 400-edge graph
benchmark completed in 59.1-61.3 microseconds per operation across five runs,
with 40 allocations per operation. This isolates in-process graph propagation;
it excludes embedding, database RPC, full recall, publication, and concurrent
tail latency.

## Mechanism controls

- Existing edges without an `effect` retain symmetric compatibility behavior.
- Unit tests verify directional support and supersession without feedback into
  the source node.
- Service integration verifies a negative target rank delta, atomic dependency
  invalidation, and complete rollback.
- `current` and `unchanged` are identical query by query.
- The harmful graph is active rather than a no-graph alias, although its primary
  aggregate happens to match the current graph.
- Graph propagation never nominates an event outside the retrieval frontier.
- Graph members are restricted to events available when the topology publishes.

## Classification

- **Passed in this synthetic fixture:** directional supersession removed packed
  obsolete claims without reducing relevant-hit rate at production graph weight.
- **Validated mechanism:** signed target-only rank correction and rollback.
- **Not validated:** general snap utility, promotion benefit, snap admission,
  external certificate coverage, or production semantic embeddings.
- **Not tested:** certified Bayesian posterior pooling and natural Anti-Pigeon
  shock frequency. This fixed-topology experiment installs no sharing authority.
- **Not causal:** `supports` and `supersedes` are predictive rank relations.

Reproduce both runs:

```sh
go run ./cmd/eventframe-synthetic-snap-eval
go run ./cmd/eventframe-synthetic-snap-eval \
  -graph-weight .25 \
  -out testdata/sheaf-public-facts/results-graph-weight-025.json
```

Machine-readable per-query outcomes, aggregates, and the frozen-rule verdict are
in `results.json` and `results-graph-weight-025.json`.
