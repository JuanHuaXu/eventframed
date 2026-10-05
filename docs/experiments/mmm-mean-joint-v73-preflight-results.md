# V73 joint mean/dispersion preflight

## Result: Coherent Finite Model, Not A Quality Rescue

`research/mean-joint-v73-preflight.mjs` completed with 217,333 scalar checks,
maximum absolute defect 1.1879386363489175e-14 (unchanged tolerance 2e-11),
192 two-trial histories, common/independent noise and hazards 0, 1/16 and 1.
The source and results are frozen under `research/mean-joint-v73-preflight/`.
No empirical cohort, private/sealed outcome or reserved seed was collected.

Retain ALL 27 V34 mean maps with their original prior, not a smaller grid chosen
to meet memory or performance. Combine them with V71's three finite N20 urn
families and three noise states: 243 hyperstates. This is a new finite model;
V34's continuous Beta performance does not transfer to it. The extra maps are
fixed toy functions of member index, not learned EventFrame ontology features.
Mean/family hyperstates are static; member-rate transitions are dynamic.

Candidate inference uses adjacent-ratio urn probabilities and normalized
rank-one filtering. The independent reference uses urn dynamic programming,
dense trajectory-path summation and explicit SAME-latent-outcome emission.
Independent member noise is summed jointly in the reference, rather than
reusing the candidate's marginal-first computation. Check joint evidence,
clean/noisy forecast, all 243 posterior states, mean/family marginals,
posterior normalization, both paired branches, all-target tower and actual
versus covariance Gini gains. A same-Y pair is not a second distinct outcome.
One distinct outcome per member leaves dispersion at its prior in the tested
joint family. The largest nonlocal paired clean-forecast change was
0.36337730839728744: a local-only value calculation cannot be reused unchanged.

Both tested baseline pairs preserve the declared initial mixture mean to
floating-point tolerance. They do NOT show an initial-mean defect: baseline
(.25,.925) maps to (.25000000000000006,.925), and (.47,.7) maps to
(.47000000000000003,.7). Conditional calibrated mean is not necessarily the
unconditional mixture mean. The later Go layout checks below also test
full-member initial means; delayed fitted laws still need independent tests.

## Memory Constraint Changes The Next Implementation

A naive single FP64 243-by-22 conditional table for 200 members costs
8,553,600 bytes, already above the unchanged 8,388,608-byte constructor cap
before journals, priors, caches or overhead. Blindly cloning V71 27 times
therefore cannot meet the bound. A packed 1+22+21 table costs 5,702,400 payload
bytes; that is an estimate, NOT a measured constructor allocation or full layout.

Keep the point family implicit, omit the unsupported spike atom in the free
family, avoid a second full next-state vector table and share raw journals.
Compute priors on mutation or use exact immutable shared objects only when
mathematically identical. The spike's calibrated rate differs by mean map;
it must not silently reuse the original baseline emission. Measure all actual
allocation, fit, replay, query, publication and constructor costs at 150 AND
200 members. No frontier shrink, grid pruning or cap relaxation is authorized.

## Measured Go Layout Preflight

`internal/researchmeanlayout` now implements the proposed storage and complete
initial prior. It retains all 27 means, three families and three noise states.
Point-family rate vectors are implicit, current keeps 22 atoms, free keeps 21.
Reserve the full 64-trial-per-member raw journal, member log evidence, conditional
noise odds, next means and global finite-sum/zero-support caches. There is no
second full conditional-state table or persisted duplicate prior table.

Measured constructor allocations are 5,796,424 bytes at 150 members and
7,722,152 bytes at 200, below the unchanged 8,388,608-byte cap. This closes the
initialization/storage-feasibility question, not the full learner resource gate:
mutation scratch, fitted replay and observation queries are not implemented.
It is not RSS, loaded serving, freshness, durability or an empirical rescue.

The independent Go urn-DP checks add 1,326,816 scalar comparisons across sizes
2, 4, 150 and 200: full prior probabilities, mass, conditional means and
unconditional baseline means. Invalid constructor inputs are rejected. Fresh
unit, race and vet commands terminate 0. The archive at
`research/mean-layout-v73-initial/` freezes source/dependency/generated-main
copies, logs and allocation. No updated posterior is claimed or tested.

## Research Grounding And Remaining Work

[Gelman (2006)](https://sites.stat.columbia.edu/gelman/bayescomputation/Gelman2006.pdf)
supports examining sensitivity to hierarchical priors. It studies normal
hierarchical variance priors and does NOT prescribe this Bernoulli grid.
[Adams and MacKay (2007)](https://arxiv.org/html/0710.3742v1) motivates separating
observation modeling from change dynamics. Neither source establishes this
prototype's empirical quality, error control or serving latency.

Next: integrate the full family in an isolated, bounded Go model, independently
audit delayed/missing evidence and observation values, then run the complete
controlled screen. Preserve baseline/current controls and all previous failed
arms. Only after useful controlled results should fresh replication be frozen.
Anti-Pigeon authority, untouched-agent utility, useful splitting and durable
loaded freshness remain separate requirements. All seven WHOLE goals are OPEN.
