# Reproduction And Boundary

## Verify Saved Evidence

Run from a relocated artifact directory for saved numerical verification:

```sh
node evaluate.mjs
node independent.mjs
node --test evaluate.test.mjs negative.test.mjs
```

verify.mjs additionally checks all original prospective source paths. Those paths
identify the run, not installed dependencies. Its exported verify(root,
{sources:false}) verifies saved evidence when original scratch paths are gone;
do not claim a current-source readback in that mode. The evaluator independently
recomputes row metrics, not the Go pseudorandom vector oracle. Generated geometry
and public DESIGN question strings have no private session payloads.

## Reconstruct Fresh Sources

1. Copy the pinned LibraVDB v1.6.13 module and lexer v0.1.12 into a new scratch
   directory. Apply the sibling sharded-capture-load-v1/dependency-v3.patch to
   the copied module, following its REPRODUCE.md and original source hashes.
   Never patch installed module caches or production installations.
2. Make fresh control/candidate copies. Apply each arm's patches named in
   SOURCE_ARCHIVE.json with patch -p1 from that copied module root. Every patch
   binds before/after hashes; verify both, not merely patch's exit status. V1's
   archived before-lock source is historical failure evidence, not the repaired
   source used in its cost run. V2 and later include the common aligned-key fix.
3. Install compiled fixtures at their exact paths from SOURCE_ARCHIVE.json only
   when that target appears in the selected arm's prospective SOURCE_PINS.json.
   Tests directly inside the dependency use its inherited ../lexer replacement.
   EventFrame service tests instead resolve the main module's pinned lexer;
   submodule replace directives are ignored by Go. Preserve this boundary.
4. Recreate service code from published f3231fa plus the isolated frame-mask
   candidate and prior public_capture_load/sharded_quality fixture templates.
   Follow sharded-quality-v1/REPRODUCE.md for exact inherited helper/source pins.
   BOTH V5 service arms are unsharded; only dependency module replacements differ.
5. Use a NEW output directory and path/source manifests. Run the exact explicit
   preflight-full and control commands in their recorded JSON receipts. Require
   each intended named test's PASS; regexes with no match are not evidence.
6. Freeze all reconstructed source/config/protocol hashes BEFORE running the
   original 24-cell run.mjs plan, then service-run.mjs. Preserve failed commands
   and stop on errors; do not skip initialization, compaction or future phases.
   Fresh measurements can fail the historical gates. Do not overwrite these runs.

Geometry is a finite synthetic index test, not task accuracy. Public quality is
a cold-law DESIGN study, not untouched outcome-label confirmation. Serving/load,
open-loop offered rate, async compaction, byte/RSS bounds, crash/cross-epoch and
learned-law correctness remain separate required work. All seven goals OPEN.
