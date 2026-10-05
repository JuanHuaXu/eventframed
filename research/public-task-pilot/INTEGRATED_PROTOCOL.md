# Pre-packing sparse integration check

Frozen before running integrated outputs. This reuses consumed pilot tasks for
integration verification, not new confirmation. Arms: unchanged baseline;
direct lexical union coverage; direct frozen sparse probabilities from the
design-only export. No fitting during requests, no changes to source corpus,
extractor, hash embedder, existing ranker, recall50, pack10 or token budget10000.

Use actual CaptureTurn and Recall with fresh in-memory services per query and
arm, as in earlier preflights. The ranking callback must see all nominated
candidates, before packing. Corpus has13 facts: this is a full nominated pilot
frontier, NOT a 200-candidate or million-record validation. No LLM calls.

Record callback frontier count, packet confidence, original backend score,
research delta and full forecast bundle with RankScore separately removed for
cross-arm comparison. Require per-record law equality on shared packed records;
newly promoted records need additional baseline full-pack visibility to verify
their law rather than assuming it from an intersection-only check.

Record support top1 and survival; no presumed pass from offline results. Run all
sixteen queries across all three arms. Oracle opens after every query in an arm
is complete. No production access or feedback. Null/ambiguous tasks have no
single support; do not count them as missing positive answers. Check maximum
and mean absent scores without claiming calibrated abstention.
