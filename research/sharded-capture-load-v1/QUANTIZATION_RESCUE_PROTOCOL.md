# Quantization declaration persistence rescue

Frozen before this repair, 2026-10-05. The discovery-only candidate reached the
original topology test's post-reopen assertion and failed with a nil error:
Sharded=true but HasQuantization=false. Quantization loss is now CONFIRMED,
not inferred from the earlier discovery error. The library's engine config
omits quantization and both reload constructors omit it; the index recovery
bridge also reconstructs without it. This is a creation-to-durable-config
transition failure, not a one-off stored-row repair or a sharding-only symptom.

Invariant: a newly created collection's validated quantization declaration must
survive WAL/checkpoint/reopen, including recovery-index construction. Derived
indexes can retrain from retained raw vectors; HasQuantization is a declaration,
not proof of trained quantizer state or RAM compression. No old database with
missing metadata can be retroactively repaired by assuming an unknown config.

The temporary fork extends the existing version-2 length-prefixed optional
config block, without changing global codec versions. Preserve declaration and
graph fields, and use a placeholder declaration field when quantization is the
only extension so previous readers can skip the tail. Previous readers do NOT
understand the new quantization field; opening/re-writing with an old reader is
not a supported compression-preserving downgrade. This is research only.

Before repair: a new library regression creates ordinary/sharded collections
with none, scalar and FSQ declarations, inserts records, and closes/reopens;
the positive quantization declaration must fail to survive. After repair:
repeat with race detection, retain all original topology assertions, run codec
roundtrip/invalid-config guards and existing storage/library regression suites.
Cancellation, record metadata and logical-listing coverage stay intact.

No fixture, source-pin or timing gate is weakened. This additional dependency
change is not part of the old frozen one-line candidate. A repaired load pilot
must record new fork source hashes and cannot claim untouched confirmation.
