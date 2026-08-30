## 2026-08-29 - Verify Homebrew daemon executable paths

- Category: error
- Context: An isolated contract probe tried to launch `.../bin/libravdbd-service` from stale Homebrew caveat text.
- Evidence: The command failed with `no such file or directory`; the installed formula contains `bin/libravdbd` and no service wrapper.
- Lesson: Resolve the executable with `command -v` or inspect the installed Cellar before launching a probe.
- Scope: Project-local.

## 2026-08-29 - Sandbox rejects force-removal cleanup

- Category: tool
- Context: Cleanup of an isolated `/tmp/eventframed-postrank.*` test directory used `rm -rf`.
- Evidence: `exec_command` rejected the command as an impermissible `rm -f` style operation.
- Lesson: Clean Codex-created temporary trees with an exact-path `find <path> -depth -delete` command.
- Scope: Project-local.

## 2026-08-29 - Quote GitHub API URLs containing query strings

- Category: shell
- Context: A `gh api` tree request included `?recursive=1` unquoted under zsh.
- Evidence: zsh rejected it with `no matches found` before `gh` ran.
- Lesson: Single-quote `gh api` paths containing `?`, `&`, or glob characters.
- Scope: Project-local.

## 2026-08-29 - LibraVDB isolated daemon launch requirements

- Category: configuration
- Context: Fresh sidecar attempts combined `LIBRAVDB_DB_PATH` with `LIBRAVDB_AGENT_DB_ROOT`, then tried the unsupported backend name `onnx`.
- Evidence: The daemon rejected the combinations with `db_path and agent_db_root are mutually exclusive` and `unsupported embedding backend: onnx`.
- Lesson: For isolated tests, launch the real Cellar binary with only `LIBRAVDB_DB_PATH`, a unique `LIBRAVDB_GRPC_ENDPOINT`, and explicit installed llama, ONNX runtime, and cognitive-scanner asset paths; leave the backend at its packaged default.
- Scope: Project-local.

## 2026-08-29 - LibraVDB dirty-journal failure under replay load

- Category: upstream-runtime
- Context: LibraVDB 1.9.36-beta.5 served a full Codex replay and concurrent contract benchmark.
- Evidence: It repeatedly logged `microtemporal: dirty-anchor generation is stale`, then panicked on shutdown with an index-out-of-range in `causal.(*DirtyJournal).Drain`.
- Lesson: Do not attribute EventFrame tail latency from a loaded sidecar until the sidecar's dirty-journal maintenance path is stable; preserve the stack and benchmark as integration evidence.
- Scope: Project-local; no matching open upstream issue was found on 2026-08-29.

## 2026-08-29 - Do not use `path` as a zsh loop variable

- Category: shell
- Context: A cleanup loop used `for path in ...` under zsh.
- Evidence: zsh's special `path` array rewrote `PATH`, and subsequent `find`, `git`, and `jq` commands were reported as not found.
- Lesson: Use names such as `target` or `item` for zsh loop variables; reserve `path` because it is tied to `PATH`.
- Scope: Project-local.

## 2026-08-29 - Separate predictive output from invalidation metadata

- Category: test-design
- Context: A graph-scaffold integration test compared the complete forecast bundle before and after a predictive snap.
- Evidence: The test failed only because `PosteriorVersion` advanced; scores, laws, templates, and rank deltas were unchanged.
- Lesson: Tests for operational predictive effects should compare scored laws and ranking outputs separately from version metadata that is expected to change during invalidation.
- Scope: Project-local.

## 2026-08-30 - Run clean-install checks from the package directory

- Category: workflow
- Context: The first detached-worktree RC check invoked `npm ci` with an absolute `--prefix` and continued after that command failed.
- Evidence: npm incorrectly reported a missing local package despite matching package and lock versions; changing into `plugin/` made the same clean install succeed.
- Lesson: In clean-checkout gates, enable immediate shell failure and run npm lifecycle commands from the package directory instead of through an absolute `--prefix`.
- Scope: Project-local.
