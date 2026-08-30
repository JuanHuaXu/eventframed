# Contained OpenClaw E2E Clean-State Confirmation Results

## Outcome

The rerun removed the prior pilot's answer-reuse path. The frozen LibraVDB
control, EventFrame pass-through, and active EventFrame arms each answered all
eight positive memory queries and the absent-fact control correctly using the
production `openai/gpt-5.6-luna` model.

This confirms the contained integration under stronger isolation. It still
does not show an active-ranking quality gain: active EventFrame changed scores
but produced exactly the same packed order as pass-through on all nine queries.

## Accuracy And Ranking

| Arm | Positive answers | Absent fact | Positive target ranks |
|---|---:|---:|---|
| Frozen LibraVDB | 8/8 | 1/1 | 1, 1, 1, 8, 1, 1, 1, 1 |
| EventFrame pass-through | 8/8 | 1/1 | 1, 1, 1, 3, 1, 1, 1, 1 |
| EventFrame active | 8/8 | 1/1 | 1, 1, 1, 3, 1, 1, 1, 1 |

The current `Andean` record was eighth in the control context and third in both
EventFrame packets; the old `Baltic` record remained ahead of it in all three
arms. The model resolved the explicit supersession correctly. The quoted
instruction in the Quartz record was also treated as data: every arm returned
`QUARTZ`, not `MALACHITE`.

Active EventFrame assigned a nonzero rank delta to all 90 packed candidates.
Deltas ranged from 0.01166 to 0.01476 with mean 0.01261, but no candidate order
changed relative to pass-through. No candidate used a committed Bayesian
posterior in this short cold-start run; the Bayesian records remained shadow
evidence. The correct conclusion is that active scoring executed, not that it
improved these answers.

## Isolation Evidence

| Arm | Query prompts compiled | Prior messages | Last compile before first completion |
|---|---:|---:|---:|
| Frozen LibraVDB | 9/9 | 0 in every query | 2.131 s |
| EventFrame pass-through | 9/9 | 0 in every query | 2.583 s |
| EventFrame active | 9/9 | 0 in every query | 2.241 s |

- Seed and query phases used distinct OpenClaw state directories, workspaces,
  and session keys in every arm.
- The three arms used distinct new tenants; the EventFrame arms used distinct
  daemon and rank-delta databases.
- No query workspace contained `MEMORY.md`, a `memory/` tree, seed text, or
  answer text. Internal `session-memory` was disabled.
- The positive query prompts did not contain their expected answers. The sole
  declared exception was the `UNKNOWN` text in the absent-fact control.
- All eight positive targets were visible at the provider boundary in every
  arm, inside retrieved context.
- EventFrame query configurations had `capture=false`. Each EventFrame query
  trace contains exactly nine recalls and zero observations.
- The frozen control lacks a capture switch, so all queries were dispatched
  concurrently. All nine prompts compiled before any model completed, proving
  that no query answer could be captured in time to influence another query.

The first control query launch was rejected before model execution because its
fresh state treated an explicitly loaded Codex harness as untrusted. It created
no trajectory and no answer. The managed trusted plugin was then selected and
the untouched query batch was run; no failed-query text appeared in retrieved
memory. Those rejected calls are not part of the scored result.

## Latency And Resources

| Arm | Recall median | Recall p95/p99/max | Mean |
|---|---:|---:|---:|
| EventFrame pass-through | 36.29 ms | 58.90 ms | 41.33 ms |
| EventFrame active | 38.59 ms | 62.00 ms | 45.04 ms |

With nine observations, p95 and p99 both select the maximum. Recall timing is
adapter-observed and includes the remote mTLS LibraVDB contracts but excludes
model generation. Both arms remained below the predeclared 100 ms gate under
the concurrent batch. Active added 2.70 ms at the median and 3.70 ms on the
mean relative to pass-through in this sample.

One end-of-run sample reported 44,336 KiB RSS for pass-through and 64,976 KiB
for active, both at 0.2% CPU. Each daemon event database occupied 652 KiB; rank
delta storage occupied 12 KiB for pass-through and 96 KiB for active. Model
generation under nine-way concurrency took seconds and is not attributed to
EventFrame.

## Contract And Sink Checks

Every EventFrame query packet used
`libravdb.ipc.v1.LibravDB/SearchTextCollections` for nomination and
`libravdb.ipc.v1.LibravDB/RankCandidates` for ranking. OpenClaw trajectories
showed the retrieved facts in the provider-visible prompt before each answer.
No channel delivery or agency action was enabled.

## Claim Status

| Claim | Result |
|---|---|
| Clean cross-session OpenClaw memory retrieval | Validated in this confirmation |
| No dependence on persisted markdown/session context | Validated for this harness |
| Provider-visible EventFrame context injection | Validated |
| Contract-native LibraVDB nomination and ranking | Validated |
| EventFrame recall below 100 ms | Validated for this nine-query workload |
| Query capture can be disabled for EventFrame | Validated |
| Active ranking improves packed order or answer quality | Not observed; inconclusive |
| Population or production-workload superiority | Untested |

## Limits

This remains a small synthetic test with easy top-10 targets, one model, and one
run per query. It does not test hard rank-11-to-50 rescues, corpus-scale storage,
targets outside the 50-candidate frontier, long-run posterior commitment,
changepoints, autonomous agency, or production latency distributions. The raw
states, prompts, and trajectories remain in the ignored `.eventframed/e2e`
directory and are not publication artifacts.
