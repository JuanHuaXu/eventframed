# Temporal shadow reuse v4

Frozen before persistent run. Repeat future-arrival v3 unchanged except enabling
TemporalReuse on the diagnostic worker. Three on/off pairs, same50 initial
records,64 future arrivals,256 reads,250ms job deadline and bounded numerical
callback. No learner or production authority. Default TemporalReuse remains off.

Carry the request AsOf in the immutable job. Both store implementations check
snapshot and recorded ingestion history under their existing read lock. Reuse
requires every intervening version be accounted for by future-available ingestion,
semantic versions unchanged, and matching runtime/evidence-epoch increments.
Zero cutoff, missing history, semantic changes and overflow boundaries reject.
Old stores without this optional interface retain exact-snapshot behavior.

Keep prior zero-error,10% paired p99, overlap and shutdown-accounting criteria.
Additionally require >=90% accepted jobs complete in each enabled future trial.
Retain all durations/errors even on failure. A pass would validate this bounded
diagnostic case, not a learned pipeline or all backfill workloads.

Tests must confirm future-only reuse, backfill rejection, policy/graph/posterior/
residual/abstraction/agency/contract rejection, unknown-history rejection,
cancellation, zero cutoff and exact-snapshot fallback. No stale serving allowed.
