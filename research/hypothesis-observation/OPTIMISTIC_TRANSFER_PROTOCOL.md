# Guarded optimistic target

Frozen before scoring. Same10 allocations,16 counterfeit masks,16-credit cost,
noise20, local baseline, ambiguity laws, epsilon=.01 and180 acceptance gates.
No change to the projection solver or its tolerances.

Replace the hierarchical proposal by the conditional forecast under the fixed
all-genuine mechanism. Always use that same declared optimistic target, regardless
of which environment is being scored. It depends on observed reports and the
declared likelihood/prior, never the actual mask or latent target. Every supported
mask still participates in the risk guard. The candidate minimizes expected
Brier under the optimistic branch subject to worst-family regret constraints:
equivalently project that branch's forecast onto the same feasible intersection.

This changes the decision objective, not evidence likelihoods or posterior
weights. The emitted forecast is not the full Bayesian mixture posterior.
Interpretation: pursue accurate use of genuinely new evidence while bounding
the cost of being wrong about its provenance under the declared finite family.
Its protection guarantee remains conditional on family adequacy.

Keep the original certain/local/hierarchical score controls. Assert that the
optimistic target matches the all-genuine conditional law, but never feed the
true selected mask into projection. Require all180 gates, including useful
gains and false-confidence reduction, not protection alone. Compare to the
previous projected-hierarchical result without reinterpreting its failures.
Retain cycle/gap accounting and deterministic full replay. No production or
whitepaper changes; any pass still needs new cases and adaptive validation.
