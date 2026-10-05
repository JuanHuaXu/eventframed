# Source identity index v46: primitive and lookup scaling

Frozen before measurement. Opt-in identity index only; no service-owner adoption
or existing default database migration. The unique key is tenant/stream/contract/
service-journal/event. Index and original share the same SQLite insert transaction;
separate identity reservations cannot be acknowledged ahead of original bytes.

Test single/batch/prepared duplicate rejection, exact retry, late-conflict rollback,
scope separation, concurrent source claims, terminal retention, existing duplicate/
invalid-history rejection without data rewrite, wrong-index rejection and process
exits before/after original commit. Lookup must reject missing index, oversized
keys/results and cancellation. Full forecast canonical/provenance validation is
still required by the future owner; the generic ledger is not a truth verifier.

Scaling diagnostic: three fresh histories of 100, 1,000 and 10,000 bound storage
records, one trial each, populated in prepared batches <=200. Then alternate 128
indexed hits and 128 misses per history; preserve raw timings and the SQL query
plan. Require a SEARCH through all five indexed equality fields, not a full scan.
Misses are sql.ErrNoRows, never partial/canceled responses. Record source hashes.
No timing sample is excluded and no threshold is fitted to results. This small
diagnostic does not establish a population tail or billion-record performance.

Index construction/activation may scan history and has separate startup cost;
on-disk index and retained tombstones grow with history. No bounded-retention
claim follows. Point lookup returns at most one bounded original under the unique
index, without building a lifetime Go map. Run unit/crash/race/vet checks before
measurement; no production configuration or published claims change.
