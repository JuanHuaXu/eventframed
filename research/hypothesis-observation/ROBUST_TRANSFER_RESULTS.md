# Conditional transfer guard: protection rescued, gains incomplete

The [frozen protocol](ROBUST_TRANSFER_PROTOCOL.md) applies a finite ambiguity-set
guard to the unchanged hierarchical proposal on all10 allocations and16 masks.
The overall screen FAILS at175/180 gates. Unlike the unguarded model, all160
protection gates and all10 false-confidence gates pass. Five of10 all-genuine
gain gates fail the unchanged0.005 minimum.

## Mechanism and limits

Let p0 be the local forecast, p1 the hierarchical proposal and d=p1-p0. The
largest allowed move p0+lambda*d satisfies

`||d||^2 lambda^2 + max_q[2 d dot(p0-q)] lambda <= .01`,

where q ranges over ALL declared counterfeit-pattern conditional target laws
consistent with the observed reports. The selected true pattern is never passed
to the guard. Only exactly impossible patterns are excluded. This is our direct
Brier derivation using the broad ambiguity-set idea discussed by
[Rahimian and Mehrotra](https://arxiv.org/abs/1908.05659), not an inherited
statistical coverage or authentication theorem from that survey.

If the actual conditional law belongs to this set or its convex hull, the
inequality bounds conditional expected regret against p0, and integration
preserves that bound. The required premise is substantial: known likelihood
family and prior, correct conditioning, no omitted source modes. This does not
bound individual realized loss or establish source authenticity in real data.
It does not grant a corresponding classification-accuracy guarantee.

## Complete stress result

Maximum exact population Brier harm across160 cells falls to0.005473 from the
unguarded maximum0.011908. All earlier seven harm counterexamples are covered;
none was removed. In exchange, genuine-data gains range from0.003703 to0.005537.
The five failing count allocations are[1,1,2,2], [1,1,3,1], [1,2,2,1],
[1,3,1,1] and[2,1,2,1], ordered by types[0,1,2,7].

The guard clips5,592 of10,240 enumerated report vectors. That unweighted count
is not a production activation rate. Probability-weighted retained correction
ranges from about68.4% to100% across the true-law cells. The smallest individual
lambda is0.203884. No altered prior or selected true mask supplies those choices.

## Verification

The quadratic solution agrees with independent direct-loss bisection on3,000
random simplex tests, maximum lambda difference2.43e-12. Tests include zero
budget, identical forecasts, wholly beneficial correction and invalid inputs.
All37,152 supported conditional-law checks satisfy the0.01 regret bound within
roundoff (maximum computed0.010000000000000231). The full guard output replays
byte-for-byte, and its unguarded controls match the independently verified
allocation artifact.

Artifacts: [all guarded cells and gates](guarded-renewal-allocation.json),
[component/replay verification](guarded-renewal-verification.json).

Constructing all conditional laws is not free: counterfeit-pattern enumeration
grows exponentially with measured source types. The final line guard is O(M*C)
for M laws and C classes; no serving benchmark or production integration is
claimed. This changes the output forecast only, not the latent source posterior.

## Next lead

The current guard only shortens one straight-line correction. A closest-feasible
forecast in the intersection of the simplex and the same conditional-regret
constraints might preserve more useful change without enlarging the risk budget.
That is a geometry/optimization hypothesis, not a promised gain. It must retain
all180 tests, prove solver feasibility and account for its additional computation.
Do not retune epsilon to make the failed gain cells pass. All seven directions
remain open; production and the whitepaper are unchanged.
