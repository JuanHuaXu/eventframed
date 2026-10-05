# Chow-Liu input component checkpoint

Implemented only in `internal/observationlearners/chow_input_test.go`, not serving.
This estimates P(X), unlike the earlier failed outcome-tree learner. Outcome
labels never influence the tree or its parameters. Pair pseudo-count .5 gives
consistent singleton pseudo-count1, a deterministic maximum spanning tree,
and a positive product law compiled to512input weights.

Primary algorithm basis and limitations are recorded in
`chow-input-component-contract.md`. The original method and the modern paper
motivate maximum-mutual-information spanning trees, not an accuracy guarantee
for our smoothed finite-data adaptation:
https://arxiv.org/html/2011.04144

## Verification

Race-enabled contracts PASS1.461s, vet PASS. Checked:

- Total probability1, positive support, exact uniform-input limit.
- Flipping every outcome label leaves the entire fit bit-identical.
- Tree acyclicity and Prim objective equals independent Kruskal computation.
- Every selected pair marginal agrees with its smoothed training pair table.
- Exact copied-field fixture recovers agreement probability513/514.
- Empty, oversized and out-of-range training data rejected.

The tests validate algebra on these fixtures, not statistical generalization.
Only pairwise structure is represented. Pure higher-order input dependencies
can be missed even with unlimited data; this differs from whether the separate
outcome learner can represent parity.

## Isolated Cost

Apple M4, darwin/arm64, three300ms repetitions, fit64 plus512-weight expansion:

    5673 / 5633 / 5629 ns/op
    4864 B/op, 1 allocation/op

This excludes fitting the outcome model, conditional-table compilation,
retrieval, persistence, concurrency and serving. Do not compare it directly
with full coherent-count fitting or claim a daemon speedup.
Pair counting is O(N*d^2), dense Prim O(d^2), full-weight expansion O(d*2^d).
The tree itself is compact, but this prototype's compiled table is exponential.

## Next Test

Freeze a finite-data comparison with the existing subset outcome learner held
constant: uniform input, empirical histogram, tree input. Include independent,
copied and higher-order-dependent inputs, plus low/high/null outcome noise.
Check all declared partial masks and preserve full-input forecast equality.
This must precede adaptive integration; cheap arithmetic is not quality evidence.
No parameter sweep on previous failed integration cohorts.

SHA256:

    chow_input_test.go 27ec3ae60019341c03426a2660891794d31f8c22b04d84056ee84571ab85276b
    contract af5b33d0334e37d44b1e5c28646d9f1da466f0faac43a29e772f85f0b501b7b3

All seven goals remain open. No production, whitepaper, commits or remote changes.
