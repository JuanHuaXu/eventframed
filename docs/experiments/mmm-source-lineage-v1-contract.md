# Research-wrapper event continuity v1: frozen rescue contract

2026-10-01. The source-witness diagnostic found a same-ID, same-fields
delete/recreate false positive on both store backends. The original HMAC
result remains a negative identity result; do not silently relabel it.

Add a per-event mutation ledger to the opt-in publication wrapper only. The
wrapper exclusively owns its backend and already serializes mutations under a
one-token writer gate. Under that same gate, record the post-mutation runtime
version for every successful event Put, PutComposition, Delete,
DeleteComposition and DeleteBefore. Duplicate/no-op operations do not create
an event touch. The ledger is in-memory: a new wrapper starts at the current
snapshot and cannot certify histories issued before its start. No production
store schema, daemon constructor, or default code path changes.

`ResearchEventUnchanged(from,target,tenant,event)` must acquire the same gate,
confirm the target is the current committed snapshot, require `from` no older
than the wrapper's initial runtime version, and reject any recorded touch with
version greater than `from.RuntimeVersion`. A busy/quarantined/closed wrapper
fails closed. The service source-witness validator must require this proof
before it can return `retain=true`; an HMAC match alone remains insufficient.

Frozen checks:

1. On memory and temporary persistent LibraVDB, deletion of A does not revoke
   untouched B; deletion of B revokes B; exact B recreation under the same ID
   stays revoked for the original bound record.
2. A new wrapper around the same backend rejects the old binding even if the
   event still exists with matching fields. Put duplicates/no-op deletes do not
   spuriously mark a currently bound event as changed. Retention and
   composition mutators are covered by focused unit tests.
3. Race and vet checks pass. No store mutation may interleave a positive
   continuity check while the wrapper gate is held. The proof is not durable
   across wrapper restart, out-of-band backend writes, or crash uncertainty;
   those must continue to fail closed rather than claiming goal-6 completion.
