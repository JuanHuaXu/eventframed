# Supplement Setup Failure

2026-10-03. The original frozen auditor failure was reproduced and its log saved.
The new post-hoc supplement then exited1 before checking a trial because its VM
did not expose Node's Buffer object, required by the unchanged projection checker.
The immediate repair exposes Buffer and reuses the original log only after an
exact hash comparison. Frozen auditor/runtime/protocol/raw data remain unchanged.
This is a checker setup failure, not a failed or repeated normal experiment.
