# Bound worker v5 research contract

Status: frozen implementation contract, not a production claim.

Question: can a new research epoch continue learning after a source-validated
component transfer, then resume from its own original-forecast journal without
silently substituting a different transferred model?

The caller rebuilds an Adapter from cutoff-eligible, externally validated bound
labels. It computes a 32-byte keyed commitment over the epoch, seed, target
snapshot, cutoff, source identities, terminal records, and the retained sample
sequence. The key and raw labels are not stored in the new epoch ledger. This
component treats that commitment as opaque and grants no service authority.

For one exclusively owned epoch ledger:

1. The first open may bind the commitment only if the log is empty. The binding
   is durable before any admission. A restart must present exactly the same
   commitment and a rebuilt Adapter for the same epoch and seed.
2. Ordinary cold replay must reject a bound ledger. A bound replay starts from
   the rebuilt Adapter and applies only this epoch's original admission and
   terminal records. It never recomputes historical forecasts.
3. Invalid identity, changed commitment, out-of-order feedback, missing
   admission, corrupt record, or changed transferred sample set fails closed.
   Absence of feedback never becomes a negative label.
4. Reopen with identical source evidence and the same journal must reproduce
   the uninterrupted snapshot and next prediction. A changed commitment must
   reject before the worker starts.

This validates the durable epoch mechanism only. The service-side source
validator, commitment construction, store/durable lock order, loaded p99, and
untouched agent-task outcomes remain separate required gates for Goal 6.
