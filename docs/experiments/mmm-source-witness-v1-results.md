# Keyed source witness v1: component result

2026-10-01. Frozen contract: [mmm-source-witness-v1-contract.md](mmm-source-witness-v1-contract.md).
This research-only component PASSES its keyed binding, privacy, and cost
screens, but FAILS as a standalone event-identity certificate.

- The HMAC binds the admitted journal/event, six 5W1H values, query digest,
  feature bits, baseline, snapshot, and feature contract. Unit controls reject
  altered inputs or key. A non-feature content change remains valid by design.
- The persisted witness contains only version, key ID, contract ID, and MAC;
  the JSON test excludes raw field/query/content text and key material. This
  is a fixture-level leak check, not a general side-channel or key-management
  audit.
- Both memory and temporary persistent LibraVDB retained an untouched B
  after A was deleted, and revoked B after its own deletion. Original records
  and witnesses replayed from the durable SQLite file.
- The diagnostic found that deleting and then recreating the same ID with
  identical fields can pass the HMAC and current-event equality checks on
  **both** backends. The MAC authenticates recorded bytes, not uninterrupted
  event identity. This is a negative result, not an accepted source proof.
- With `EVENTFRAME_WITNESS_COST=1`, three repetitions of the 200-generation
  in-process screen measured p95 809.792, 558.667, and 584.958 microseconds;
  all passed the frozen 2 ms component threshold. This is not loaded service
  latency.

The [event-continuity rescue](mmm-source-lineage-v1-results.md) adds a
research-wrapper mutation history to reject same-ID resurrection. It does not
turn the witness alone into a durable generation token. No production key,
OpenClaw session, or default daemon path was used.
