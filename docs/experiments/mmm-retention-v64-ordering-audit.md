# V64 independent-ordering clarification

The complete600-arm collection is unchanged. Original frozen data and the failed
audit remain under retention-v64b-diagnostic. A separate independent diagnostic
reproduces the failure on tight/curved/immediate/uncertainty round9. Candidate42
versus53 flips at the cutoff when independently normalized probabilities near
0/1 differ by endpoint ULPs; entropy scores are approximately1e-13.

No change to learner, nomination rule, stored data or scientific thresholds.
The additional checker still reconstructs EVERYlaw and score under the original
3e-10 numerical tolerance. It independently recomputes stable input-order sorting
of those verified recorded scores and requires exact agreement with25selected
members. Separately, reference-score cutoff regret must be <=6e-10 (twice the
original numerical tolerance). Report all ambiguous reference orders, maximum
score difference and maximum cutoff regret. Do NOT claim independently bitwise
reference rankings in endpoint near-ties.

Duplicates, altered choice scores, wrong laws, receipts, requests, missingness,
expiry, snapshots, metrics or accounting still reject. Re-run corrupted-field
and future-data tests with the revised checker. The original failure and its
source/log hashes are preserved, not erased or called a passed audit.

Model/data source hashes must match the old frozen collection, while the new
checker has its OWNcopied closure/freeze and command log. This is an arithmetic
audit repair, not an empirical rescue or a new confirmatory evaluation cohort.
