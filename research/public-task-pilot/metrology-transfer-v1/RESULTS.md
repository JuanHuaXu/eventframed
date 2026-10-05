# Untouched Non-Code Retrieval Transfer

2026-10-03. COMPLETED432isolated cases,72constructed public measurement
questions,36records/12families,6whole families per split. Same ECMAScript
FIT-selected lambda1, no refit; baseline/incumbent passthrough/selected and
structured Observe/text CaptureTurn imports. Exact rational conversions and
independent compound/temperature formulas PASS4preparation tests. Reference
URLs, scope, frozen gate and limitations are in [PROTOCOL.md](PROTOCOL.md).

## Results: Transfer Adoption FAIL

| Import / split | Baseline top1 /36 | Lexical top1 /36 | Gains / losses | Target packed /36, both | Screen |
| --- | ---: | ---: | --- | ---: | --- |
| Structured / design |34|34|0 /0|36|FAIL: no gain|
| Structured / confirmation |31|30|0 /1|36|FAIL: regression|
| Text / design |34|34|0 /0|36|FAIL: no gain|
| Text / confirmation |32|31|0 /1|36|FAIL: regression|

Each arm nominates and retains all36records, packs10with default occupancy
and token rules. Zero literal top1 losses; the same one paraphrase loses
under BOTH imports, NOT two independent discoveries. Confirmation paired
family mean difference -0.027778, one-sided95% lower -0.083751, two-sided
95% interval[-0.099183,0.043627], sign p1. Design vector/interval zero.
Six-cluster descriptive t intervals do not establish population non-inferiority;
the preregistered gain screen fails. All four finite screens must pass; none do.
Full per-question/wording/cluster vectors are in [results.json](results.json).

The regressed question asks for one nautical mile per hour, called a knot.
Baseline ranks knot first; lexical promotes ordinary mile per hour because
the latter matches more individual tokens:1/3 versus1/4coverage. Knot still
appears second in the final packet. This is evidence of an alias/composition
limitation of bag-of-words coverage, not an observed negation failure, missing
nomination, token-budget failure, future-data leak or lost text convention.
No held-out rescue weights, special-case aliases or question rules are fitted.

## Integration, Leakage And Bug Audit

Frozen auditor PASS:432cases,15552journal decisions,4320packed decisions;
all corrected laws unchanged across ranking arms WITHIN each import. Both
six-field captures independently match source/expected post-contract fields;
source and target unit distinctions retained. This confirms this compact
declarative-text boundary, not broad conversation extraction or real agents.
Different import modes may have different baseline laws; no false cross-mode
law-invariance claim. Incumbent passthrough packets equal native baseline.

Race suites sparse/fusion/magnitude PASS; focused service race PASS all7
named capture/frontier/ranking tests. Vet PASS. Frozen20non-identity corruption
controls rejected: source inventory/hash, frame/law/score/formula, duplicates,
foreign IDs, import/nomination/retention, warm-cache and output conservation.
Separate POST-OUTPUT [readback.json](readback.json) checks packed ordered
subsequence/top candidate/all10occupancy and rejects packed-order swap.
That21st control is supplemental, not retroactively preregistered.

Collector opens only corpus, questions, old FIT-model and immutable source
inventory; no task oracle/constructor/metadata features or feedback. Evaluator
reads oracle AFTER sealed raw. Prior code model predates transfer freeze,
which predates collector header. No online answer history/shared service state:
each arm/query/import starts fresh; only bounded role-specific embedding memo
is shared. Model/corpus/source/auditor hashes verified before and after run.

Preserve pre-output wrapper guard failure in [PREFLIGHT_RECOVERY.md](PREFLIGHT_RECOVERY.md)
and post-output verifier wrong-hash-target failure in
[READBACK_RECOVERY.md](READBACK_RECOVERY.md). Neither changes tasks, scoring,
raw outcomes, model or gates; old source/negative artifacts remain intact.

## Measured Cost

| Import | Warm baseline median /p99 ms | Warm selected median /p99 ms | Selected callback p99 ms |
| --- | --- | --- | --- |
| Structured |1.673 /3.409|2.216 /3.376|0.0120|
| Text |1.833 /2.944|2.589 /3.821|0.0122|

36-record import median, warm:~0.299ms structured,~2.26ms text. Cold first
document embeddings produce1.124s/0.911s import-tail outliers in baseline/
incumbent; they are INCLUDED in full import timing, not mislabeled extractor
cost. Query embeddings primed OUTSIDE timed Recall.144memo misses,144entries,
16272hits; total cold acquisition3.878s across documents and queries. No
misses inside any timed Recall. Entire collector8.386s, including fresh imports,
query acquisition, serial outputs/fsync and repeated warm arms. Acquisition
is recorded, not hidden; none of these unloaded in-memory timings includes
loaded LibraVDB/SQLite durability or OpenClaw generation/transport.

Magnitude200callback arithmetic benchmark319.3-323.0ns/op,1792B/1alloc,3
repetitions on AppleM4. NOT feature extraction, sorting, packing, embeddings,
stored-journal cost or a whole Goal6/7 latency result. No serving-policy adoption.

## Interpretation And Next Leads

Prior ECMAScript finite retrieval improvement remains valid within its cohort.
It does NOT transfer here: already strong semantic baseline has little room
to improve and lexical substitution loses a compositional unit distinction.
Code-only FIT is not a cross-domain authority for a global lexical override.
This negative result narrows Goal5's evidence rather than falsifying every
possible retrieval correction or the entire EventFrame framework.

Next investigate a predeclared conservative fusion/abstention policy or
semantic alias/contrast representation, fitted across multiple DESIGN task
families and evaluated on NEW untouched outcome-labeled families. Do not
promote raw baseline usefulness or lexical overlap to calibrated confidence,
or tune lambda against consumed confirmation. Compare downstream answers
where alternate records can support derivation; exact-target retrieval alone
does not label every other record objectively useless. General task robustness,
other generators/delays/shifts, valid splitting, learning/freshness, and equal
TOTAL-cost observation remain required. All seven WHOLE goals OPEN/ACTIVE.
Production/private data/whitepaper/branches/commits/pushes untouched.

## Reproduction

Original exclusive-create run: `node research/public-task-pilot/metrology-transfer-v1/run.mjs`.
For replay use a separate isolated checkout/output directory preserving original
artifacts; rerunning in-place intentionally refuses overwrite. Audit-only:
`node research/public-task-pilot/metrology-transfer-v1/audit.mjs research/public-task-pilot/metrology-transfer-v1/raw.jsonl /tmp/eventframe-metrology-audit-NEW.json`.
Model/source hashes must still match. Raw is public factual constructed tasks,
not private conversations or generated personal biographies.
