# Audit Log

Confirmed construction issue: see frozen protocol and predecessor source pins.
New prototype is a recommendation authorized for isolated experiments only.
No installed dependency, dirty runtime file, production service or paper edited.

1. Setup repair: dependency copies inherited read-only permissions. First patch
   failed before changing either file. Made ONLY the isolated target writable.
2. Fixture repair: first preflight set ML=0, which HNSW rejects. Raw failure is
   preflight-before.txt; no capability was exercised. Set the ordinary positive
   ML parameter .36; thresholds, cap, dimensions and success gates unchanged.
3. Confirmed lock bug: preflight-lock-before.txt passes four adjacent tests but
   snapshot-while-prepared fails. Old generation reader lock was held across
   WAL, allowing checkpoint serialization to deadlock its own commit. Preserve
   overlay-lock-before.go.txt. Separate mutation ownership from generation
   reader lock; prepare never changes visible state, Commit only switches it.
   This refines the protocol's initial serialization design without weakening
   precommit/abort/reader-retirement requirements. A concurrent mutation cannot
   publish through the reserved writer; a checkpoint reads old state normally.

Still inspect retained bytes, overflow cost, snapshot/reopen and whole service.
No synthetic mechanics label can close untouched outcome-labeled agent goals.

4. Setup repair: moving the fork changed its ../lexer replacement target.
   database-preflight.txt records setup failure, not a capability rejection.
   Copied the already isolated original sibling lexer into the new sibling
   path; no installed module or version was changed. Pin copied sources before
   timing. Do not infer library execution from the failed setup command.
