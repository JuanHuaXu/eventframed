# Frozen input-model comparison

576 independently fitted datasets:3input laws x2targets x3noise levels
x2sample sizes x16indices. Uniform; bit0=bit2; bit0=bit1 XOR bit2.
Targets bit2/paritybits1..4, noise.05/.25/.5, n64/128.
Seed2026092200+law*1000000+target*100000+noiseIndex*10000+n*20+index.

Same subset outcome learner in all arms. Input weights uniform, empirical
histogram with total pseudo-count1, or frozen Chow-Liu pair pseudo-count.5.
Exact expected Brier over512raw inputs on masks0/1/3/31/511. No tuning.
All full-input forecasts must agree to1e-12. Header snapshots sources before
collection. Preserve every cell, including null and higher-order limitations.

Exploratory advancement screen: on copied inputs/bit2/noise.05/mask1,
tree-versus-uniform mean gain>=.005 with paired mean-minus3.5SE>0 at both n.
Across all cells/masks, tree-versus-uniform lower gain bound>=-.01. Also report
tree versus histogram, without substituting that contrast for primary failure.
Fixed masks only; no new adaptive, real-agent, delayed or whole-goal claim.
