# Explicit Hash Targets

2026-10-03. First post-output readback EXIT1 after normal frozen audit and
packed-order corruption tests passed. The supplemental check incorrectly
compared final-readback.checkpoint_sha256 with the prose checkpoint.md.
Observed prose SHA c6723ab278120586ab86ba81ecb8e2db9a37ac0e40574d9519b2b89ff5681dd5
does match that prose document's entry in checkpoint-verification.json.docs.
The5d6d3b10f695d1d6b4b9409d7ae8e1ddd8f0c3cdc389ba8aeebe8e57963b5b95
hash instead identifies checkpoint-verification.json itself.

Confirmed supplemental verifier path mismatch, NOT source corruption or an
experiment/predictor failure. Verified old terminal manifests independently;
only this new supplemental path check repaired. Original runtime/auditor
freezes, tasks, model, raw, metrics, thresholds and completed manifest unchanged.
Additional missing guessed filename adjacent-checks.json is a navigation error;
no missing artifact inferred from that guess and no file created in its place.
Project-local lesson: verify an explicitly identified artifact/hash pair;
never infer the target from an ambiguous checkpoint_sha256 key.
