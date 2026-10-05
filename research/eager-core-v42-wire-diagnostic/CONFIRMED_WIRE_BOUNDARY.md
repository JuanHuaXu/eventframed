# Confirmed assertion error, not eager publication failure

The diagnostic run fails in BOTH controls with snapshotEqual=true, packed=10,
savedDecisions=150, returnedDecisions=150, wireEqual=true and no marshal errors.
The first unequal field is EvidenceGroupKey: opaque transient group identity in
the returned object, empty in the persisted one. Its declaration explicitly
uses json:"-". This falsifies the nil/empty-slice hypothesis and establishes a
test boundary error: object equality is stronger than the durable wire contract.

The correction compares the COMPLETE serialized report, checks journal identity,
durability, snapshot and unchanged150/10decision/packing counts. Seven negative
controls must reject changed law, snapshot, omission, durability, journal identity,
report support flag and nonfinite data. This does not alter production code or
the private field's serialization. Mutation/source/epoch gates stay unchanged.

The diagnostic patch can be reconstructed from its predecessor by the exact
diagnostic-body patch recorded in WIRE_INVESTIGATION.md and the preserved race
log; source hashes in both freezes disambiguate versions. Every failed artifact
remains intact. Next run uses a NEW freeze and label, not replacement output.
