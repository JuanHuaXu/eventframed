# V6 Identity Boundary Audit

Confirmed P1 upstream cause: immutable provider-free HNSW assigns physical
ordinal by nextOrdinal, not entry.Ordinal. V5 public ordinal-only hydration
assumed equality. V4 external masks/shadow deletions had the same assumption.
Compaction sorted future prefixes before past IDs, changing this accidental
dense ordering. A sparse standalone witness returns physical7instead of1216;
the separate collection prefix witness loses the true nearest past-0007 and
returns past-0020. Neither uses availability metadata or private payloads.

Other explanations were availability-probe crowding, score arithmetic and ANN
topology. V5 common-score/law drift is zero, and the sparse-index failure occurs
without availability or hydration. HNSW's provider-free assignment plus independent
Collection wrong-identity witness establish this invariant first. ANN topology
can still change, so full public equality remains an independent required test.

Patch: immutable physical->external table matches serial base insertion order.
Translate BEFORE external graph filter or shadow tests, and restore public
ordinal from canonical identity before hydration. Overlay ordinals are already
external. No renumbering of durable rows, provider read during navigation,
extra candidates, exact corpus fallback, threshold relaxation or changed labels.
Snapshot/reopen regenerate the map from the same serial build. Unknown physical
ordinal is rejected, never treated as external0. Byte diagnostics add4bytes per
base vector; construction sorting/HNSW cost and copied O(ND) payload remain.

The happy path and prepare/WAL/publication/abort ownership are unchanged. Old
generation tables are immutable while readers hold generation locks. Exact
external/storage identity is required across compaction, deletion, filters and
reopen, not only dense initial insertion. Existing ownership/profile/cold/WAL/
key-capacity tests remain, with ordinary/race sparse and real prefix checks.
The original full-rebuild control also runs the prefix witness as a negative
control. A passing core is insufficient if the public replay still fails.

Related upstream: unchanged18515d4ce0620b24fd2512b0f9023b8b1f4d8355 and prior ten
issue/PR listing. This fixes our wrapper's misuse of the existing backend's
explicit provider-free identity rule, not a new upstream HNSW requirement.
No production/global instruction/cache edits. All seven WHOLE goals OPEN.
