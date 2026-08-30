# Public-Fact Sheaf-Inspired Snapping Cases

This benchmark derives compatibility graphs from the sourced conversations in
`testdata/text-public-facts`. It targets EventFrame's implemented bounded
sheaf-inspired scaffold, not a general mathematical sheaf or causal graph.

Each of 32 cases contains:

- a current graph with an injected correction/misconception overmerge;
- a wrong edge to an unrelated public-fact distractor;
- a missing edge to a related but non-interchangeable distinction;
- a directional supersession relation in the expected rescue candidate;
- an active harmful control whose unrelated distractor supports an overmerge;
- a four-member candidate family containing unchanged, split-only,
  split-and-rewire, and harmful-overmerge graphs;
- two chronological design queries and two later confirmation queries;
- fixed comparison obligations and an oracle topology;
- expected dependency closure and explicit falsifiers.

The factual relationships come from the NASA, NOAA, USGS, NIST, and Smithsonian
sources recorded in each case. Only the malformed graph topology is synthetic:
it is a deliberate benchmark perturbation and is not presented as a public fact.

All case, node, edge, event, posterior, and tenant labels are deterministic
dataset-local references required to express graph membership and expected
rewiring. None identifies a person, account, machine, production session, or
external system.

Generate both public-fact datasets in order:

```sh
go run ./cmd/eventframe-synthetic-text
go run ./cmd/eventframe-synthetic-snap
```

`manifest.json` binds this dataset to the SHA-256 digest of the text corpus so a
snap result cannot silently use different underlying events.

Run the fixed-topology retrieval experiment with:

```sh
go run ./cmd/eventframe-synthetic-snap-eval
```

See `RESULTS.md` for the current interpretation and limitations. This experiment
evaluates graph-conditioned ranking, not snap publication admission or certified
Bayesian posterior sharing.
