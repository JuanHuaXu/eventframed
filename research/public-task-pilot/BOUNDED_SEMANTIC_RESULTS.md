# Bounded correction preserves baseline, adds no accuracy

**FAIL the improvement criterion; PASS finite non-harm and rank/law separation.**
42 actual CaptureTurn/Recall runs on consumed Nobel/NASA tasks, local semantic
embedding, all19 records retained. Sparse weights and correction scale frozen.

| Arm | Literal top1 | Paraphrase top1 | Positive support retained |
|---|---:|---:|---:|
|Semantic baseline|6/6|5/6|12/12|
|Sparse direct ordering|6/6|2/6|12/12|
|Bounded additive sparse correction|6/6|5/6|12/12|

Maximum absolute rank delta0.004767037796067797, below.005. All19 candidate
forecast laws per query are exactly equal across arms in the recorded outputs
(max numerical difference0). Therefore this cannot be presented as improved
proper-score calibration. Both unsupported tasks remain unsolved by this
experiment; no abstention rule was fitted or evaluated.

The bounded arm avoids direct sparse ordering's paraphrase regressions by
preserving semantic information, but repairs none of the baseline's errors.
This is useful failure containment, not evidence that the residual learned
additional semantic knowledge. Do not tune the scale on these consumed results
or claim a new confirmation sample. Richer sparse features alone do not provide
a validated semantic correction signal.

Raw scores/laws/source hashes are in
`nobel-v1/bounded-semantic-results.json`; verify with
`node research/public-task-pilot/check-bounded-semantic.mjs` (passed:false for
the combined criterion). Runner: `cmd/public-semantic-bounded/main.go`.
Memory backend, not persistent LibraVDB; sequential local public-text embeddings,
no generation or production. No concurrent performance claim.

Next research needs genuinely new information that the baseline lacks, such as
verified contextual corrections or entity/role disambiguation, rather than a
stronger blend of the same failed sparse signal. Such evidence must arrive
before prediction and be independent of the evaluation oracle. Until then,
semantic ordering remains the incumbent; no production policy change.
