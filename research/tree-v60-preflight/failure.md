# Initial preflight failure and local repair

Command: `node research/tree-v60-identities.mjs`; terminal exit 1 before any
result artifact was written. Error: `AssertionError: NaN != 0.5000000000000007`
at the recursive/exhaustive pooled-weight comparison.

Preserved initial source: `failed-identities.mjs`, SHA-256
`1d4c4f5e45c0091f14ac6b78d76156efc59169e898fc449e6259856fcc5f50ca`.

CONFIRMED isolated numerical bug, not a production defect. Contradictory W1/W2
under eta=0 gives zero likelihood for every parameter. Subtracting its -Inf
normalizer from -Inf terms creates NaN; multiplying that NaN by zero global
weight does not neutralize it. Entire zero-mass noise branches must never be
normalized. Finite placeholders are permitted only because their posterior
global weight is exactly zero. All-zero whole-model support must still reject.

Local repair precedes any rule change. Adjacent controls: supported eta=.1/.2
branches are fully normalized; the independent joint enumeration must agree;
removal of a discordant factor must revive eta0 support; initial forecast and
same-baseline divergent examples remain checked. No global instructions,
durable skills, old source/results or production files changed.

Second command also terminated 1, at the NONVACUITY test for incorrectly
noise-marginalizing members separately. The test used an absolute 1e-6
likelihood gap, but the two likelihoods are 0.000035535954563260255 and
0.00003646458046752851: a 2.61320096% relative difference. This is a test-scale
error, not a failed equality or scientific rescue gate. Use a dimensionless
>1% relative discrepancy to verify the negative control is material; do not
change the exhaustive-equality tolerance or any experiment requirement.
Preserved second source: `second-identities.mjs`, SHA-256
`9957bd5a217b3f50a7806243c0deddbbd5f831cb399750805571935b2181329b`.
The diagnostic evaluated an in-memory copy only and wrote no result artifact.
