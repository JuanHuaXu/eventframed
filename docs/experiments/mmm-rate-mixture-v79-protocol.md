# Fixed/long/short evidence mixture v79

Frozen before fresh observations. Research only. Retain allocation/augmentation
and all three archived components: fixed .25, long32-rate v76, short8-rate v78.
Use weights(.5,.25,.25), chosen to keep half the initial evidence on the fixed
component and split the adaptive half equally. Do not select weights from a
sweep. At every step average the current component wealth, not their alarm
bits, with these fixed weights. Trigger at100, retain eight starts/sign pooling.

Under the common conditional null each component is a nonnegative
supermartingale initially1. Their deterministic convex combination is also
one, without independence. Ville gives the same1% crossing bound. Component
correlation must not be treated as extra evidence. The fixed component alone
now requires wealth200, so validity does not promise unchanged detection delay.

Same512-step, ten-scenario,512-stream cells; design/confirmation bases2026117901/02,
seed=base*1e6+scenario*1000+stream. Total10240 fresh paired streams. Five arms:
uniform, fixed augmented, long rate, short rate, weighted mixture. Same query
budget and selected observations for the four augmented arms. No phase tuning.

Mixture gates unchanged: sparse128/256 restricted-mean gain>=10% vs uniform,
paired z3.3 lower gain>0, no extra premature alarms; null Wilson95 upper<=.02;
other alternatives at most10 steps mean delay harm; all six alternatives
excess misses<=.01 and v73 simultaneous paired upper<=.02 with alpha=.05/12.
Report both phases; any failed protection cell rejects adoption. Retain weak
and late outcomes, including absolute misses and comparison to the long model.

Tests: component/query parity, direct probability-space wealth vs log-space,
initial unit wealth, positive factors, rejection without partial state changes,
common snapshot dependencies, bounded state and exact replay with source hashes.
Separate balanced/positive microbenchmarks, focused race tests and vet. No real
MMM/agent/target-law-diameter success follows from this isolated finite screen.
