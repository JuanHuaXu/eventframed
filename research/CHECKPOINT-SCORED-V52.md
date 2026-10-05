# All-mixture score alignment checkpoint

Reproduce from `<LOCAL_ROOT>`, never production OpenClaw.

1. `node research/scored-v52-gen.mjs` verifies or generates only new owned
   sources; optional preflight-repair mode is forbidden after experiment freeze.
2. `node research/scored-v52-run-gen.mjs` generates the harness only before its
   exclusive root exists. Both generated scripts must pass `node --check`.
3. `node research/scored-v52-diagnostic.mjs` seals the complete research source
   closure, records terminal commands/logs, generates the fresh full diagnostic,
   runs race/static/leakage/corruption/allocation checks, collects every candidate,
   and independently audits all-head and same-world outer-only forecasts.
   Its166-file freeze has THREE compiler-dependency omissions; do not call it
   complete. Future experiments must derive closure before freezing/running.
4. Read ACTUAL weekly usage from the tool; run
   `EVENTFRAME_WEEKLY_USAGE=<observed> node research/scored-v52-readback.mjs`.
   It separately recomputes all metrics, exact compatibility and contrasts.
5. Update the results and research-direction record from raw evidence, then
   refresh usage and run
   `EVENTFRAME_WEEKLY_USAGE=<observed> node research/research-checkpoint-scored-v52.mjs`.

Before step5, run the independent post-run `scored-v52-closure-audit.mjs` and
`scored-v52-closure-binding.mjs`. The latter binds three omitted calibration
source hashes to a prior-day record WITHOUT amending the original freeze.
Preserve the initially unsupported Go-template and overcounting-audit failures.

The checkpoint chains to `checkpoint-2026-10-04-public-feedback/manifest.json`,
verifies all fourteen preexisting tracked dirty-file hashes, and copies small
assets while hash-linking files larger than8MiB. Large assets remain local;
this checkpoint is not self-contained without those linked files.

Compatibility readback is supplementary, not a scientific gate. Its serialized
readback process overlapped the beginning of Brier/arithmetic collection; retain
that timing confound alongside fixed-order/shared-host variation. Do not infer
a causal speedup from the observed loop differences. No measured failed gate
is silently rerun or replaced. All remaining required managed jobs must reach
terminal state before reporting this checkpoint as complete.

All seven goals remain OPEN/ACTIVE until their complete criteria are proved.
No production, private corpora, whitepaper, commits, pushes or installations.
