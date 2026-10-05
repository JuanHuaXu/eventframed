# Batch conversion parity v1: frozen protocol

This is a Goal 6 test-only extension of the real-store batch intent prototype.
Do not change `Service.CaptureTurn`, `Service.Observe`, the production backend
constructor, or shared publication/lineage APIs.

On a fresh control LibraVDB store, submit four same-tenant raw turns through
`Service.CaptureTurn` with a deterministic four-dimensional test embedder. On
an independent fresh prototype store, explicitly perform the same
post-contract `frame.FromTurn` conversion, validate the EventFrame, embed its
`FrameText()` (not `Content`), and derive the v1 raw-turn digest. Submit all
four prepared writes in one authority-protected batch. Compare the durable
event, vector, corpus frame and raw-content metadata for every ID against
the control, before and after reopening the prototype. Match duplicate
decisions through an exact raw-turn retry and reject a changed raw turn.

Negative controls must establish that embedding the full raw content would
yield a different vector and that the EventFrame digest is not a substitute
for the raw-turn digest. Report initial and final runtime-version deltas;
full snapshots may differ because `Service.New` binds its Bayesian policy
before control ingestion. Run the focused test with `-race`, package tests,
and `go vet`.

This tests deterministic capture conversion and storage parity only. It does
not show a batching queue, per-call acknowledgement timing, external index
parity, mixed `Observe`/`CaptureTurn` traffic, production integration, or
loaded Goal 6 freshness.
