# Per-message quotation-mask reuse preflight

Frozen before candidate execution, 2026-10-05. This is a component lead toward
research goal 6, not seven-goal success or a loaded-serving experiment.

## Boundary and mechanism

Control: published `f3231fab2244c5c6bca8f7f822e5669ac58d13cd` (the repaired,
quotation-abstaining extractor, NOT the unsafe historical extractor). Copy its
source into an isolated checkout. The candidate retains the same masking
function, patterns, confidence rules, capture-span rejection, statement summaries,
and participant resolution. It computes a byte-aligned mask once per immutable
message and reuses it for the five extracted fields and participant fallback.
The cache belongs to one call; no shared mutable or cross-session cache exists.

Cost hypothesis: repeated linear scans and allocations decrease. This does not
remove regex scanning, normalization, identity processing, full-text metadata,
database work or queue latency. Eager masking may regress early-match fixtures
whose assistant input was previously never inspected; retain such regressions.

## Semantic audit

Run the frozen control's entire frame suite. For the candidate, run every reused
semantic frame regression. The old `TestTurnFallbackAuditBaselineFidelity` and
its historical benchmark assert exact pre-optimization source identity and are
not a test of this new candidate; leave those files unchanged and run that source
identity test on CONTROL only. Candidate fidelity instead requires source-hashed
patch review and direct differential equality with a separately compiled copy of
the complete repaired control. Do not skip any semantic assertion or loosen a
scientific gate. Record this distinction in the result.

New differential cases compare full Event values, QueryText and FromText outputs
including source/confidence/provenance/offset metadata, not only ranking text.
Cover unquoted/quoted/mixed/malformed input, Unicode, invalid UTF-8, contractions,
collective references, role changes, empty inputs and input sizes up to 16 KiB.
Identity tests and parallel independent calls remain mandatory. All inputs are
new public synthetic text. No chats, outcome labels, corpora, production services
or whitepaper edits are authorized by this experiment.

## Timing and decision

Use five profiles (early, late, no-match, quoted, collective), sizes 256/2048/16384
bytes per role, and turn/text/query workflows. Run both repaired-control and
candidate functions in the same benchmark binary with alternating arm order,
three forward and three reverse samples per cell, `-cpu=1`, `-benchmem`, 100 ms
minimum per benchmark. Setup is outside the measured component boundary.
Record full raw timings, allocations, source hashes, Go version and architecture.
Report per-cell medians; do not call them loaded latency or statistical proof.

Promotion preflight: zero semantic discrepancies; geometric-mean median latency
improvement at least 5% for turn/text/query separately; no cell with more than
10% median latency regression. Allocation regressions must be reported. A failure
is retained; no protocol adjustment after observation. Passing is eligibility for
later service integration and loaded freshness testing, not completion of goal 6.

## Patch review gate

Confirmed local cause: FromTurn/FromText/QueryText call firstField repeatedly,
which recreates the identical quote mask. Alternative cost sources (regex,
bounded normalization, identity enrichment) remain and are measured, not blamed
on masking. The state is immutable raw source plus optional prepared mask.
Falsifier: any changed source/confidence/offset or increased hidden-data access,
or the frozen per-workflow performance criterion failing.

Production, private data, all seven original criteria, negative historical
results and live working-tree extractor files remain unchanged.
