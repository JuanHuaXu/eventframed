# Current-state information acquisition

Frozen before quality scoring. All2688 consumed fixed-forecast v120 trajectories.
Same six experts, switch prior, acquisition pools/clocks/costs as expedite-v120.
Compare against its natural, random, entropy and heuristic-disagreement policies.
Only acquisition utility changes; no expert refitting or extra label budget.

For each nominated unknown origin j, refilter with hypothetical Y_j=0 and1.
Evidence ratios yield p(Y_j=y|available history). Conditional terminal expert
weights yield p(K_t|history,Y_j=y). Score H(K_t|history) minus expected conditional
entropy. Verify outcome masses sum1 and their mixture reproduces the original
terminal expert marginal. Max score, earliest-origin ties. Reveal after current
forecast at t+1. Do not read Y_j or Q to select. Natural later arrival is deduped.

This is exact mutual information for the declared fixed-tape switching working
model, marginalizing past latent states. It is not a universal calibrated model
of the external world. K_t sums active/stopped states sharing an expert; nuisance
state uncertainty is not itself the acquisition target. Model mismatch remains.

Require nonharm lower>=-.01 against all four controls in all168 cells; terminal
delayed changing cases1,2,4,5,7,8,19,20 require mean gain>=.005 and lower>0.
Intervals paired mean +/-3.5SE over32 trajectories, exploratory not simultaneous.
Keep both consumed phases, all costs and failure cases. No tuning after results.

Verify with independent joint latent-path enumeration and as-of poisoning;
assert identical per-run query counts to each paid control, zero queries and
identical scores in immediate complete schedule. Full byte-exact replay.
At most31 acquisitions; each scores<=8 candidates with two O(K*T) hypothetical
refilters. Reference cost O(K*T^2 + acquisitions*pool*K*T), slow-path only. Report
offline wall time without implying serving latency. No fresh evidence claim.
