# Reproduction Layout

The stored absolute temporary paths identify the original run; they are not
production dependencies. Reconstruct a fresh layout and write NEW manifests and
results to a separate artifact directory, never overwrite historical evidence.

1. Make two isolated eventframed checkouts at the exact `revision` in
   repaired-manifest.json. Include the checked-in public fact corpus and verify
   its SHA256. The existing runtime/learner helpers are in that revision.
2. Copy pinned LibraVDB v1.6.13 to a temporary `libravdb-fork` sibling. Verify
   the six original hashes, make ONLY those copied source files writable, and
   apply dependency-v2.patch. Do not patch the installed module cache. For direct
   library tests, copy the pinned lexer v0.1.12 to a temporary `lexer` sibling to
   satisfy the dependency's pre-existing `replace ... => ../lexer` declaration.
3. In each eventframed checkout use a local module replacement for the temporary
   fork. Candidate alone gets candidate.patch. The public load fixture comes
   byte-identically from the earlier public-capture-load-v1 archive and is
   formatted with gofmt. Candidate also gets the gofmt-formatted topology fixture.
4. Install gofmt-formatted admission/discovery/quantization/codec fixture
   templates at the four paths in `libraryTests`. The manifest carries compiled
   hashes, and verify-v2.mjs checks template-to-compiled equality. Preserve the
   original-before artifacts: they are historical evidence, not fresh replays.
5. Regenerate the source/path manifest for this NEW layout. Run the exact
   11-command validation plan recorded in repaired-validation-manifest.json.
   Require named PASS entries; opt-in SKIPs are not executed tests. Run the
   eight load commands in repaired-run-manifest.json only after preflight passes,
   with the declared frontier and ten-worker environment in run-v2.mjs.
6. verify-v2.mjs independently recomputes transcript hashes, counts, intervals,
   quantiles and gates. Regenerate checksum files only after the new artifacts
   and protocols are frozen. New timings may fail; do not relabel failure as a
   successful reproduction or change thresholds after looking at the output.

The optional supplemental race/partial-shard checks have their own protocol;
they do not alter the original eight-cell timing criteria. This reconstructs
the finite pilot, not a complete seven-goal scientific replication. Go version
and environment used here are in the original manifest; module dependencies
must remain pinned and commands use -mod=readonly.

Raw native candidate-live-sample.txt remains local. It was collected while
waiting for the original timeout, contains local process details and is not
needed for the authoritative Go timeout/regression evidence. Its omission from
the scoped checkpoint is deliberate, not an omitted failing experiment.

For V3, additionally apply training-v3.patch (or the combined dependency-v3.patch
instead of V2's patch), add the formatted training and missing-child fixtures,
and use v3-manifest.json plus the separately named validation-v3.mjs/run-v3.mjs
plan. The local fork name in its module replacement is libravdb-fork-v3.
The twelfth validation command is the complete candidate/200 race fixture.
Its instrumented timing-only nonzero exit is explicitly not a passing ordinary
performance command. verify-v3.mjs and finalize-v3.mjs retain the original V2
race failure and the small pre-patch unit's non-reproduction. New output belongs
in a new artifact directory; these recorded results must remain immutable.
