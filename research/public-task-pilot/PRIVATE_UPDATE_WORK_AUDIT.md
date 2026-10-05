# Private ANN update work audit

## New finding

The captured count of 31-113 changed logical records is not a bound on the work
required to discover those records. Reading the pinned backend shows two separate
costs that a copy-on-write adaptation must preserve or replace explicitly:

1. `delete.go:443` replaces a deleted entry point by scanning `h.nodes.Len()`
   registry slots for the highest-level surviving node, breaking ties by ordinal.
   This cost depends on registry length, including holes, not only live records.
2. `delete.go:223` repairs each level's neighborhood using a D-by-D float32 matrix
   and D(D-1)/2 pair distance calls before subsequent reconnection/heuristic work.

`audit-delete-work.mjs` derives neighborhood sizes from the before-state captures
using the source's outgoing-plus-valid-incoming construction. The output pins
source, input, and harness hashes. This is a static single-writer work estimate,
not an instrumented count or benchmark. Cross-level mutation is not assumed to
change another level's adjacency. Concurrent updates would require a new trace.

## Derived results

Across the 16 captured deletions:

- 766-2266 pair-distance calls in the matrix construction alone.
- Up to 47 neighbors at one level.
- 6340-18828 float32 matrix bytes summed across levels, excluding arena alignment,
  scratch buffers, and any additional heuristic distance evaluations.
- The final N6400 deletion changes global state and invokes the entry-point
  replacement path. The captured ordinal span implies at least 6408 registry
  slots scanned; the actual registry may be longer.

The preceding 138-139us supplied-edit benchmark excludes all of this work.
Consequently it cannot support an end-to-end update-latency claim.

## Implementation implications

Private preparation must account for read/repair work as well as changed records.
Silently stopping repair at 128 changed records or treating EfConstruction as a
strict total-work limit would change the algorithm without validation.

For entry-point replacement, a persistent highest-level/minimum-ordinal summary
could eliminate the registry scan while retaining this deterministic selection
rule. That summary must be maintained on insertion, deletion, and level changes,
and belong to each historical root. It is a proposed adaptation, not implemented.

For neighbor repair, first instrument actual distance calls and touched/read
nodes before moving the routine to a private overlay. A private record accessor
alone is insufficient: current insertion also mutates ID registries, vector
storage, node allocators, in-flight state, size, and entry-point state. Direct
atomic writes in deletion cannot be intercepted by changing one link setter.

The next private-update experiment should therefore use explicit transactional
state access and measure discovery-plus-preparation, followed by durability and
the original sustained offered-load test. The existing owner still lacks a byte
budget and nonallocating durable commit. All seven whole goals remain open.
