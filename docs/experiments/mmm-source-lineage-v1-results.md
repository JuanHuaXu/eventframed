# Research-wrapper event continuity v1: guarded rebuild result

2026-10-01. Frozen contract: [mmm-source-lineage-v1-contract.md](mmm-source-lineage-v1-contract.md).
The in-process continuity rescue PASSES the principal source-revocation
screen on both memory and temporary persistent LibraVDB, with one failed
duplicate-operation clause. Goal 6 remains open.

The opt-in wrapper records successful event mutations under its existing
writer gate. A source witness is accepted for transfer only when the exact
target snapshot is current, its old journal and present event still match,
and no touch of that event occurred since the original snapshot. The source
validator checks continuity both before and after its store reads. The
research-only rebuild still needs an installation-time publication guard;
these checks do not themselves install a model.

In the two-backend service test, two forecasts were admitted with witnesses,
received guarded labels, were processed, and replayed from a reopened durable
SQLite file. Deleting A retained exactly B for a fresh-epoch rebuild.
Deleting B retained neither label. Recreating B under exactly the same ID and
fields did not revive its old label. Re-wrapping the backend at its current
snapshot, as a simulated wrapper reinitialization, failed closed for the old
bindings; no cross-restart lineage recovery was demonstrated. Focused tests
also covered retention, composition deletion, a busy writer gate, and no-op
deletion.

**Failed clause:** a duplicate `Put` did not record an event touch, but the
current publication wrapper quarantined itself on the duplicate error. The
continuity check then failed closed instead of keeping the otherwise unchanged
event usable. Thus the frozen duplicate/no-op usability requirement is not
met. This pre-existing publication behavior is recorded as a negative result;
no production mutation semantics were changed to hide it.

Focused `-race` tests, `go vet`, and the full `researchmemory`,
`researchpublicationstore`, and `service` package suites passed after the
integration test. This is finite correctness evidence, not a crash-recovery,
out-of-band-write, or loaded tail-latency result. Key provisioning/rotation,
durable event generations, full history transfer, and guarded new-model
publication remain open.
