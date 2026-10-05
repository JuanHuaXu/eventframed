# Frozen full-vector endpoint diagnostic

Use the existing full-vector population solver and the all-genuine .30-noise
risk (mask0 coefficient10) as objective. Keep all176 protection coefficients,
epsilon=.01, interval[.10,.30], baseline, allocations and900 gates unchanged.
Keep5000 sweeps,1e-8 dual gap and1e-12 primal tolerance. No actual evaluation
world is supplied to prediction; .30 is fixed at design time for every world.

This is an endpoint diagnostic, NOT minimax optimization. If the valid dual
bound D for endpoint regret implies -D<.005 in an allocation, no forecast in
this full-vector coefficient-constrained family can pass that endpoint gain.
Store these numerical ceilings including a reporting cushion, without claiming
formal interval-arithmetic proof. If all endpoint gains pass but other worlds
fail, actual balanced/minimax design remains a lead. Do not discard any world.

Repeat complete optimization, original-control and polynomial-risk checks,
alternate scoring, and existing solver unit tests. Record worst pointwise
history/outcome harm separately from population protection. Preserve all
failures and any convergence issue. These are consumed finite model scenarios,
not independent real-data validation. No production work or publication.

