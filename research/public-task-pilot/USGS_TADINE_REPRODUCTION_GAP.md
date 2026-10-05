# Archived priority-overlay protocol cannot run on current service

The initially frozen [Tadine protocol](USGS_TADINE_TRANSFER_PROTOCOL.md)
specified the July 2019 `ResearchTemporalPriority` service arm and its
`TemporalPriority` / `PacketCalibrationStatus` journal checks. Before any
Tadine retrieval call, compiling the new research runner against the current
service failed: those fields are absent from `service.Config`, `ContextPacket`
and `BayesianJournalEntry`. The archived `calendar_task_lexical_overlay.go`
helper exists under the `researchpriority` build tag but has no call site in
the current service. The archived command itself is untracked research code,
not a runnable baseline at the current source state.

Classification: **confirmed reproduction/interface gap**, not a quality
failure. No retrieval result exists under the original gate. It must not be
declared passed, silently rewritten, or compared numerically to the 2019
result. Restoring a service hook would modify production-owned code and is
outside this isolated continuation.

The source snapshot and 72 questions were already frozen before this compile
check. The separate [current-contract protocol](USGS_TADINE_CURRENT_CONTRACT.md)
uses the currently supported candidate-ranker extension. It tests a different
ranking mechanism and must report its outcome under a different name.
