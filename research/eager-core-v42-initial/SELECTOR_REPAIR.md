# Preserved Initial Compile Failure

The initial preflight terminated before any fixture ran. Two barrier-test
selectors called `s.authority` on `*loadEagerV42`, whose authority is reached
through its explicit eager store field, not a promoted load-wrapper field.
Compiler lines163/168, raw race.log and failure.json are preserved.

Confirmed test-wiring error, not a native persistence or model bug. Repair:
`s.authority.owner.Lock()` -> `s.eager.authority.owner.Lock()` and the matching
Unlock selector. Same mutex and unchanged barrier/authority semantics. No gate,
workload, native method or production instruction changes. Reverse those two
exact selectors in the repaired source to recover the initial frozen source
hash, verified before the new preflight. New artifacts must have a new label.
