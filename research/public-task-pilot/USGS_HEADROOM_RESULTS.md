# USGS near-duplicate retrieval: design and held-back confirmation

The public source is a frozen response from the official
[USGS earthquake catalog API](https://earthquake.usgs.gov/fdsnws/event/1/).
The 24-record corpus contains 12 July 4 and 12 July 6 Ridgecrest-area events.
Each event has two paired magnitude questions. Two absent questions were in
the design split only. The source snapshot, corpus, questions and oracle were
frozen before the design service run. The [confirmation gate](USGS_CONFIRMATION_PROTOCOL.md)
was written after design and before the held-back July 6 run. No algorithm
parameters changed between splits.

| Split and arm | Top1, literal | Top1, paraphrase | Survival@10, literal | Survival@10, paraphrase |
| --- | ---: | ---: | ---: | ---: |
| Design focus | 3/12 | 1/12 | 8/12 | 8/12 |
| Design priority | 12/12 | 12/12 | 12/12 | 12/12 |
| Confirmation focus | 1/12 | 1/12 | 4/12 | 4/12 |
| Confirmation priority | 10/12 | 10/12 | 12/12 | 12/12 |

The predeclared design-headroom rule and the subsequently frozen finite
confirmation gate both **PASSED**. On confirmation there were 18 paired-query
top1 gains and zero losses, representing nine distinct improved events among
12 event clusters. All 24 confirmation target records were in the ranker
frontier. Priority preserved every baseline success and rescued all 16
baseline packing misses, representing eight distinct events. The two priority
top1 misses were event IDs
`ci37257780` and `ci38457519`, each missed in both wordings; the same
nearby timestamped record (`ci37219484`) was first in both cases. The score
gaps were small (about 0.009 and 0.013 in the inspected literal traces).

Every confirmation ranker frontier contained the same 24 unique records,
the passthrough numeric ranker made no change, full-frontier forecast laws
and original numeric scores were identical between arms, packet explanations
matched journals, and the independent scorer found no invariant errors.
The isolated nearest-rank Recall p95 was 22.03 ms for focus and 23.21 ms for
priority, both below the frozen 100 ms feasibility ceiling. These are 24
sequential in-memory calls per arm, not concurrent service p99 or full-agent
latency. Both design absent controls packed unsupported records; no
abstention or answer-generation claim follows.

The gain is interpretable: these questions contain UTC times, and the
existing research lexical overlay strongly favors the matching timestamp.
This demonstrates a useful structured retrieval behavior on a dense,
near-duplicate public corpus. It does **not** validate general EventFrame
reasoning or continuous learning. The overlay failed the earlier RFC
support-survival gate, and the NARA transfer had a top1 loss. The USGS
design and confirmation clusters come from one earthquake sequence with
fixed query templates, and no LLM or agent answered the questions. This is
not the prospective, multi-task outcome evidence required to complete goal 5.
The unchanged overlay remains research-only; production stays untouched.

Reproducibility: source snapshot SHA-256
`bbad31a11967c08d01b454a06b34e4714ad80fa4672fb87481746e235b823583`;
design raw SHA-256
`f91cfca99a431f9164dce4855161c3561d5722c38bff14910dd8ece39569fb48`;
confirmation raw SHA-256
`81e5647ada3b2309eee6ea10289331a4abd038649547e9b3a583936ff5f65571`;
oracle SHA-256
`472b0bef316c94a7c0dcdc2dd4928b3c9d2e14855696b9fb1145d1b7aaf94602`.
All raw traces, metadata, oracle and scored rows are under `usgs-headroom-v1/`.
The confirmation runner compiled under the race toolchain and passed vet.

Next research step: evaluate semantic paraphrases without shared timestamp
strings and, separately, actual agent tasks with observed answer outcomes.
Freeze new populations and controls before inspecting output; do not tune
the present overlay on this now-consumed sequence. Goal 5 remains open.
