# National Archives packet-fusion transfer: failed

The [frozen protocol](NARA_FUSION_PROTOCOL.md) used 15 new National Archives
[milestone-document](https://www.archives.gov/milestone-documents/list) clusters,
two question wordings per cluster, two absent controls, and 13 fixed NASA
distractors. Thirty-two isolated queries ran in each of the focus and
priority service arms, using the unchanged semantic model and research
overlay. The offline fused packets were computed before the oracle was read.
This set is now consumed for tuning.

| Wording | Focus top1 / survival | Priority top1 / survival | Fusion top1 / survival |
| --- | ---: | ---: | ---: |
| Literal | 15/15 / 15/15 | 15/15 / 15/15 | 15/15 / 15/15 |
| Paraphrase | 15/15 / 15/15 | 14/15 / 15/15 | 14/15 / 15/15 |
| Absent | 0/2 / 0/2 | 0/2 / 0/2 | 0/2 / 0/2 |

The predeclared finite fusion gate **FAILED**. There were no top1 gains and
one top1 loss (`nara-rights-paraphrase`). Focus had the correct Bill of Rights
record first. Priority put Washington's first inaugural speech first; that
record was already in the focus packet, so the fusion rule moved it ahead of
the correct record. All 32 fused packets retained exactly the focus support
set. This confirms set-preserving fusion does not imply top1 non-harm.

The positive set was also too easy for measuring possible top1 improvement:
focus scored 30/30. The failure to gain is therefore a ceiling effect here,
not evidence that no harder task could benefit. The one observed loss is an
actual counterexample to the proposed non-harm claim. Both absent questions
still returned ten unsupported records; this is not abstention.

Audit: every arm had 28 ranker inputs, its passthrough ranker left them
unchanged, all packet IDs were valid and unique, packet explanations matched
their journals, and full-frontier numeric scores and forecast laws were
identical across focus and priority. The independent scorer reported no
invariant errors. Raw SHA-256 is
`51706ae54d8c534c6755a25bc3a00486c8f0b7bb6cd55c5146971f14ba5ef5e6`;
oracle SHA-256 is
`f6eff5abd160b7bb13d274ea9eddbfec41f2b3723f5363bfe3248ef126860c5c`.
Raw traces and the scored rows are in `nara-fusion-v1/`.

The fusion component benchmark on Apple M4 measured 239.8-242.6 ns/op,
616 B/op, four allocations across three runs. Isolated Recall median was
20.51 ms for focus and 21.31 ms for priority (32 calls each). Those are
separate service calls, not the cost of a single integrated fused request;
their sequential sum and the microbenchmark do not establish an OpenClaw
latency or throughput claim. Race tests and vet passed for the isolated
research package and runner.

Next: do not tune a confidence threshold on this or the consumed RFC set and
call it confirmation. A new prospective agent-task set must include harder
cases with baseline misses, observable answer outcomes, and an independently
frozen selective-promotion rule. Goal 5 remains open. No production or
default runtime code changed.
