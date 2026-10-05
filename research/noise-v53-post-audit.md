# Supplemental closure readback

The first supplemental inline Node readback exited1 after successfully checking
the compact Go JSON digest. Its file-list comparison wrongly included three
absolute generated Go cache test mains outside the repository. It omitted the
runner's final `absolute.startsWith(repo + '/')` filter.

Classification: CONFIRMED supplemental auditor-boundary mistake, NOT a missing
repository source or a mathematical/experiment failure. The original runner
already filtered these generated nonrepository files. No frozen sources,
outcomes, seeds, model, gates, old artifacts or global instructions changed.

`node research/noise-v53-closure-readback.mjs` applies the same explicit
repository boundary, verifies the compact input digest and all saved source
copies, and records excluded regenerable test-main paths without reading them.
The original checkpoint remains immutable; the final supplemental checkpoint
binds this clarification to it instead of rewriting its manifest.
