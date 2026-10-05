# Relation-witness packet selector: component feasibility

Status: research-only component pass against the frozen pair decisions, not
Goal 5 completion or production promotion. Protocol:
[`RELATION_WITNESS_PROTOCOL.md`](RELATION_WITNESS_PROTOCOL.md). The input
hardening is recorded there as a post-freeze amendment.

## Decisions

The fixture-supplied relation identity resolves the previous ambiguity in the
fixed-order packet mode: a spacing-only repeat with one declared relation ID
occupies one slot, while `A > B` and `A < B` with different IDs occupy two.
Same exact bound keys cannot be split by inconsistent IDs. Missing or tampered
bindings, different origins and unsupported selector modes fail back or closed
as declared. Certified Anti-Pigeon splits and explicitly distinct observed
occurrences retain precedence. A 50-case fixed-order parity matrix checks
the unbound selector's IDs, tokens and suppression accounting; a separate
bound-input parity check covers unchanged relation decisions.

The relation ID is supplied by the test fixture, not discovered or
authenticated by EventFrame. Retaining two contradictory records is not
evidence that they provide independent corroboration. No downstream forecast
law, agent answer or user-visible retrieval benefit was scored.

## Component cost

Command: `go test ./research/originbinding -run '^$' -bench
BenchmarkRelationWitnessFixed -benchmem -benchtime=300ms -count=2` on an
Apple M4, darwin/arm64. The fixture has 50 or 200 unique events, pack cap 20,
and registry construction outside the timed loop. Values below are the range
across the two runs, not latency quantiles.

| Candidates | Raw legacy | Bind all + legacy | Witness selector |
| --- | ---: | ---: | ---: |
| 50 | 3.879-3.884 us, 29,992 B | 150.945-150.986 us, about 244.6 KB | 83.839-83.937 us, about 138.3 KB |
| 200 | 13.068-13.078 us, 46,376 B | 490.741-498.464 us, about 828 KB | 112.207-114.935 us, about 158.3 KB |

The fair comparison for an already requested origin-binding workflow is
"bind all + legacy" versus "witness selector". Raw legacy is a lower-cost
control that does not verify fixture bindings. The witness selector checks
all identities but hashes/describes only reached packet candidates. It is
still allocation-heavy and has not been tested under daemon concurrency,
remote storage, write load or end-to-end agent service.

## Audit

`go test -race ./research/originbinding ./internal/packing
./internal/epistemic -count=1` passed. `go vet` on the same packages passed.
The bug hunt added rejection of empty identities and negative token estimates
and a subtraction-based budget comparison. The registry copies supplied
bindings, and each usable witness is checked against the full event payload
at selection time. These checks do not authenticate the supplier.

The next experiment needs a capture-side, time-local source witness and
outcome-labeled agent tasks with an untouched confirmation set. The current
selector receives an already nominated candidate list; it neither enforces
as-of nomination nor proves the capture relation ID was available before the
answer. Those are explicit leakage and authority boundaries, not implied
passes. Production, whitepaper and remotes remain unchanged; all seven whole
research goals remain open.
