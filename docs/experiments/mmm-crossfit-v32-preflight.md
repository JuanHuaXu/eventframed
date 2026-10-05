# Exact leave-one-out predictive combination: V32 preflight

2026-10-03. Isolated next lead after the frozen negative V31 study. No V32
outcome cohort has been collected. All seven whole goals remain OPEN;
production, the whitepaper and existing tracked modifications stay untouched.

## Reasoning Gate

Confirmed: V31's independently optimal original-row fit fails broad quality,
calibration and protection screens. Its surviving baseline weight is large
in several failed regimes. Not confirmed: that early immature forecasts,
the ridge anchor, limited outcomes, or inadequate families are the dominant
cause. Exact arithmetic is not evidence for any particular explanation.

The chosen next candidate replaces original prequential rows with exact
leave-one-out child predictions using only currently arrived labels. The
same three children, ridge coefficient1 over SUM Brier losses, simplex,
and (.98,.01,.01) anchor remain fixed. This changes validation construction,
not a threshold or prior on V31. The hypothesis is falsified if omitted
rows disagree with independently rebuilt children, depend on their own
omitted label, or fail a freshly frozen outcome/protection screen.

Adding methods in separate research-only source files leaves all sealed
V27-V31 model files and manifests unchanged. No runtime service imports
the new combination. Because this is a research candidate rather than a
bug fix, no production repair, cleanup, deployment or release is proposed.

## Exact Model and Objective

Let D_t contain at most one arrived binary label per fixed member. For
every i in D_t and each child k, compute x_ik=P_k(Y_i=1 | D_t minus i).
This is a FIRST-label predictive, not the future repeated-member mean
that already incorporates y_i. The baseline entry is its original b_i.
Fit convex weights minimizing:

`sum_i (w dot x_i-y_i)^2 + ||w-(.98,.01,.01)||^2`.

The published research law combines the three ordinary full-D_t FUTURE
child predictives using these fitted weights. Weights are predictive
coefficients, not Bayesian family probabilities. For Bernoulli outcomes
the combined mean defines the full Bernoulli law; no calibration or
Anti-Pigeon certificate follows from proper scoring alone.

Affine posterior masses are divided by the held-out member's original
Bernoulli likelihood and renormalized. For a tree leaf with full counts
(s,f), omission changes its mean to (1+s-y_i)/(1+s+f). The conditional
likelihood of the omitted observation is s/(1+s+f) for a positive label
and f/(1+s+f) for a negative label. Divide full tree mass by that likelihood,
renormalize, and average the omitted means. No fresh model or full
history replay is required for an individual omitted row.

## Boundaries and Cost

Only already arrived evidence enters D_t. Leave-one-out refitting may use
other arrived outcomes that arrived later than i; that is deliberate
validation on today's prefix, NOT a claim that its row was issued before
the historical outcome. Current served/test forecasts must still be
journaled before their own future feedback. No temporal leakage claim
may conflate these two phases.

Initial nomination supports uniform random and predeclared randomized
strata only. Label-adaptive entropy/disagreement nomination is excluded:
removing a label from the likelihood does not undo its influence on
which OTHER labels were observed. Leave-one-out asymptotics are not
inherited for a selectively observed or dependent finite stream.

The fit scans at most200 arrived members and27 components per child;
the fixed3-weight solver still solves at most seven4x4 systems. Measure
update, construction and forecast costs rather than claiming constant
cost independent of the explicit frontier cap. Single-owner, fixed-world,
immediate-feedback only; no epoch migration, durable delayed tickets,
concurrent serving, or source-authentication authority is provided here.

Independent full omitted-member refits, own-label flips, state immutability,
partial-failure quarantine, source hashes, races and measured costs precede
fresh outcome collection. Adaptive selection, changing regimes, agent
answer quality and loaded publication remain separate unresolved tests.

## Primary Source

[Yao, Vehtari, Simpson & Gelman (2018), Using Stacking to Average Bayesian Predictive Distributions](https://sites.stat.columbia.edu/gelman/research/published/stacking.pdf),
equations2.1 and3.2-3.3, motivates omitted-observation predictive scoring
and then combining full-data predictions. This ridge/Bernoulli adaptation
uses exact finite sums, not PSIS or the paper's continuous-data examples.

## Verified Pre-Collection State

Independent complete omitted-member refits agree within2e-13 for both child
families at2,11,150 and200 members, up to all200 observed labels. Flipping
the omitted member's own label leaves its row unchanged. Full-model state
readback is unchanged by omission, and rebuilt combination weights/actual
laws agree within5e-12 across successive prefixes. Unused-label flips,
nonadaptive matched nominations, caller ownership, bad input, duplicate,
unsupported adaptive nomination and partial-failure quarantine pass.
Three-package tests under race and vet pass. A test-harness control policy
and RNG-offset mistake was corrected before collection; no model changed.

Apple M4 three-repeat microbenchmark: construction50.505-51.912us,
76248-76249 bytes/117 allocations;150 forecasts9.553-9.569us;
weight refit with32 labels2.756-2.830us,64 labels4.979-5.036us,
150 labels11.042-11.110us. Forecast/refit operations allocate zero bytes.
These omit child-update, retrieval, persistence, queues and evidence costs;
the [raw output](mmm-crossfit-v32-benchmarks.txt) is retained separately.
No outcome quality or completed-goal claim follows from this preflight.
