# V49 exact construction rescue

2026-10-04. Isolated computational ablation,118frozen sources, five terminal
commands. Original V48 mathematical policy/gates/labels remain unchanged.
Fixed128-entry exact-mean FIFO cache is local to each moment constructor and
discarded before returning. Strength/mode fixed per invocation; eviction only
recomputes. Original constructor and all runtime methods remain controls.
Epoch replacement deliberately uses original construction; no reset-cost claim.

## Actual measured results

| Measure | Original | Exact memoization |
| --- | ---: | ---: |
| Maximum constructor allocated bytes | 7,886,136 | 7,886,152 |
| Preliminary nine-loop maximum | 399.280ms | 371.046ms |
| Full252-arm maximum | 419.144ms | 383.283ms |
| Mean complete loop | 388.240ms | 353.141ms |
| Mean constructor phase | 77.623ms | 42.969ms |
| Full-loop400ms gate | FAIL | PASS |
| Constructor8MiB gate | PASS | PASS |

Mean complete work decreases9.0406%; mean constructor decreases44.6446%.
All252paired loops improve; minimum measured saving27.621ms. Alternate which
constructor runs first across arms to reduce systematic ordering bias. This is
one controlled run, not a host-independent tail guarantee. Original V48's prior
422.621ms FAIL remains; its preliminary409.151ms FAIL is not erased by this run's
faster original preliminary maximum. Retain ALL measurements.

The complete paired ablation covers SAME28consumed worlds/84cells/252paired
arms/4200snapshots per variant/67200distinct labels total. Original and candidate
agree BIT-FOR-BIT on every field except costs: advice, head forecasts, issued
law, receipts, local/global/scope weights, snapshot laws and pending counts.
Both also match the independently audited original V48 raw trace. A separate
dense-reference audit reran on every candidate arm; independent streaming
readback verifies hashes, cardinalities, costs and exact equality. No new
quality benefit or confirmation arises from reusing diagnostic worlds.

## Scope and residual risks

Race/vet pass; exact-state tests cover narrow/rich, density/moment, shared/private,
two strengths, two hazards, delayed/canceled outcomes, epoch replacement,
invalid contracts, input ownership and full200member eviction. All9future-prefix
checks pass, including head advice. Original V48 owner/global-fence/compact
and adjacent moment tests execute again. No approximate arithmetic, changed
prior, outcome-dependent cache, global cache or hidden serving work.

Cache payload bound22,528bytes is fixed, not corpus-proportional. Allocated-byte
screen is not RSS or full working-set measurement. Timers include construction,
issue, delayed resolution, snapshot and inspection. Initial output buffers
precede timer; fixture generation, serialization, auditing/scoring are separate.
No loaded-serving/freshness/sub100ms or full Goal6 claim follows from383ms for
this offline2400-label workload. All seven WHOLE goals remain OPEN/ACTIVE.

Artifacts: `research/memo-v49-ablation/{freeze,completed,readback,cost-report}.json`,
original/memo JSONL tapes, exact frozen source copies and terminal logs. Generation
manifest proves the collector differs only in constructor/function identifiers.
Normal quality design/confirmation cohorts remain unrun; next run unchanged
all-cell/both-cohort gates, then integrate a qualified policy with loaded serving.
Production/private corpora/whitepaper untouched; no commit, push or deployment.
