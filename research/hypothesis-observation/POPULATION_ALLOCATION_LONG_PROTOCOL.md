# Extended-work population allocation run

The original500-sweep solver reached gap1.40538e-7, above the frozen1e-8
tolerance, while its repaired primal remained feasible. Preserve that failed
run; it supplies no complete quality result.

Allow5000 sweeps in a separate offline solver copy. No objective, constraint,
quality gate, primal feasibility tolerance or dual-gap tolerance changes.
The original POPULATION_ALLOCATION_PROTOCOL.md applies otherwise. This is
additional computation, not a production performance rescue. Record a further
timeout as failure without claiming completion or discarding any allocation.
