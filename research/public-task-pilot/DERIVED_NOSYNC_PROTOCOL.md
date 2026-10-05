# Disposable derived-index sync screen

Repeat PARTITION_LOAD_PROTOCOL.md without changing any operation cap, arrival,
deadline, partition count, nomination effort or authoritative transaction.
Only BuildHNSWBase's new disposable database uses DurabilityUnsafeNoSync.
The authoritative.libravdb database retains default synchronous durability.

This is not asynchronous visibility: the derived graph must be fully constructed
before publication. A process restart must discard/rebuild derived graphs from
the authoritative snapshot, never trust a possibly partial no-sync index file.
No production setting changes; isolated research overlay only.

Hypothesis: redundant derived WAL synchronization adds enough build cost to
contribute to saturation. Falsifier: same failures or insufficient clearance
despite dropping only derived sync. Require all six unchanged combined gates;
do not rescue by weakening authoritative acknowledgements. Preserve negatives.
