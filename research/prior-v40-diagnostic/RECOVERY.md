# Interrupted Diagnostic Recovery

2026-10-03. The interrupted exec session 63347 is missing. A fresh process
inspection found no prior-v40 runner, TestPrior, researchdispersion.test, or
go-test process. The only completed command is model-race (exit 0); its
seven tests passed. No diagnostic or normal experimental outcomes exist.

Classification: confirmed orchestration interruption, not a model failure.
No frozen source, candidate, seed, gate, or prior artifact is changed.
`research/prior-v40-resume.mjs` verifies source and completed-command hashes,
then executes the original remaining diagnostic commands with exclusive
artifact creation. Its recovery metadata and helper hash supplement the
original freeze. Normal collection still requires the original frozen source
inventory and completed diagnostic, with no tuning from its outcomes.

Falsifier: any changed frozen source or inconsistent saved command/log hash
aborts recovery. Unrecorded output is not silently overwritten or recollected.
This local repair does not change durable instructions or production.
