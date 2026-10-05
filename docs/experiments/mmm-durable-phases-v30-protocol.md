# Durable phase diagnostic v30

Repeat unchanged v29 nine-arm workload with monotonic per-group admission,
integrity-readback and discard durations. Admission includes wrapper validation,
retry lookup, cold prediction construction and actual SQLite COMMIT; discard
includes terminal lookup/validation/COMMIT. These are API durations, not isolated
fsync time. Verify their sum fits inside callback duration. No changes to
durability PRAGMAs, labels, caps, retries, record comparisons or scheduling.

Record all raw phases and source hashes in a new exclusive artifact. Use the
dominant measured API to choose the next bounded group-commit investigation;
do not attribute all callback time to disk or delete integrity checks by guess.
Existing age/read screens and failed v29 outcome remain unchanged.
