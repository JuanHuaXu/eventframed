# Prospective Git-manual rank fusion: failed

2026-10-03. All500 isolated cases completed on100 new public questions, five
arms,48 facts and16 command-family clusters. [Frozen protocol](PROTOCOL.md).
Both design and confirmation were frozen before collection; no fitting or
tuning occurred between them. Confirmation is now consumed. Retrieval labels
are documentary support targets for constructed agent questions, not measured
generated-answer or command-execution outcomes. Whole Goal5 remains OPEN.

| Split / wording | Baseline top1 / survival | Lexical top1 / survival | RRF top1 / survival | Protected top1 / survival |
| --- | ---: | ---: | ---: | ---: |
| Design literal |20/24 /24/24|22/24 /24/24|20/24 /24/24|20/24 /24/24|
| Design paraphrase |10/24 /24/24|9/24 /23/24|7/24 /24/24|7/24 /24/24|
| Confirmation literal |20/24 /24/24|24/24 /24/24|20/24 /24/24|20/24 /24/24|
| Confirmation paraphrase |14/24 /23/24|7/24 /24/24|12/24 /24/24|12/24 /23/24|

Primary protected screen **FAILS both splits**: zero top1 gains, three losses
in design and two in confirmation. All losses are paraphrases. Mean command-
cluster top1 differences are -.0625 and -.041667; one-sided95% lower bounds
-.120279 and -.093346. Two-sided95% paired t intervals are
[-.134614,.009614] and[-.106167,.022834]. Eight command clusters per split,
not48 independent questions; small purposive-family intervals are descriptive,
not population-wide guarantees. All discordant clusters favor baseline (three
design,two confirmation). The simpler lexical control also fails both screens.

Exploratory unprotected RRF also **FAILS both splits**. It shares the same top1
losses, but retrieves one previously unpacked correct answer into position9:
`question-068`, the fetch-pruning paraphrase. The answer was incumbent frontier
rank20, lexical rank2. Protected prefix blocks that promotion and cannot repair
this miss. There are NO nomination misses: all96 target answers enter the
frontier in every arm. This finite result localizes the gap to ranking/packing,
not corpus discovery, and does not measure million-record nomination recall.

Four deliberately unsupported controls return ten unrelated records per arm.
They are not correct abstentions. They are excluded from positive accuracy.
The incumbent passthrough control is exactly baseline in order and packet
membership. Protected pre-packing top-ten sets and actual packet sets agree
with baseline on all100 queries. This empirical packet equality is not a
general guarantee under correlation occupancy or token-budget constraints.

## Audit And Localization

Independent JS arithmetic reconstructs all hook scores from recorded six-field
5W1H values, not raw corpus text or metadata. All500 journals match the packets;
full-frontier and packed corrected Bernoulli laws are bit-identical across arms
for the same event/query. Candidate IDs, permutation application, bounds,
role-separated warm embeddings and source freezes pass. Nine corrupted traces
(score,law,duplicate,missing case,cold Recall,freeze,journal,lexical,order) are
rejected. Go tests independently compute ranks by O(n^2) counting across every
frontier size0..200; cancellation, invalid values, input ownership, monotone
score transforms, prefix-set and top1 counterexamples pass. Race/vet pass for
the isolated core and runner. Existing service-hook and embed/frame race
regressions were additionally rerun; see the checkpoint for terminal status.

170 local runtime/input/source hashes and eight separate auditor/label hashes
were frozen before outputs. The runner never opens the facts construction file
or oracle. It only reads corpus, queries and its label-free source manifest.
The auditor alone reads the oracle after raw output is complete. Questions
exclude their target option/subcommand token. Researcher authorship and possible
embedding pretraining on Git manuals remain explicit limitations, not eliminated
leakage mechanisms. All factual distinctions were checked against the official
manuals linked in the protocol; no command represented by a fact was executed.

[Post-hoc diagnostic](diagnostic.json) changes no rule or adoption criterion.
All five RRF top1 losses have correct incumbent rank1 but lexical ranks3/4;
weak token overlap promotes a different answer. Nine lexical repairs include
eight incumbent-rank2/lexical-rank1 cases: symmetric equal-weight RRF ties these
against incumbent-rank1/lexical-rank2, discarding lexical score margins, and the
stable tie retains incumbent. The remaining repair is incumbent rank3. Therefore
increasing a rank-only confidence or repeating this screen is not an established
rescue. A future confidence-aware or learned combination needs disjoint fitting
and a NEW untouched task domain, with paraphrase/packet protections retained.

## Performance Boundary

| Arm |Warm Recall median ms|p95 ms|p99 ms|max ms|
|---|---:|---:|---:|---:|
|baseline|2.008166|2.374917|2.746833|5.486917|
|incumbent hook|2.853500|3.302792|3.631958|4.526875|
|lexical|2.811250|3.451625|3.533209|3.556792|
|RRF|2.810916|3.313208|3.508666|3.572334|
|protected|2.816000|3.386541|3.512583|3.973375|

100 serial calls per arm, rotating arm order each query. These are fresh
in-memory service/store calls with installed local Nomic768 embeddings, not
libravdb, OpenClaw, loaded durable serving or generated-agent answers. Timed
Recall excludes all embedding misses: exact role-separated512-cap memo acquired
148 entries (48 frame documents,100 queries),4432.222118ms total,24852 hits.
Baseline capture includes the initial document acquisition; other arm captures
are warm and cannot be treated as fair cold-ingestion comparisons.

Three fusion component benchmark runs:10 candidates127.3-127.8ns/216B;
50 candidates1361-1364ns/888B;200 candidates9243-9255ns/3640B; four allocations
each. Callback totals across100 requests are .841703ms RRF and .768926ms
protected; those exclude the service's feature extraction. All hook arms cost
roughly .8ms more than baseline median, including the passthrough hook. This
suggests extraction/experimental instrumentation dominates over the sort in
this finite workload; no concurrent profile or causal speedup is claimed.
Goal6's durable loaded latency/freshness requirements remain untested here.

## Artifacts And Reproduction

Raw SHA256 `979779295adcb4e2dd6d5235c3f02b0131fd764ee39186f81c7ba5be1da8f41e`.
Scored SHA256 `848a1536d41145d806d16ad3b975aa48e67c34e409f5eaad64b4217b0cecc6ac`.
Diagnostic SHA256 `2c0b8e956a316f991becd0f0e693468bb746f1ffa3f85719094fd98da03622e0`.
Tests SHA256 `44562b59d3e064ca4653e3f94c4d5f9d148e23035cfb3438ea3ffe1ce8e2baf4`.

All collector/prepare/freeze outputs exclusive-create and already exist. Do
not overwrite/relabel them or count verification replay as fresh confirmation.
Running `score.mjs raw.jsonl NEW.json` from repository root re-audits, including
negative controls, without any embedding/model execution. Use a NEW output.
`node score.mjs --self-test` also tests independent arithmetic without data.
Core/runner test commands are recorded in `tests.json`; sources in the two
freeze manifests. No existing runtime, private dataset, production service,
whitepaper, installation, credential, commit or push was changed.
