# Empty-residual Recall ablation v12: frozen diagnostic

Date: 2026-10-02. This is an isolation experiment, not a proposed serving
policy or a promotion gate. V10's admitted full Recall has zero stale
rejections but misses the <100 ms offer-to-done p99 target by about 5x.
V11 found up to 300 residual-record point reads per 150-nominee Recall.

Clone the v10 test fixture without changing its private 256D rows, 128
visible write and 128 Recall offers at nominal 4 ms, four Recall workers,
cap-16/16-ms writer, published-LSN Search, read-to-journal admission,
durable journal, top-150 oracle, or timing thresholds. Change only the
test service configuration to `ResidualModeDisabled`. Before the offers,
verify that the fixture has no residual records and record the initial
`ResidualVersion`; after the offers, verify the version did not change.
This ablation models the zero-residual case only. It must never be
interpreted as a valid setting when residuals are present.

Run two normal fresh trials, record counts, stale retries, call and
offer-to-done p50/p99, writer age, view age, exact oracle and journal
checks. Compare against v10's two normal trials without retuning.
Preserve the v10 frozen gates: 128/128 writes and Recalls, zero stale or
semantic violations, writer age p99 <250 ms, call/offer p99 <100 ms,
view age <250 ms. If the ablation fails, say so. If it passes, the only
supported inference is that empty-residual reads consumed material
capacity under this fixture; general residual handling still needs a
version-safe design, nonempty controls and a full fresh test.

Run the focused test with `-race` in correctness-only mode and ordinary
package tests and vet. Do not alter the v9/v10 source, production service
or existing records. Preserve the negative controls in their reports.
