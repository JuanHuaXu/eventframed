# Local ledger ownership v13

A new regression initially failed: two independent Open calls could own one
research ledger. This violates the single model/ordered acknowledgment assumption
even though SQLite serializes individual writes. The local repair acquires a
nonblocking lifetime flock sidecar before SQLite opens on macOS/Linux. Existing
symlink paths are resolved first. Close is idempotent and releases after database
close; failed initialization releases ownership. The sidecar is never unlinked,
avoiding different lock inodes for waiting owners. Other platforms fail closed.

The regression now rejects duplicate and symlink-alias owners and permits reopen
after close. Three combined researchmemory/researchledger race runs and vet pass;
three additional ledger race runs include the alias case. Existing actual-child
abrupt-exit recovery tests continue to pass, exercising ownership release on exit.

This is cooperative local process ownership, not a distributed lease or protection
against malicious file replacement. Hard-link aliases, lock-file tampering and
network filesystem semantics are outside the exclusively owned local-path
contract. No daemon database was opened or migrated. Performance artifacts keep
their original pre-lock source hashes; constructor cost is not rebenchmarked.
Service dependency binding and durable loaded serving remain required.
