# Keyed source witness v1: frozen research contract

2026-10-01. Goal 6 continuation after the bound-label component. This is a
research-only admission and post-mutation validation test, not a production
source-authentication scheme or an automatic learner handoff.

An opt-in durable original forecast may carry a `SourceWitness` generated
under a caller-supplied, separately managed 32-byte-or-longer key. Its HMAC
binds the source tenant/journal/event and old snapshot, the durable journal's
query digest, the six 5W1H field values, event identity/times/producer, the
issued feature bits and baseline,
and the exact research feature-contract identifier. Only a key ID, witness
version, and MAC persist; no raw query, field text, or key may enter the
record. The original service admission guard must first validate that these
were the actual pre-outcome values. Older witness-free records remain readable
but are ineligible for this validation path.

Post-mutation validation must check one exact current target snapshot before
and after reading the old journal and current event. It must require the old
journal to contain exactly one matching event and baseline. An absent current
event or changed six-field value yields `retain=false`; changed key, journal,
query digest, features, baseline, binding, MAC or feature contract fails
closed. A snapshot race aborts. The caller must still hold/recheck publication
authority when installing any rebuilt model; a true witness does not itself
authenticate label truth, statistical independence, or future generalization.

Frozen checks before results:

1. In both memory and temporary persistent LibraVDB, original A/B bound
   records validate; after deleting A, A does not survive and unchanged B does.
   After deleting B, B also does not survive. The persisted original survives
   a durable reopen without changing its witness.
2. Unit controls reject wrong key, tampered MAC, changed 5W1H field, changed
   query digest, features, baseline, binding or feature-contract version. A
   non-feature `Content` change must not invalidate an otherwise same record.
3. The witness JSON contains no raw 5W1H values, query, event content or key.
   Measure 200 witness generations in a warm in-process loop; p95 below 2ms
   is a finite component screen, not end-to-end serving latency.
4. Run focused package tests under `-race`, vet, and existing durable replay
   and mixed-mutation regressions. Preserve failures without tuning the gate.

No production OpenClaw instance, production secret, stored user data, or
default eventframed path is touched. Key provisioning/rotation, complete
multi-record durable history transfer, and general-mutation recovery remain
future work even if this screen passes.
An exact same-ID, same-metadata resurrection is not distinguishable without a
store-generated event generation token; run a diagnostic for this case and
report the limitation rather than treating an equal MAC as identity proof.
