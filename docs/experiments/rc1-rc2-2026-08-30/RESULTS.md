# RC1 and RC2 release-candidate run

Date: 2026-08-30

## Decision

RC1 passed the local binary, restart, bounded-load, and adapter fail-open gates.
RC2 did not pass its evidence gate and is not promoted. The ChatGPT corpus is
useful as a design pilot, but its untouched confirmation block contains only
one eligible case and the apparent design gain is reproduced by shuffled
feedback.

No RC1 or RC2 tag was created by this run.

## RC1

The exact `v0.1.0-rc.0` source reproduced the previously recorded binary:

```text
SHA-256 5c3bf24eef06bc4ed368acfa0e9749b2eeaafad6c44e1b74bfe2586c1ef72aa8
```

`make check build VERSION=0.1.0-rc.0` passed the Go race suite, `go vet`,
TypeScript checking, all adapter tests, and both builds. A fresh private daemon
state was then loaded with 50 records and exercised by 500 recalls at
concurrency 10.

| RC1 check | Result |
|---|---|
| Exact RC0 binary reproduction | Passed |
| Go race, vet, TypeScript, and unit gates | Passed |
| 500-recall bounded soak | Passed |
| Soak latency p50 / p95 / p99 / max | 2.27 / 3.73 / 4.65 / 7.51 ms |
| Process restart and durable recall | Passed; 50 recalled, 10 packed |
| Adapter hook with absent daemon | Passed; warning recorded and prompt build continued |
| Bounded agency | Disabled |

The adapter outage check runs the actual `before_prompt_build` callback against
an absent Unix socket. A duplicate sink-visible OpenClaw outage run was also
attempted, but a contained OpenClaw startup migration tried to snapshot its
approximately 2 GB managed plugin tree with only 1.1 GB free. It stopped before
model execution. The successful clean contained OpenClaw path remains recorded
in `openclaw-e2e-clean-confirmation-2026-08-29`; this run does not claim a new
sink-visible outage confirmation.

## RC2 corpus

The private extraction contained 38 ChatGPT conversations, 86 connector pages,
and 680 raw turns. The loader accepts only completed user/assistant turns with
valid availability ordering, deduplicates pagination overlap, hashes all event
identities, and uses the frozen conversation-level split in the private
manifest. One completed record with an impossible completion-before-start time
was excluded rather than repaired.

After the ten-prior-turn eligibility rule, 18 conversations and 548 turns were
usable: 15 design conversations and 3 confirmation conversations. The objective
label rule yielded 29 design cases across 8 source-conversation clusters and
only 1 confirmation case in 1 cluster.

## RC2 results

| Metric | Baseline | EventFrame | Shuffled feedback |
|---|---:|---:|---:|
| Design Brier | 0.46018 | 0.42619 | 0.42589 |
| Design Recall@10 | 75.86% | 96.55% | 96.55% |
| Design packed recall | 65.52% | 79.31% | 79.31% |
| Design MRR | 0.4870 | 0.6387 | 0.6584 |
| Confirmation Brier | 0.48619 | 0.48619 | 0.48619 |
| Confirmation Recall@10 | 100% | 100% | 100% |
| Confirmation packed recall | 100% | 100% | 100% |

The design EventFrame arm improved Recall@10 by 20.69 percentage points and
packed recall by 13.79 points over the no-update baseline. Those intervals were
positive across the eight design clusters. However, shuffled relevance feedback
produced the same recall gains and a slightly larger Brier gain. The experiment
therefore does not attribute the improvement to correct Bayesian evidence; it
is consistent with generic update-path behavior. EventFrame and `update_all`
were identical, as expected under the frontier-update-all release policy.

The single confirmation case was already rank 1 and packed in every arm. It
cannot support clustered inference or a promotion decision.

## Privacy

Raw ChatGPT pages, source identifiers, titles, prompts, answers, and local
derived datasets remain under ignored mode-0700 `.eventframed` storage. The
tracked record contains aggregate counts and metrics only. A scan of the
derived result directory found no home paths, source IDs, raw message types,
email addresses, or transcript text.

## Next gate

RC2 requires a prospectively accumulated confirmation block with enough
objective hard-retrieval cases and at least two, preferably many,
conversation-level clusters. Promotion also requires EventFrame to beat both
the no-update baseline and the shuffled-feedback control; beating only the
former is insufficient.
