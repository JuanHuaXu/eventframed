# V75 anchored-prefix exact-computation preflight

Status: proposal in an isolated research package; no scientific adoption.

## Pre-Patch Gate

Confirmed repeated work: V74 replays all member issue rows in `prepare` and
`origin`; unknown rows contain no evidence factor. This is not a diagnosis of
the scientific-quality failures. Mean misspecification, sparse evidence, noise,
prior sensitivity and static hyperparameters remain competing quality causes.
There is no upstream runtime patch to adopt: this is an unpublished isolated
finite-model research fork. No production package or frozen V74 input changes.

Keep all 27 mean maps, three dispersion families, three noise states, existing
member/trial caps, likelihoods, same-outcome pair semantics and query modes.
Store the conditional rate belief at the latest ordinal with a revealed factor
(the anchor), not the last issued ordinal. The complete journal remains.
Issuing an unknown trial changes the prospective issue clock but not evidence
or hyperstate odds. Derive the next-rate mean using every intervening transition.
New evidence strictly after the anchor permits incremental filtering. Evidence
at or before the anchor requires replay through the last revealed factor.
Second measurements replace the original first factor; they never advance time.

Smoothing a query at the anchor is the stored conditional filter. Older query
ordinals still use forward/backward smoothing through the revealed prefix.
Unknown future suffixes integrate to one; deleting them from smoothing cannot
delete observed evidence. Unsupported hyperstates remain unsupported; no reset
transition revives zero evidence. Every dependent cache and hyperstate odds must
validate before publication, including prospective count/anchor binding.

The falsifier is any disagreement with the separately implemented dense V74
joint reference in laws, hyperstate posteriors, actual branches, tower identities,
late replay, same-outcome forecasts or public lifecycle behavior. Additional
tests cover out-of-order first arrivals, long unknown suffixes, canceled holes,
zero/full hazard, 64-trial journals, both second outcomes and fault atomicity.
Neither an empirical quality gate nor a numerical tolerance may be weakened.
Small tests do not establish whole-study equivalence or goal completion.

Initial scope: unit/race/vet, full-journal reference and constructor checks after
V74 timed collection terminates. Then serial timing and complete controlled
screen/replay if the implementation survives. The original consumed generators,
controls and gates stay fixed; reserved confirmation seeds remain untouched.
No private or sealed agent outcomes, paper edits, publication or production work.

## Mathematical Invariant

For a fixed member and hyperstate, let `pi` be the normalized finite rate prior,
`T = (1-lambda) I + lambda 1 pi^T` the row-stochastic issue-clock transition,
and `F_n` the revealed likelihood factor at ordinal `n`. An unknown or canceled
row has `F_n = 1`. A first-only row uses `P(W1 | R_n, eta)`; a revealed pair
uses `P(W1,W2 | R_n, eta)` with one shared latent outcome, replacing the first
factor. All these factors retain the original joint model from V74.

Write `a` for the greatest ordinal with a revealed factor. The stored filter is
`P(R_a | F_0,...,F_a, hyperstate)` and its log normalizer. If no factor exists,
`a = -1` and the stationary prior is stored. For `count` issued trials, the next
trial is at ordinal `count`; its conditional rate law is the anchored filter
propagated by `count-a` transitions, or `pi` if `a = -1`. Unknown suffixes do
not alter the log likelihood or the hyperstate posterior. The implementation
uses repeated transitions/normalization, not an untested closed-form power.

When a new first reveal lands strictly after `a`, start with that filter and
apply the missing transitions and new factor. When an old factor changes,
replay every issued row through the new maximum revealed ordinal. No revealed
factor is removed, and no first observation is counted twice by a late pair.
Queries at `a` use the filter; older queries still require the backward factor.
The omitted unknown suffix has total conditional probability one. Combining
each conditional filter with V74's unchanged shared/local/individual hyperstate
weights therefore preserves the same predictive and query law mathematically.

This algebra does not prove floating-point behavior, public API transitions or
correct code. The independent dense reference, numerical tolerance, actual
branch tests and full-journal fault tests are separate required evidence.
