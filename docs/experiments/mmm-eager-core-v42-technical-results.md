# Eager immutable publication V42: technical results

2026-10-03 local. Goal6 component preflight, NOT a loaded latency success.

`research/eager-core-v42-snapshot-selector-repair/completed.json` records
terminal race/vet PASS, unchanged source hashes at execution. Four root tests
include seven wire-corruption controls, lazy/eager owner-held reachability,
future/visible source-motion preservation and cancelled/gapped/interrupted
publication rejection. Eager preparation permits nomination while owner is
held; BOTH controls still defer returned packets to original FULL durability.

## Preserved investigation

- `eager-core-v42-initial`: compile FAIL on two invalid receiver selectors.
- `eager-core-v42-selector-repair`: actual wire assertion FAIL and independently
  detected concurrent runtime source change. User changes were NOT reverted.
- `eager-core-v42-wire-diagnostic`: both controls have snapshot equality,
  10packed/150saved/150returned and byte-identical decision JSON. Object equality
  fails ONLY on EvidenceGroupKey, intentionally `json:"-"`. The nil/empty-slice
  hypothesis is falsified; this was a test-boundary error.
- `eager-core-v42-wire-boundary-repair`: compile FAIL on the new negative
  control's Snapshot.Epoch selector; actual field is EvidenceEpoch.
- `eager-core-v42-snapshot-selector-repair`: PASS checks the COMPLETE serialized
  report, identities, durability and unchanged150/10counts; mutations to law,
  snapshot, omission, durability, journal identity, support and NaN are rejected.

No failure artifact is overwritten. Repair notes give source reconstruction
information. A passing unchanged-source invocation does not retroactively turn
the earlier changed-source/failing runs into passes.

## Measurement limitation and next check

Post-preflight audit identifies a measurement gap: eager construction adds
publisher-owner work and can build a head which no request consumes. Therefore
the old read-build+hit=129 accounting cannot simply be reused for eager arms.
The CURRENT prototype separately counts attempts, preparation owner acquisitions
and elapsed preparation, including failure/same-head work. This instrumentation
is newer than the first passing freeze. Its OWN unchanged-source race/vet run
now passes in `research/eager-core-v42-publication-accounting/`: eager owner
control records2attempts/2publisher-owner acquisitions/1prepared core/0errors,
while disabled control records zero preparation work. Those race-run times are
NOT a calibrated latency benchmark. Exact1318-source preflight snapshot is
preserved in `research/eager-core-v42-checkpoint-publication-accounting/`;
two unrelated research sources were recovered from their previously preserved
versions instead of reverting the current workspace.

All physical preparation must count, including unused cores. Full offered-load
constructor/copy/durability/writer-return cost and the original100/250ms serving
and freshness gates remain required. Do not equate a held-owner technical barrier
with demonstrated p99 benefit. V37's ALL16 negative loaded outcomes remain valid.
Production, private data, deployment and whitepaper remain untouched by this work.
