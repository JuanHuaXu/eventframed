# Public-Grounded Chain Translation Controls

This directory contains deterministic EventFrame controls grounded in NIST's
published Celsius/kelvin relation `T/K = t/degree Celsius + 273.15`. It contains
no private transcripts, personal identifiers, unpublished research records, or
production IDs.

Each split contains eight query variants for each of three frozen families:

- `translation`: 0 to 100 degrees Celsius and 273.15 to 373.15 kelvins propagate
  through aligned freezing/boiling descriptions;
- `invariant`: an observer-label change is erased before the terminal water
  phase result;
- `divergence-control`: the target trajectory changes an undeclared location
  coordinate at an intermediate stage and must fail locality.

The inputs are generated text wrapped as EventFrames, while the values and
correspondences are grounded in the cited public NIST source. They test a
declared mapping and decision contract. They do not test autonomous mapping,
prove a physical causal chain, or establish general domain translation.

The divergence controls are rejected by the structural pass and therefore carry
`prediction_evaluated=false`; their zero-valued prediction fields are explicitly
not measurements. Translation and invariant controls evaluate the predictive
law with `prediction_evaluated=true`.

On Apple M4, five local benchmark runs of a 32-stage, 256-dimensional audit gave
48.3-48.5 microseconds per warm exact audit with stored vectors and cached raw
scores, 245.7-247.7 microseconds cold, and 34.1-34.2 microseconds for structural
divergence. The warm path allocated about 22.1 KB in 399 allocations. The
benchmark excludes store and remote-embedding latency.

Regenerate inputs and results with:

```sh
go run ./cmd/eventframe-translation-eval -split design -cases > testdata/chain-public-facts/design-cases.json
go run ./cmd/eventframe-translation-eval -split confirmation -cases > testdata/chain-public-facts/confirmation-cases.json
go run ./cmd/eventframe-translation-eval -split design > testdata/chain-public-facts/design-results.json
go run ./cmd/eventframe-translation-eval -split confirmation > testdata/chain-public-facts/confirmation-results.json
```

Source: https://www.nist.gov/pml/special-publication-811/nist-guide-si-chapter-4-two-classes-si-units-and-si-prefixes
