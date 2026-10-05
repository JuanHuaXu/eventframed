# Non-Evidence Setup Failures

The first admission regression command failed to build because the copied
dependency's existing `replace lexer => ../lexer` sibling was absent. Its raw
setup transcript is retained. Copying the pinned lexer to that temporary sibling
allowed the actual BEFORE regression to run and fail on admission, as recorded
separately. No global cache or installed source was edited.

The first repaired pilot preparation attempted to overwrite the read-only modes
inherited by a temporary six-file audit copy from the Go module cache. Node
returned EACCES on that temporary tx.go. The following validation invocation
returned ENOENT because preparation had not produced its manifest; no tests
ran. The partial audit copy is preserved. The corrected preparation creates a
fresh temporary audit directory and makes ONLY its copied files writable before
generating the diff. Installed cache permissions remain unchanged.

One guessed topology fixture filename caused an unsuccessful copy; listing the
original isolated candidate identified the real file, which was then copied
byte-identically. These command/setup failures are not library regressions,
negative scientific controls, or passing test coverage. The reusable local
workflow lesson is to check actual fixture names and copied-file modes before
staging the next bounded experiment, not to change durable global policy.
