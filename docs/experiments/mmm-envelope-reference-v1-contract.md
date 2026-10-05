# Normalized envelope references: frozen diagnostic

Test-only schema, unchanged daemon. Compare existing prepared append with its
service identity unique index enabled against an envelope plus unique per-event
key/source references. One atomic transaction per batch in both. Sizes 50/200,
32 fresh batches, three rotated trials, two arms. Same payload bytes and FULL/WAL.
No staging/acceptance split or weakened durability. Event payload JSON includes
the real Binding field names and padding; do not tune after seeing results.

Capture all ledger Go files and module/protocol hashes. Time full append calls,
including candidate parsing/encoding, and report post-reopen per-event lookup
cost separately. Verify all original bytes, source identities and sequences,
plus exact retries. Unit gates: conflicting/duplicate sources, mixed retries,
late rollback, lost acknowledgment, invalid offsets and missing bodies.

This prototype has no actual learner, service guard, feedback lifecycle,
cross-process crash test or complete retained-history replay. It restores source
uniqueness as a storage obligation, not the complete daemon contract. Accept no
whole-goal or serving claim. All previous failed results remain unchanged.
