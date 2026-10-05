# Held-out forest: fixed-mask screen PASS

576 fresh fits, same data budget for uniform, histogram, full tree and held-out
forest. Outcome learner unchanged; forest uses first-half input fitting and
second-half input validation, with no refit. All2primary copied-field gains
and180/180 non-harm screens pass. This is finite component evidence, not a
whole-goal, adaptive-observation or real-task result.

| Low-noise bit case | n | Mask | Uniform | Histogram | Tree | Forest |
|---|---:|---:|---:|---:|---:|---:|
| Copied |64|1|.105048|.049646|.049840|.050400|
| Copied |128|1|.100835|.048405|.048455|.048556|
| Independent |64|3|.250503|.263464|.257375|.250801|
| Higher-order XOR |64|3|.228589|.050548|.233541|.228704|
| Higher-order XOR |128|3|.232848|.048183|.235281|.232848|

Copied-field primary gains are.054648 [.050363,.058932] and.052279
[.048095,.056463], using paired mean +/-3.5SE over16fits. These are descriptive
screens, not simultaneous or anytime-valid population certificates.

The full tree fails3/180 non-harm checks on the SAME fresh datasets. In the
independent n64/mask3 case, forest improves over tree by.006574 with interval
[.000315,.012834]. This supports pruning in this component comparison, rather
than claiming rescue just because a different seed cohort happened to pass.

Protection is not exact neutrality: the worst forest lower gain bound is
-.009058 for copied inputs/noise.05/n64/root mask, with mean-.004275. This
passes the frozen-.01 tolerance but remains a limitation. Higher-order XOR
is still essentially unresolved, while histogram recovers it. No universal
dependency-learning claim or dismissal of that large missed benefit.

## Next

Test the unchanged forest input law within retained-subset adaptive observation,
with matched fixed-observation and own-observer arms. Freeze fresh seeds and
keep original gate/selector/training schedules; no split-ratio tuning. Include
parity4 and dependent cases that defeated histogram integration. The previous
component-to-integration failures make this mandatory before promotion.
Separately retain the higher-order input modeling gap in the research roadmap.

## Verification and Cost

- Collector576fits completed15.57s test/15.944s package.
- Six pre-collection source/contract snapshots verified against embedded/current
  files. Unique seeds, cell counts, metric ranges/floors and full-input equality
  checked. Summary replay byte-identical.
- Component race checks and vet pass. No collection replay claimed.
- Prior measured forest fit/select64 ~20.8us excludes subset outcome fitting,
  marginal compilation and serving. No daemon performance claim.

Raw:`mmm-forest-input-v1.jsonl`.
SHA256:`95a8f2428a0f82bb05c53b302bdfd14cfbd65ec49ad6ac0603213b148dcce025`.
Contract:`mmm-forest-input-v1-contract.md`; checker:`research/forest-input-summary.mjs`.
All seven goals remain open. Production, remotes and whitepaper untouched.
