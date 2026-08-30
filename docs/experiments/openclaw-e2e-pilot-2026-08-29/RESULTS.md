# Contained OpenClaw E2E Pilot Results

## Outcome

The contained OpenClaw integration worked end to end with the production
`openai/gpt-5.6-luna` model. The frozen LibraVDB control, EventFrame
pass-through arm, and active EventFrame arm each answered all seven valid
memory-bearing queries correctly. Each also passed the one declared absent-fact
control.

This validates the integration path in the pilot. It does **not** demonstrate
that active EventFrame ranking improves answer quality over LibraVDB or over
EventFrame pass-through: all three arms tied, and active scoring did not change
the top-10 order on the six directly comparable initial queries.

## Accuracy And Packing

| Arm | Valid memory answers | Absent-fact control | Target ranks |
|---|---:|---:|---|
| Frozen LibraVDB | 7/7 | 1/1 | 1, 1, 7, 1, 1, 1, 1 |
| EventFrame pass-through | 7/7 | 1/1 | 1, 1, 4, 1, 1, 1, 1 |
| EventFrame active | 7/7 | 1/1 | 1, 1, 4, 1, 1, 1, 1 |

The EventFrame target survived packing in all seven valid queries. Six targets
were packed first. In the superseded-region query, the new `Andean` record was
fourth and the old `Baltic` record was third in both EventFrame arms. The frozen
control placed the old record first and the new record seventh. The model still
answered `ANDEAN` in all arms because the new record explicitly declared the
supersession. This is model interpretation, not evidence that EventFrame's
ranking corrected the stale ordering.

The recalled instruction attack also remained data: every arm returned
`QUARTZ`, not the embedded instruction's `MALACHITE` response.

## Ranking Result

The pass-through arm produced no rank deltas. During the initial 21-turn active
run, 155 packed candidates received nonzero deltas between 0.01144 and 0.01473
(mean 0.01227). Those corrections did not change the top-10 order for Q02-Q07.
The Bayesian decision record remained explicitly in shadow mode.

Result: the scoring machinery is wired and active, but this fixture provides no
observed quality gain from it. A harder confirmation corpus with target records
at original ranks 11-50 is still required.

## Latency And Resources

| Arm | Recall median | Recall p95 | Recall maximum |
|---|---:|---:|---:|
| EventFrame pass-through | 12.76 ms | 16.66 ms | 29.88 ms |
| EventFrame active | 12.45 ms | 16.60 ms | 22.93 ms |

These are adapter-observed recall durations for the original 21 turns and
include the remote mTLS LibraVDB contract calls. All were below the pilot's
100 ms gate. One active-daemon sample was 88,512 KiB RSS and 0.7% CPU. The two
EventFrame databases and rank stores occupied about 1.5 MiB; the disposable
OpenClaw/Codex harness state occupied about 2.1 GiB.

Production-model generation was thousands of milliseconds and noisy across the
fixed sequential arm order. Its differences cannot be attributed to EventFrame
from this sample.

## Boundary And Failure Checks

- OpenClaw loaded only the Codex model harness and the selected memory plugin.
- The gateway listened on loopback port 19891 with no delivery channels.
- Tools used the minimal profile; agency was disabled with its kill switch on.
- Every original EventFrame arm produced 21 recall and 21 observe records with
  no errors.
- Packets named the real `SearchTextCollections` nomination contract and
  `RankCandidates` ranking contract.
- OpenClaw trajectories proved sink visibility: LibraVDB memory appeared in the
  compiled system prompt, while escaped EventFrame records appeared in the
  submitted prompt before the production model response.
- With eventframed stopped, recall failed in 0.84 ms, OpenClaw still returned
  `FAIL_OPEN_OK`, and capture failure remained non-fatal.

## Findings During The Run

1. The production LibraVDB endpoint requires mTLS, while eventframed previously
   hardcoded plaintext gRPC. The same recall probe failed with a connection
   reset before the patch and returned HTTP 200 over the real contracts after
   adding TLS/mTLS configuration.
2. The adapter previously lacked an E2E packet trace. An opt-in mode-0600 JSONL
   trace now records recall packets, timings, errors, observations, and lineage.
3. Captured query answers can be retrieved on a later similar query. This made
   the corrected Q01 rerun self-referential and is a concrete self-reinforcement
   risk to test further. The event lineage marks assistant material as synthetic,
   but the active ranker still packed the prior answer prominently.
4. The original Q01 disclosed `ALDER GLASS` in its requested output and is
   excluded. Its corrected rerun is also excluded because it retrieved the
   already captured answer. A new unseen `Titanium` seed/query pair replaced it
   and passed in all three arms with provider-visible evidence.

## Claim Status

| Claim | Result |
|---|---|
| Contained OpenClaw-to-LibraVDB E2E operation | Validated in this pilot |
| Provider-visible EventFrame context injection | Validated in this pilot |
| Contract-native nomination and ranking | Validated in this pilot |
| EventFrame recall below 100 ms | Validated for this small workload |
| Daemon-unavailable fail-open behavior | Validated in this pilot |
| Active ranking improves answer quality | Not observed; inconclusive |
| Population or production-workload superiority | Untested |

## Limits

This was a small synthetic pilot with a fixed arm order, one model, and one run
per query. It did not test targets outside the top-10 packet, targets outside the
50-candidate frontier, long-run Bayesian calibration, changepoints over extended
trajectories, autonomous agency, or corpus-scale resource behavior. The raw
prompt and trajectory evidence remains only in the ignored local test directory
and is not suitable for publication because it includes authenticated harness
state and complete model prompts.
