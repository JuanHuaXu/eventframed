# Design preflight: real-field transfer boundary

Run 2026-09-12 with cmd/public-pilot-preflight. Raw output:
[design-preflight-v1.json](design-preflight-v1.json). This used the actual
CaptureTurn -> internal extraction -> Recall -> packing service flow, but with
an in-memory store and 256-dimensional hash embeddings. It is NOT an OpenClaw,
libravdb, production-embedding, LLM-answer, or latency experiment.

Each design query gets a newly constructed service and freshly captured corpus.
The corpus consists of the 13 previously verified public facts. No private data
is used. Oracle labels are loaded only after all retrievals finish. All ten
design queries ran; confirmation queries and LLM generation remain unexecuted.

## Observations

- Eight of eight answer-bearing design queries retain their support in the
  packed ten. Five place it first and three second. Neither ambiguous nor
  absent queries have a designated single support; rank zero is not an answer
  failure metric for those tasks.
- Six of eight positive queries have incorrect packed candidates with exactly
  the same research-5w1h-overlap-v1 signature as the support. Counts are 3 each
  for the four Mariner flyby queries and 1 each for Voyager 1 Jupiter/Saturn.
- Actual extracted fields often put a date in `where`, `user and agent` in
  `who`, and the original assertion in the `what`/`why` fallbacks. Therefore
  nonempty fields are not evidence that semantic roles were extracted correctly.

The feature collision is confirmed. It means the count/tree experts using these
bits cannot distinguish those candidates by features. The outer model can
still differ through its continuous baseline input; this is not an impossibility
theorem for the complete adapter. The generic extractor's role assignment is
also observed, but a production extractor repair requires broader regression
analysis and is not performed here.

## Diagnostic rescue

feature_audit.py preserves which query terms occur in the compressed frame,
instead of collapsing matches into nine threshold bits. On the same design
output this reduces support/negative collision cases from 6/8 to 0/8. The
signature is computed without record IDs, answer labels, or full-text metadata.
Labels are used only to evaluate collisions afterward.

This is a post-design representation diagnostic, NOT a trained policy result,
accuracy gain, generalization result, or confirmation. It motivates a bounded
sparse-feature learner rather than forcing real retrieval through nine synthetic
coordinates. Query-term signatures also do not solve negation, roles, paraphrase,
missing evidence, or arbitrary unseen predicates.

## Next decision

Do not claim retained MMM transfer from this adapter. Preserve v1 and its failed
mapping evidence. Test a versioned sparse-feature challenger, with capped query
terms and retained baseline, against lexical coverage and baseline controls on
new sources and changed question phrasing. Fit only on design feedback and freeze
before confirmation. Keep semantic role extraction as a separate ablation so a
learner improvement is not confounded with an ingestion change.

Validation: the command completed, feature audit reproduced 8/6/0 totals, and
go vet ./cmd/public-pilot-preflight passed. No fitting parameters were changed,
no production services were touched, and no existing evidence was overwritten.
