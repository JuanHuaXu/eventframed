# Tadine current-contract ranker transfer

Frozen after discovering the archived overlay interface gap and before any
Tadine service recall or oracle scoring. Reuse the already frozen official
source, 24-record corpus and 72 questions from
[`USGS_TADINE_TRANSFER_PROTOCOL.md`](USGS_TADINE_TRANSFER_PROTOCOL.md).
This is a **new candidate-ranker policy**, not a rerun of the archived packet
overlay. The fixture source and answer labels are unchanged.

## Arms

`focus`: the existing passthrough candidate ranker. `task-lexical`:
at the current `CandidateRanker` boundary, run the unchanged
`researchcalendar.PlanTaskRole` and `LexicalOrder` on the full nominated
frontier. Return every candidate in that order with search score
`1-rank/(n+1)`, where rank starts at zero and n is the frontier length.
This deterministic score is an ordering device, not a calibrated probability.
No fitting, case-specific threshold, oracle access or candidate pruning.
The numeric ranker is intentionally different between arms; the scored
forecast law must not change. Both arms retain recall 50, pack 10, diversity
enabled, the pinned Nomic digest and a fresh in-memory service per case.

The 36 design and 36 confirmation questions are all run without tuning. The
confirmation stratum is reported separately, but the finite pass gate below
applies to the complete 72 cases, with cluster-level paired counts to avoid
treating three wordings as independent events.

## Frozen gate

1. At least eight paired top-1 gains overall, at least one in each wording,
   no top-1 losses and no target-survival losses. Per-wording survival must
   not fall versus passthrough.
2. Each ranker receives exactly the same 24 candidate IDs and pre-rank scores
   per case, returns the same 24 IDs without duplicates, and the new arm's
   output scores follow the declared rank formula. Each full-frontier scored
   law is identical across arms; each packet matches its own durable journal.
3. Sequential isolated nearest-rank Recall p95 is below 100 ms for both arms.

Report design and confirmation strata, all three wordings, paired event-level
gains/losses, pack survival, law/invariant failures and source/code hashes.
The source snapshot and questions are untouched, but this run follows a
documented interface correction after fixture preparation. It is not an
independent reproduction of the old overlay, an agent-answer study, a loaded
p99 measurement or Goal 5 completion.
