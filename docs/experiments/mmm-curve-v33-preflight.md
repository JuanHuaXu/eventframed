# Fixed-model label-budget curves: V33 preflight

2026-10-03. Research-only diagnostic after V31/V32 failed broad quality,
calibration and protection. No V33 outcomes collected at this preflight.
All seven whole goals remain OPEN; production and whitepaper untouched.

## Reasoning Gate

Confirmed: V32's exact LOO fit improves some curved-case predictions but
large packed-confidence defects and reversed-case losses remain. Needs
investigation: scarce labels, regularization, latent-family mismatch and
member-specific variance assumptions remain competing explanations.
This is NOT a justified production patch or an identified runtime bug.

The next experiment changes ONLY observed-label volume. Child models,
priors, ridge1 over SUM Brier losses and anchor(.98,.01,.01) stay frozen.
Six learners share the same label-independent nomination tape. Evaluate
prefixes16,32,64,128,150, including future risk at ALL150 members rather
than scoring training labels again. The evaluator alone knows true rates.
Separate future risk among observed and unobserved members to expose a
changing mixture of those populations, not mistake it for generalization.

A fresh full-frontier failure falsifies the claim that scarcity of distinct
member labels alone explains the remaining defect on these generators.
It does NOT identify the dominant model problem or prove that more repeated
outcomes, different features, or delayed agent evidence cannot help.

## Sampler and Phases

The fixed32 helper stratumV30 cannot track repeated passes; using it beyond
32 could repeat already seen evidence. Add an isolated label-independent
sampler with the SAME bit-reversed32-bucket schedule, unseen bookkeeping,
uniform draws within the next nonempty scheduled bucket, and exact
conditional probabilities. Cross-check every draw against the frozen
Model.Select implementation through exhaustion. Candidate nomination
marks its index unavailable before feedback; every learner then receives
that distinct label once. No label enters nomination decisions.

For each nomination, issue all six current forecasts BEFORE reading its
outcome. Stack uses the original opaque ticket and retained child row.
All models update after arrival. Snapshot forecasts/weights/LOO rows at
the five declared prefixes without changing model state.

Latent inputs reuse the frozen V30 generator. It also executes legacy
control arms on the evaluator side; those discarded runs are not evidence
for this study or part of learner costs. Record input-generation time
separately. Change the geometry seed stride to100million, by an explicit
90million offset to the generator argument, to prevent overlap between
geometry1/regime0 and geometry0/regime10. Earlier per-cell studies remain
unchanged; their32-world per-cell intervals did not claim cross-cell
independence. No data repair or retrospective resampling is proposed.

## Audit and Cost

Learners are single-owner bounded research models, not daemon integrations.
Unit falsifiers: repeated nominees, wrong conditional probability, future
labels affecting an earlier prefix, incorrect LOO rows, changed historical
ticket rows, and snapshot queries mutating future state. Test small and
full frontiers, input failures, exhaustion and matched controls; freeze
collector/tests/protocol/checker hashes before collection.

Record per-model setup, cumulative pre-feedback probe and update time,
current snapshot time, previous snapshot time, and their AccountedNS sum.
This SUM is measured phase work, NOT standalone wall latency. Nomination
is shared external work; record it separately. ElapsedNS covers the six
interleaved learners and diagnostics, excluding input generation and
evaluator true-rate scoring. No hidden acquisition cost or serving claim.

Exact seed replay and independent joint-model/QP/tape/score reconstruction
cover both splits. Preserve negative results; no ridge/threshold sweep on
this consumed cohort. Prefix comparison does not establish an equal-cost
falsification policy, changing-regime adaptation or external AP authority.

## Primary Methodological Source

[Viering & Loog, The Shape of Learning Curves: A Review](https://arxiv.org/abs/2103.10948)
defines evaluation against sample size and discusses nonmonotonic learning
and misspecification. Here the nested label budgets and expected future
Bernoulli scores are our bounded-world adaptation; no iid, universal-shape,
asymptotic or monotonicity guarantee is inherited.

## Pre-Collection Verification

The isolated sampler matches frozen Model.Select exactly through full
exhaustion at2,11,150 and200 members on three seeds, with no repeated
evidence. Earlier prefix laws/weights/LOO rows survive flips of unarrived
labels at16,32,64 and128. Suppressing diagnostic snapshots leaves all150
later issued laws/rows unchanged. First32 laws match frozen local, affine,
partition and blend runners bit-for-bit. Observed/unobserved future risks
reconstruct total risk; the empty150-label unobserved group stays null.
Phase sums are checked explicitly, and shared nomination timing includes
sampler construction. Independent auditor omission/simplex self-tests,
three-package race tests and vet pass before outcome collection.
