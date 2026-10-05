# Public Executable Retrieval: Finite Component Pass

2026-10-03.432valid V2 cases,54exact imported EventFrames,108questions,
18API clusters. FIT36questions/six clusters; DESIGN36/six; untouched
CONFIRMATION36/six. Original216FIT cases replayed only to repair measurement,
not counted as additional evidence. Shared predictions/fields match exactly.
No private data, OpenClaw/LLM call, production store or generated answer test.

## Frozen Fit And Results

FIT actual packed top1:lambda0=11/36,.25=11,.5=11,.75=15,1=25.
Selected lambda1 BEFORE held-out outputs: this is **pure lexical coverage**,
NOT evidence that hybrid fusion or confidence modulation improves retrieval.
Magnitude on the one lexical channel and incumbent rank utility were tested;
raw semantic scores are absent from the existing hook. Lower weights lose on
FIT. We do not retune after held-out results or erase the earlier Git RRF loss.

| Split | Baseline top1 | Selected top1 | Gains/losses | Baseline survival | Selected survival |
| --- | ---: | ---: | ---: | ---: | ---: |
| Design | 11/36 | 27/36 | 16/0 | 21/36 | 36/36 |
| Confirmation | 12/36 | 27/36 | 16/1 | 22/36 | 36/36 |

No literal top1 losses; one confirmation paraphrase loses. No target-packet
survival losses. All72held-out targets nominated before truncation. Both
splits PASS the frozen finite adoption screen: >=2net top1 gains, no literal
or survival loss, positive cluster-paired one-sided95% t bound.

Design cluster gain vector:0,.5,1/3,.5,5/6,.5; mean .444444,
one-sided lower .220550, two-sided95% [.158824,.730065], exact discordant
cluster sign p=.03125. Confirmation:1/3,.5,1/3,1/3,.5,.5; mean .416667,
lower .341570, two-sided95% [.320867,.512467], sign p=.015625. Six clusters
per split, correlated tasks/wordings, assumption-dependent intervals; not
108independent samples or a simultaneous audit certificate.

## Technical Audit And Performance

Source/model freezes, exact54 imported frames, all54journal laws, stable
utility formulas/order,50retained prepack entries, packet/journal identity,
baseline/incumbent packet parity and no timed embedding miss verified.
Fourteen non-identity corruption controls reject. Package race/vet pass.
Ordinary occupancy/correlation/token packing remains ON. Scores use only
six5W1H fields, never full-text content/source metadata/target IDs/oracle.
No properly scored forecast law changes across arms: gains are SEARCH ORDER.
Scorer time200candidates340.6-343.6ns/op,1792B,one allocation; excludes
feature extraction and service sorting. Held-out warm service timings:

| Arm | Median Recall ms | p99 Recall ms | p99 callback ms |
| --- | ---: | ---: | ---: |
| Baseline | 1.864667 | 3.070291 | 0 |
| Incumbent control | 2.767458 | 3.727583 | .076541 |
| Selected lexical | 2.722584 | 4.262750 | .043500 |

Recall includes existing extraction/rank/packing/journal/memory-store work;
no loaded durable serving test. Cold acquisition aggregate1.887236377s FIT
and1.756443918s held-out is excluded from warm Recall and NOT per-query
latency. Phase collectors4.787292s and3.681328s include compilation/import/
acquisition/serialization, not a full production e2e pipeline. Memo role
separation/512cap; installed Nomic digest unchanged. No sub100ms full-load
Goal6 proof or automatic text-to-frame extraction success follows.

## Honest Repair Trail And Scope

First prospective auditor wrongly assumed50callback entries; fails54!=50.
First supplement wrongly assumes54tap entries; fails50!=54. Both freeze
manifests/scripts and original failure/log remain. Correct service boundary:
Recall50*default overfetch3 ->54available nominees ->score all54 ->store
all54journal decisions ->trim50 ->tap50 ->pack10. Missing original full
journal measurements cannot be retrospectively claimed verified.

V2 adds full journal diagnostics ONLY, freezes before replaying FIT, proves
original stable prediction equivalence, audits it plus14controls, selects
the unchanged fit grid, then freezes model before untouched held-out calls.
This is an explicit post-hoc measurement repair, not an original-protocol
technical PASS. No scorer, method, fit objective or quality gate changed.

The task corpus stores public known answers; we measure retrieving them, NOT
inferring unseen semantics or improving an agent's generated response. Code
literal overlap, punctuation loss in ASCII extraction, related API families,
embedding pretraining and author-designed questions limit transfer. Direct
Observe imports bypass text conversion; compare future retrieval from raw
text as a separate extraction test. All six confirmation clusters improve
on average, but remaining9/36top1 misses need diagnosis, not hindsight rules.

Next: repeat the frozen lexical component on a NEW outcome-labeled non-code
task family with semantically confusable distractors, and compare to a truly
score-aware bounded fusion only after its contract exposes those scores.
Do not adopt a global lexical-only override from this pilot. Goal5 has a
finite component win; all seven whole research goals remain OPEN.
Authoritative `magnitude-v2-{model,results,completed,controls}.json` and raw
FIT/held-out `.jsonl`; all previous failed artifacts are preserved.
