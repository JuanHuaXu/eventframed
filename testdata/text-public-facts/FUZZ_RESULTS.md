# Validity-Constrained Fuzzing Results

These results exercise contract 15 against the existing public-fact text corpus.
The frozen audit uses the local 256-dimensional feature-hash embedder, a
normalized two- or three-candidate embedding-nomination law, total-variation
distance, stability threshold `0.05`, required stable probability `0.90`, and
family-wise confidence `0.95`. Per-property bounds are one-sided Wilson bounds
with Bonferroni simultaneous coverage across the three declared properties.

The perturbation families are:

- `context-envelope`: declared `who` and `where` relocation without changing
  the public proposition;
- `same-answer-paraphrase-bundle`: an atomic `what`/`why`/`how` bundle copied
  from another oracle-linked, as-of EventFrame carrying the same answer;
- `cross-domain-semantic-bundle`: an atomic `what`/`why`/`how` bundle copied
  from an as-of public-fact distractor EventFrame.

## Design block

The design split contained 96 eligible queries and 264 trials.

| Property | Stable | Mean TV | Max TV | Simultaneous LCB | Result |
|---|---:|---:|---:|---:|---|
| context envelope | 96/96 | 0.001627 | 0.004730 | 0.954952 | conditional invariant |
| same-answer paraphrase | 72/72 | 0.008228 | 0.023786 | 0.940825 | conditional invariant |
| cross-domain semantic bundle | 90/96 | 0.023673 | 0.086839 | 0.862765 | not invariant |

## Confirmation block

The confirmation split contained 32 eligible queries and 88 trials.

| Property | Stable | Mean TV | Max TV | Simultaneous LCB | Result |
|---|---:|---:|---:|---:|---|
| context envelope | 32/32 | 0.001569 | 0.004234 | 0.876026 | provisional; insufficient simultaneous support |
| same-answer paraphrase | 24/24 | 0.008855 | 0.023754 | 0.841262 | provisional; insufficient simultaneous support |
| cross-domain semantic bundle | 29/32 | 0.024736 | 0.086543 | 0.741564 | not invariant |

No threshold or confidence requirement was changed after observing the
confirmation block. The two perfect point rates do not pass because their
sample sizes are too small under the simultaneous 90% requirement.

The sensitivity ordering is consistent in both blocks: context relocation has
the least movement, same-answer paraphrase is intermediate, and cross-domain
substitution has the most. This demonstrates that the implementation can produce
bounded conditional sensitivity fingerprints. It does **not** demonstrate
autonomous domain-map discovery: oracle links construct the paraphrase family,
and a cross-domain distractor is a negative control rather than a translated
analogy. The high point-stability rate for cross-domain substitutions also shows
that this feature-hash output functional is not strongly discriminative on every
case at threshold `0.05`.

## Local cost

On the Apple M4 development host, the corrected benchmark for one complete
50-event, 64-perturbation, 256-dimensional request measured a median of about
`2.414 ms/op`, `5.12 MB/op`, and `59,231 allocs/op` across five runs. The
predictor is request-local and unchanged frame embeddings are cached within the
request. This is a slow-path benchmark with the local hash embedder; a remote
embedder adds service latency for the base context and each unique perturbed
frame and must be budgeted separately.

Machine-readable outputs are in `fuzz-design-results.json` and
`fuzz-confirmation-results.json`.
