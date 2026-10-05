# New public context pilot: temporal relation miss

20 actual retrievals on the newly constructed8-record,10-question public set.
**Baseline7/8 positive top1; bounded7/8. Improvement criterion FAILS.**
All8 positive supports retained. Forecast laws are identical (max difference0);
maximum rank adjustment0.004747989819115062 stays below.005. No generation calls.

Both arms correctly rank historical Pluto, proposed-draft Pluto, adopted Pluto,
draft planet count, adopted count, initial Ceres and2006 Ceres. Both miss the
question asking for Ceres's classification after its initial planet status:
the relevant asteroid-history record is rank3; initial-planet history is rank1.
This is a temporal-relation discrimination failure, not absence of the evidence.
The two unsupported queries return candidates but have no retained support;
no abstention correctness or hallucination result is claimed.

The context corpus was frozen before running either arm. The bounded method
and its weights/scale were unchanged from the previous experiment. This is
still a small hand-designed research pilot, not independent population evidence
or proof of an online learning mechanism. It is now consumed; tuning on this
miss must not be called confirmation. Actual future-feedback learning is absent.

Sources, opened before construction:
- [NASA Pluto facts](https://science.nasa.gov/dwarf-planets/pluto/facts/)
- [NASA Dawn Ceres overview](https://science.nasa.gov/mission/dawn/science/ceres/)
- [IAU proposed draft](https://www.iau.org/Iau/News/PR2006/iau-draft-planet-pluton-definition.aspx)
- [IAU adopted resolution](https://iauarchive.eso.org/news/pressreleases/detail/iau0603/)

Draft statements are conditional proposals, not asserted timeless truths.
Historical/current labels are scoped to their record. Sources substantiate
corpus facts, not the experimental accuracy claims.

Artifacts: `context-v1/corpus.json`, `queries.json`, separate `oracle.json`, and
`results.json`. Verify via `node research/public-task-pilot/check-context.mjs`.
The runner sees oracle only after predictions. Actual CaptureTurn/Recall,
memory store, existing local semantic embeddings; no production/private data.

Next lead: represent relation roles (initial, subsequent, proposed, adopted)
explicitly, with counterexamples where the same words appear in different
relations. Do not solve the Ceres miss with an entity-specific exception or
query-ID lookup. A new held-out domain must test any resulting temporal rule.
