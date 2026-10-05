# Rejected V48 compact generator

The first compile failed before any quality experiment. The AST renamer changed
the selector `errors.New` to `errors.newCompactV48`, and Config/Receipt/finite
declarations collided with the original package declarations. Inverse lexical
syntax equality did not reveal this: the wrong renaming was reversible too.

Preserved exact prototype generator, output and metadata:
`compact-v48-rejected-generator.go.txt`, `compact-v48-rejected-generated.go.txt`,
`compact-v48-rejected-generation.json`. The local repair distinguishes selector
members from package-local identifiers and namespaces the complete local
support declarations. It must pass compilation plus paired two/three-expert
behavior tests; inverse syntax remains necessary but is not sufficient proof.
No production, durable instruction or original filter change was made.

A subsequent revival assertion incorrectly expected exact0.5 after symmetric
decimal advice. Actual Float64 `1-(1-1e-6)` is not exactly `1e-6`; the observed
weight0.49999999953983026 matches the resolved advice. The original V43 tests
already document this. Preserve `hybrid-v48-before-extreme-reference.go.txt`;
the corrected test uses an independent closed-form log calculation, with the
original comparison tolerance unchanged. No candidate arithmetic was changed.
