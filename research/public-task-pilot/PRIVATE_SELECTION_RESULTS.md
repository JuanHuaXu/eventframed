# Private construction-neighbor selection

`SelectPrivateNeighbors` implements the source selector's valid-record behavior:
preserve input order when the list fits, otherwise sort by distance/ordinal, cap
the candidate window (4*maxM at level0,2*maxM above), apply the alpha-squared relaxed
diversity cutoff, then fill remaining slots from unpicked candidates in source
order. Input/output ownership and pair-call budgets are explicit.

Initial focused tests passed three race runs in1.321s. They distinguish diversity
selection from simply taking the closest neighbors, test fill order and the
short-list fast branch, and reject cancellation/budget exhaustion without partial
results.

A separate backend capture tests six combinations: level0/1 and3/16/40 candidates,
with reversed input order, actual query distances, pair distances and configured
alpha. Capture generation passed in0.245s. The private selector matched every
selected ordinal AND distance in order across three race runs (1.350s).

This is selector equivalence on valid records with captured metrics, not completed
insertion or general SIMD/quantized-metric compatibility. It rejects duplicate,
negative/nonfinite input and metric errors conservatively; those are not modeled
as successful source skip/fallback behavior. Neighbor existence/level eligibility
must be enforced by the construction caller. The test intentionally isolates
selection from graph-layer search and connection/pruning publication.

Next compose construction traversal with selection and private reciprocal link
updates, including overflow pruning. Do not substitute top-k retrieval for this
diversity rule or inherit the public search ef floor into construction silently.
All seven whole goals remain open. Production dependencies remain unchanged.
