# Supplementary Audit, Original Failure Preserved

Original FIT collection completed216cases, but its auditor required50nominees.
It failed with54 !=50 before fitting a model or reading any held-out outputs.
`magnitude-failure.json` and original auditor/source freezes stay unchanged.
Service recallOnce uses searchLimit=RecallK*OverfetchMultiplier, whose default
is3. Thus RecallK50 requests up to150records, and this54-record corpus is
entirely nominated. This is ordinary overfetch, not a new graph expansion,
daemon regression or a reason to alter the runtime contract.

Separate `magnitude-supplement.mjs` keeps original audit/formula/selection/CIs
and changes ONLY the fixture frontier54 (length/uniqueness/loop bounds) and
its CLI filename. This is explicitly a POST-HOC technical audit repair.
It cannot make the original prospective audit pass retroactively. A new
supplement freeze covers source and existing FIT raw before any model fits.
The learner, scorer, source imports, fit grid, selection rule and quality
gates remain unchanged; FIT data are reused, not recollected. Held-out data
are still untouched when that supplement and selected model are frozen.
Negative controls test source/packet/law/formula/ID/coverage/journal corruptions
against the actual raw, and every mutation must be non-identity.
