# Sparse representation pilot results

## Design query-mission cross-validation

Eight positive design tasks, whole query mission held out per fold:

| Arm | Top1 | MRR | Candidate Brier |
| --- | --- | --- | --- |
| Existing rank score | 5/8 | 0.8125 | 0.43659 |
| Lexical coverage | 8/8 | 1 | 0.14463 |
| Sparse logistic | 8/8 | 1 | 0.05497 |
| Equal baseline/logistic blend | 8/8 | 1 | 0.16529 |
| Constant 0.1 | 5/8 with baseline tie order | 0.8125 | 0.09 |

Learned arm PASSED the frozen design screen; blend FAILED. Candidate scores from
the existing ranker and lexical coverage are uncalibrated heuristics, so their
Brier values do not constitute a proper-law comparison with the daemon. The
constant prior is the relevant sanity check against trivial class imbalance.

## Frozen Voyager 2 confirmation

After fixing the map and fit, all six confirmation queries were retrieved. No
confirmation feedback was admitted. Five positive queries had support in the
packed ten; one query had no supporting record.

| Arm | Positive top1 | Candidate Brier | Absent-task mean score |
| --- | --- | --- | --- |
| Existing rank score | 3/5 | 0.44702 | 0.66406 |
| Lexical coverage | 5/5 | 0.20965 | 0.45 |
| Sparse logistic | 5/5 | 0.06220 | 0.09154 |
| Equal baseline/logistic blend | 5/5 | 0.17338 | 0.37780 |

Learned arm PASSED the prospective pilot screen. This is one cluster, not five
independent domains. Similar templates and shared corpus documents limit the
inference. Lexical coverage ties the learned rank accuracy; a unique learned
ranking benefit is not established. A low average absent score does not prove
abstention, since an individual wrong candidate can still score higher.

This is an offline rescoring experiment on post-contract service packets. No
OpenClaw or LLM calls, real database, streaming learner, production embedding,
or pre-packing learned serving is claimed. Only the first ten baseline-packed
candidates are rescored; full frontier survival under the new score is untested.
The learner predicts direct task support, not general correctness or truth.

## Audit and next steps

Source/input hashes and all predictions are in sparse-v1-results.json and
sparse-confirmation-v1-results.json. Replay, held-out-label exclusion, feature
bounds and deterministic mapping tests passed. Existing negative experiments
remain unchanged. Confirmation is now consumed for this model family and must
not be reused as untouched validation after any tuning.

Next: different sources and question forms; absent/contradictory/paraphrased
tasks; no learned-versus-lexical superiority claim until those controls differ;
bounded Go implementation and actual full-frontier integration; fixed frozen
publication followed by delayed-feedback and load tests. The seven-direction
research goal remains incomplete.
