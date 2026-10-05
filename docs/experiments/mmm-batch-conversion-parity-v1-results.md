# Batch conversion parity v1: post-contract capture component

The [frozen protocol](mmm-batch-conversion-parity-v1-protocol.md) passes its
finite test-only parity controls. Four raw turns with one tenant and
availability time were captured through ordinary `Service.CaptureTurn` on a
control LibraVDB store. Independently, the same turns were converted after
the contract by `frame.FromTurn`, embedded from `FrameText()`, assigned the
raw-turn digest, and committed in one authority-protected batch on a separate
LibraVDB/SQLite prototype. This extends the earlier
[real-store intent result](mmm-batch-intent-prototype-v1-results.md); it does
not install a batch route in the daemon.

For all four IDs, durable EventFrame fields, full raw-content metadata and
hydrated vectors matched the ordinary capture path before and after prototype
reopen. Runtime version advanced by four accepted events on both paths.
The prepared digest was accepted as a duplicate by the control backend,
and exact prototype retry returned four duplicates without advancing its
version. A changed raw turn was rejected as an idempotency conflict.
Fixture checks confirmed that embedding full raw `Content` instead of
`FrameText()` would change the vector and that an EventFrame digest is not
the raw-turn digest. These negative controls make the parity check sensitive
to two plausible conversion shortcuts.

The focused test passed three repeated runs. The full
`go test -race ./internal/researchbatch -count=1 -timeout 3m` and
`go vet ./internal/researchbatch` passed. The deterministic four-dimensional
hash embedder is test-only. This shows same-process conversion/storage
parity for `CaptureTurn`, not a batch queue, external candidate-index parity,
mixed `Observe` traffic, per-call acknowledgement latency, or loaded
4 ms freshness. All seven whole goals remain open; production is untouched.

Reproduce:

```sh
go test ./internal/researchbatch -run '^TestBatchIntentCaptureTurnConversionParity$' -count=3 -timeout 3m
go test -race ./internal/researchbatch -count=1 -timeout 3m
go vet ./internal/researchbatch
```

At-run SHA-256:

```text
954a0e569bc38793fcf7ffe011b9269ddb86d9211a19b48578020f55d18615c8  internal/researchbatch/batch_conversion_parity_test.go
2348077d2a896539c43d69b0aa1567fbb7d3346a186658b544bfb6f701a3c7cb  docs/experiments/mmm-batch-conversion-parity-v1-protocol.md
0e50797c5b2288b0dc269d7bcc79262a681d728ac961ce4537d8da1c49b239e5  internal/service/service.go
eb28db8fdf44fdee8074055f5b31acbd2023f8371b1805dd88fe37060e8ba92b  internal/frame/turn.go
```
