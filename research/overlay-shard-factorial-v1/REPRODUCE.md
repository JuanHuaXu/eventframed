# Reproduce In Isolation

1. Restore the V6 dependency and service sources per its REPRODUCE.md and source
   archives, with exact hashes before any edits. Do not patch installed caches.
2. Read PROTOCOL.md. The manifest names four fresh service copies and two fresh
   dependency copies plus lexer. prepare.mjs copies the V6 unsharded service into
   every arm, chooses plain/overlay module replacement, and adds WithSharding(true)
   plus the original topology fixture only in shard/both. All sources are pinned.
3. In a NEW output directory run the preparation and immutable run plan. Never
   overwrite this DESIGN study. run.mjs stops on functional preflight failure,
   retains ordinary finite timing failures and executes all fresh opposite-order
   comparisons. GOMAXPROCS=10; no overlapping CPU experiments.
4. After all commands are terminal, run quality-evaluate.mjs then the independently
   derived quality-independent.mjs; load-evaluate.mjs then load-independent.mjs.
   All quality metrics use raw numerical rows, no private/source conversation
   text. Independent quality arithmetic is not a second random-vector oracle.
5. Run node --test evaluate.test.mjs negative.test.mjs, followed by verify.mjs.
   The verifier also reads the original prospective source paths. For a relocated
   artifact use its exported verify(root,{sources:false}) and explicitly report
   no current-source readback. It must not silently skip command checks.
6. Review failures separately from consistency: all seven whole goals remain
   open unless their original full criteria are actually met. No deployment or
   automatic publication follows any finite screen.

The oversized public numerical quality traces may be losslessly gzip-archived.
Keep original bytes/hashes in receipts and verify decompression. These are cold
synthetic mechanics over public DESIGN text, not unseen outcome labels, learned
posterior validation, peak-RAM proof or an open-loop production performance study.
