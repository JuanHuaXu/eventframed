# Coherent routed observation: component checkpoint

Status: component checks PASS. No new adaptive quality experiment or production
integration is claimed. This implements the first boundary in
[the routed observation proposal](routed-observation-proposal.md), advancing
directions 1, 2 and 4 without completing any direction.

## Implemented contract

The publication owns four copied conditional tables. Construction validates
positive full support, disjoint-child marginal identities and a common input
law. An immutable five-weight snapshot combines those four experts and a neutral
Bernoulli component. Virtual cell access is O(K), K=4, without rebuilding the
3^9-cell table for each forecast. Publication validation and copying remain
O(K * 3^9); this is a bounded nine-bit research model, not a general memory bound.

Candidate acquisition, confidence stopping and the final prediction use the
same mixture. Hypothetical inspection copies routing state and cannot issue
feedback. Only the acquired mask is committed. Feedback updates bank and gate
atomically, once. Reader failure, epoch changes and reentrant calls cannot
partly issue a forecast. This is single-owner state, not a concurrent API.
Source models must not be mutated concurrently while publication copies them.

The archived observer, evidence-routing kernel and all v102 artifacts are
unchanged. Different component input laws and partially available reader fields
are rejected, not silently approximated. This implementation retains one-step
planning and the existing six-coordinate budget and forced first observation.

## Bug hunt and verification

Two failing regressions were observed before repair:

- An uninitialized publication passed snapshot construction. A readiness bit is
  now set only after successful table validation and checked before forecasting.
- Equal final scalar probabilities allowed different acquisition weights to
  pass final binding. Component weights are now checked before committing state,
  including an all-neutral counterexample where scalar comparison cannot help.

Both were new local research-code defects, not evidence of a v102 result defect.

Verification performed:

- Literal full-state enumeration of all 3^9 partial assignments under uniform
  and nonuniform common input distributions; neutral and one-component limits.
- Publication ownership after source replacement; invalid models and weights.
- All 512 latent inputs compared with the archived one-expert observer;
  changing unobserved hidden suffixes leaves the acquired trace unchanged.
- Exact 256-step fixed-mask forecast, bank and gate parity with original routing.
- Duplicate feedback, reader failure, epoch mutation and reentrant-call checks.
- Race-enabled joint, routing, bank and comparative tests PASS (3.610 seconds).
- `go vet ./internal/observationlearners` PASS.

Test command:

```sh
go test -race ./internal/observationlearners -run '^(TestJointObservation|TestEvidenceRouting|TestBrierBank|TestComparative)' -count=1
```

## Performance boundary

Apple M4, darwin/arm64, benchmark suffix -10. Three repetitions, 500ms each,
after the test process completed. No fitting, SQLite, retrieval, network or
daemon serving is included. The acquisition fixture uses neutral predictions
to force the full budget; it is not a representative learned workload.

| Operation | Observed range | Allocations |
| --- | --- | --- |
| Preview snapshot plus one partial forecast | 143.8-145.7 ns | 0 bytes, 0 allocations |
| Copy and validate four-model publication | 700.240-707.462 us | approximately 1,261,568 bytes, 1 allocation |
| Single-model six-coordinate acquisition | 2.141-2.283 us | 528 bytes, 5 allocations |
| Mixture six-coordinate acquisition | 3.543-3.657 us | 528 bytes, 5 allocations |

The mixture adds computation but no measured acquisition allocations. Do not
compare publication cost with per-query cost or infer sub-100ms daemon latency.

```sh
go test ./internal/observationlearners -run '^$' -bench '^BenchmarkJointObservation' -benchmem -benchtime=500ms -count=3
```

Source SHA-256 at this checkpoint:

```text
869e6e2cdede519e1d37953b5bd4a1df11052b5def6503e33d1bd13c31de0f8c  joint_observation.go
9d4ac43bf39580abad24659575b9f2478113cf3edc1dea3721df66ad586bb3a8  joint_observer.go
54cc28faa24a303c343891d63d0c768db4bce862d3c473785e80ec440d3854e1  joint_observation_test.go
```

## Still required

Bounded joint lookahead needs its own literal reference tests. Then freeze and
run a fresh adaptive quality/cost protocol with matched-mask, own-observer,
lookahead, random-view and fixed-view controls. The v102 fixed-view pass does
not transfer automatically to selected observations, delayed feedback, arbitrary
input laws, member splitting or real agent tasks. All seven directions remain
open; production and the whitepaper were not changed by this checkpoint.
