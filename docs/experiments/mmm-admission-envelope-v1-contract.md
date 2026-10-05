# Admission envelope storage feasibility

Frozen diagnostic, not an admission implementation or candidate promotion.
Compare three encodings using the unchanged FULL/WAL prepared ledger append:
per-event rows, one envelope row, and one envelope preparation followed by one
small digest marker. All carry exactly the same ordered synthetic 1024-byte
event payloads. Sizes 50 and 200; 32 batches; three rotated trials per mode.
No parameter sweep, production access, private data or concurrent experiment.

Report total append cost and the staged marker portion separately. Only the
marker is a hypothetical guarded operation: there is no actual service guard
or compatibility check here. Include preparation in total cost. Verify every
payload and sequence after reopen and exact retries/conflicting markers.
Capture ledger source, module and protocol hashes with the raw timings.

The envelope lacks the real per-source SQL uniqueness index and lookup/replay
contract. A digest authenticates byte equality, not evidence truth or validity.
Any advantage is an optimistic storage feasibility result, not daemon speedup,
learning freshness, crash-safety proof or authorization to change semantics.
Compare all modes on the same ledger schema and durability settings. Reopen
is not a power-loss test. Fail rather than silently overwrite output artifacts.
