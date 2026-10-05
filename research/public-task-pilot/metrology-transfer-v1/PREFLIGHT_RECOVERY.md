# Pre-Output Filename Guard Repair

2026-10-03. First wrapper invocation EXIT1 before runtime/auditor freezes,
Go tests, collectors, raw outputs or outcome-label evaluation. Assertion:
`!Object.keys(files).some(p=>/oracle|cases\.mjs/.test(p))` failed because the
complete immutable internal Go-source inventory includes existing test files
with oracle in their name. Verification reads source bytes for SHA integrity;
it does not execute those tests as retrieval predictors or open oracle.json.

Confirmed wrapper boundary-check bug, not a ranking or extraction failure.
Repair checks actual artifact suffix `/oracle.json` or `/cases.mjs`;
collector's own readpath/source check remains. Separate service race selector
corrected ResearchRanking to ResearchRank from actual test names, BEFORE
freeze/run. No scientific threshold, model, task, split or runtime changed.
Reproduction: original regex against all internal Go paths. Falsifier for the
repair: any task oracle/constructor in runtime freeze or collector inputs.
All previous source artifacts and preparation freeze remain unchanged.
