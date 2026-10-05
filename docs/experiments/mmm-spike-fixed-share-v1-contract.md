# Delayed Fixed Share: frozen exploratory contract

Reuse the full672-trajectory/172,032-forecast continuous tape, unchanged.
No model refits or candidate selection. Replace only static expert weights
with a two-state forward filter using uniform prior, unit loss learning rate,
and alpha_t=1/t for one-based t>=2. This schedule comes from Section4.2 of
Korotin et al. (2019), not a grid over our failures. State0 is Markov and
state1 the arrival-refitted challenger. At each origin, the transition is
(1-alpha)*identity + alpha*uniform prior.

At issue clock i, filter from origin0 through i using loss factors
exp(-squared_loss) only for j<i whose labels have arrived by i. Apply the
transition even when a loss is missing; missing losses have factor1, not
zero loss evidence. This reference recomputation handles out-of-order arrivals
at their origin. It is not a naive update of today's weight with an old loss.

Retain the same global pending-feedback .01-per-issued-forecast guard. Compare
guarded/unguarded switching against static continuous, reset32, incumbent and
fixed-half controls on whole stream and terminal64. Preserve per-scenario
harms. All data are consumed exploratory data; no inherited regret guarantee
for permanent missingness or the modified guarded output. All original
candidate fitting cost remains chargeable.

Before tapes: enumerate every latent path for small prefixes and compare the
posterior weight, test immediate/delayed/missing/current-label behavior,
alpha=0 static-expert equivalence, equal-expert symmetry, and a naive-arrival
negative control. The policy must be frozen before viewing its tape scores.
Report O(T^2) reference filtering cost and measure it separately from fitting.
No production, whitepaper, commit or push changes.
