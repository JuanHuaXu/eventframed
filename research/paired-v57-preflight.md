# V57 preflight record

Initial module source freeze: 17 compiler inputs and seven support files,
13 test roots pass under race, vet passes, and nine batch microbenchmarks
complete. No cohort runs in that freeze. Output/source copies retained in
`research/paired-v57-unit-preflight`. Do not imply these numbers validate
external prediction quality or loaded serving latency.

Subsequently added the independent full-fixture reference and its cross-check
to the 72-origin unit audit; this is a new source version, not a rewrite of the
earlier unit artifacts. Full dispatch will re-freeze and rerun all relevant
tests. The initial snapshot copies preserve the former version.

Full fixture mechanically clones V56, adds only predictive-value selection,
stores all 150 candidate values, and adds independent value auditing and a
value-corruption rejection. Controls retain original laws, selection arithmetic,
cohorts, delay and missing-source treatment. Shape changes are explicit: eight
arms, 24 per world, 960 overall. Original 840 control arms must still match
V54 bitwise excluding costs. No production or baseline-file edits.

Raw full outputs, source freeze, command logs, metric readback, and baseline
comparison are required. No full-study verdict may be inferred from preflight.
