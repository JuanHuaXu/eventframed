# Tree input screen: useful dependency fit, overall FAIL

576 fresh independently fitted datasets;3input laws,2targets,3noise levels,
2sample sizes and16fit seeds per cell. Same subset outcome learner and labels
in all arms. Only input weights differ: uniform, histogram, Chow-Liu tree.
Exact finite-population Brier at five predeclared masks. No adaptive observation.

Both primary copied-field gains pass, but1/180 non-harm cells fails. Preserve
the frozen overall FAIL; no threshold adjustment or automatic promotion.

| Low-noise bit target | n | Mask | Uniform | Histogram | Tree |
|---|---:|---:|---:|---:|---:|
| Copied field |64|1|.103423|.048853|.049001|
| Copied field |128|1|.102467|.048246|.048314|
| Independent inputs |64|3|.250805|.263632|.256028|
| Higher-order XOR input |64|3|.226872|.050935|.229707|
| Higher-order XOR input |128|3|.234096|.048071|.237527|

The failed independent-input cell has tree gain-.005222, paired descriptive
mean +/-3.5SE interval[-.010101,-.000343]. Mean harm is below.01 but the lower
gain bound misses by.000101. This is a failed uncertainty-aware screen, not
evidence that mean harm exceeds.01. Tree beats histogram in that cell by.007604.
All null cells pass non-harm; maximum null mean harm is.002063.

The XOR input law exposes a separate structural limit. Pairwise marginals do
not reveal its three-field constraint; histogram can recover the dependency
while the tree cannot. That limit must not be hidden behind179/180 passes.
Full-input forecasts agree across all arms within1e-12, as required.

## Next Boundary

No unconditional input-model replacement. A full spanning tree always includes
eight edges even under independence; an evidence-penalized forest is a distinct
lead worth researching, not a parameter sweep on this failed cell. It would not
solve higher-order dependence. Any proposed family-selection or richer model
must explicitly cover that limitation and keep same-data overfitting under
control. Inspect prior Bayesian family-selection studies before implementing.

This advances diagnosis of goals1/4, not whole-goal completion. The earlier
adaptive histogram failure and coherent-count null failure remain unchanged.

## Verification

- Collector completed576fits in11.66s test/12.019s package.
- Frozen header contains five source/contract snapshots; hashes verified against
  embedded text and current files. All seed/cell/shape/floor/parity checks pass.
- Summary replay byte-identical. Component race checks and vet pass.
- Previously measured tree fit64 cost5.6us excludes outcome fitting and compiled
  marginal tables; this multi-arm collection is not a serving benchmark.

Raw:`mmm-chow-input-v1.jsonl`.
SHA256:`5369f12856ba28da3ce86455dce4f1d358b3ed6a48019c8dd9d91c488404428b`.
Contract:`mmm-chow-input-v1-contract.md`.
Summary/checker:`mmm-chow-input-v1-summary.json`, `research/chow-input-summary.mjs`.
No production, remote or whitepaper changes. All seven goals remain open.
