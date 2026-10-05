# Source index write-cost diagnostic

Three trials, batch sizes 50/200, 32 fresh batches per cell, alternating order
of index enabled/absent. Same phaseFixture payloads and timed prepared append,
WAL and synchronous FULL. Setup and fixture construction outside timing.
Reopen and compare every sequence, key, kind and payload through ReadAfter.
Capture source snapshots, phase samples and all cells; no sample exclusions.

The absent-index arm is deliberately invalid for service admission. A negative
control must demonstrate duplicate-source acceptance under a new ledger key
without the index and atomic rejection with it. Never deploy this ablation.

Question: how much additional row/commit time accompanies the required index
in this isolated workload? This cannot separate expression evaluation from
B-tree maintenance, page layout, or WAL effects. It cannot measure serving
tails or prove that a replacement preserving uniqueness will recover the cost.
Compare all paired trials without claiming statistical population coverage.
