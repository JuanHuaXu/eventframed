# Wire assertion investigation

The failed race run is preserved. Both lazy and eager owner-boundary cases
failed the compound assertion; mutation and gap cases passed. The runner also
detected a concurrent change to `internal/frame/identity.go`. No prior run is
accepted as an unchanged-source pass, and the concurrent file is not reverted.

Production Recall persists the full shadow report and returns that same report;
packing truncates candidates, not decisions. The existing V37 loaded checker
compares JSON decision bytes rather than Go object DeepEqual. Competing causes
are packing count, decision count, snapshot motion, or nil/empty representation
after JSON round-trip. None is established by the original failure text.

Next fresh run adds diagnostics only, retaining every original condition.
Its patch is confined to the failure body beginning at `t.Fatal("original
durable wire", e)`. Replacing that diagnostic block with the original Fatal
reconstructs this run's candidate source; the freeze records its SHA256.
No runtime, publication semantics, or adoption gate is changed.
