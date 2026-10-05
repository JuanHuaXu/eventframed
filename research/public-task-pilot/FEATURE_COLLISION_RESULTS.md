# Public feature bottleneck diagnostic

Consumed160 captured rows from16 NASA pilot queries,10 packed candidates each.
No new inference or training. Actual Go researchmemory.Extract and
researchsparse.Extract operate on captured5W1H values. Fixed baseline.75 holds
the ninth bit constant to isolate compressed-feature expressiveness. Oracle
support IDs join only after feature extraction; they are not feature inputs.

In10 of13 positive queries, the support shares its entire nine-bit vector with
wrong candidates: eight support buckets have size4 and two size2. Three are
singletons. A feature-only score must tie within each bucket. Even allowing an
oracle to select the right bucket separately for every query, uniform tie
selection gives mean support probability6/13=46.15%. This is a conditional
feature-only diagnostic, NOT actual retrieval accuracy or a universal system
ceiling. Baseline scores, candidate order or other information can break ties;
the real adapter preserves a baseline and is not subject to this simple bound.

Sparse feature vectors distinguish support in every positive query in these
packed rows. This means a distinguishing function exists on this sample, not
that training finds it or that it generalizes. Prior Nobel semantic-embedding
tests already show learned sparse direct order can regress versus semantic
retrieval. Do not replace semantic ranking based on this separability result.

Three unsupported cases have no positive bucket and are excluded from the
positive-case statistic, not scored as correct. The audit says nothing about
abstention calibration. Candidate rows are correlated, already consumed and
post-packing; no claim about the full frontier or untouched confirmation.

Artifacts: `feature-collisions.json` (source hashes, raw features, support labels),
`analyze-feature-collisions.mjs` (hash checks and bucket accounting). The Go
runner is `research/public-feature-collisions/main.go`. Exact fixture mapping
requires known captured field strings and refuses ambiguous/unmapped rows.

## Next research implication

Retain semantic retrieval as the incumbent. Test richer observation features
or semantic-conditioned residuals, with counterfactual feature collisions and
untouched tasks, before assuming a more elaborate nine-bit learner solves
semantic discrimination. This is a concrete representation lead for directions
4/5, not a proposal to expose IDs or oracle outcomes to the model.

The local generation endpoint was rechecked: only the existing nomic embedding
model is installed. Actual agent-generation calls remain0; no model download,
credentials or production access were used. Other research remains actionable.
