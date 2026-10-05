# Delayed trial identity V36: reasoning gate and preflight

2026-10-03. Goal ACTIVE, all seven whole goals OPEN. V35 sources and
cohorts stay frozen. This is a new isolated research adapter, not a runtime
patch, production change, or an authorization to install/deploy anything.

## Confirmed Boundary and Alternatives

V35 Observe requires the next arrival ordinal. That is correct for its
immediate-feedback contract, not a software defect. A delayed caller can
receive issue 2 before issue 1. Passing issue ordinals would reject genuine
feedback; passing arrival counts alone would permit replay and lose the
original forecast. Alternatives are buffering until issue order, dropping
late evidence, or identity-bound exchangeable arrival updates. Buffering
adds head-of-line blocking; dropping discards valid observations. We test
the third alternative under the existing stationary model, not all temporal
models. No upstream bug fix is being invented.

Happy path: issue private pre-outcome forecast/ticket -> pending -> exactly
one resolution OR cancellation. Several unresolved trials/member are allowed.
Resolution can arrive in any order; the same identity cannot count twice.
Epoch rotation creates a fresh model and invalidates every old ticket.
All valid operations use nondecreasing as-of ticks. Invalid operations are
atomic. The controller, not this adapter, establishes epoch boundaries.

State is bounded by N*64 issued slots, N<=200, and an explicit pending cap.
Cancellation consumes its ordinal, adds no negative label, and cannot be
reissued within the epoch. Source authenticity/independence is not inferred
from an opaque ticket. The caller still supplies genuinely distinct trials.
There is no automatic timeout, missing-at-random assertion, durable store,
thread-safety claim, or production source/tenant/SCM authority.

## Joint-Model and Timing Contract

Conditional on fixed member rates, distinct trials are iid Bernoulli. Arrival
updates are commutative batch-likelihood factors. For the same arrived set,
the shape law must agree with a separate batch Beta/atom calculation regardless
of order. Issue ordinal names the trial; received-count ordinal updates the
exchangeable learner. The receipt scores the private ORIGINAL forecast,
while the update uses the current conditional evidence likelihood. Reusing
an issue-time likelihood for today's posterior would double/miscount evidence.

Delays independent of unobserved usefulness are the primary model condition.
Outcome-dependent delays/censoring can make arrived evidence selective. In
that case this simple likelihood is not a justified full-stream posterior;
retain the failure and require a selection/delay joint model. No hidden
future label, completion order, true rate or missing-label imputation enters
Issue. A falsifier is disagreement with batch arrived-set inference, any
mutating invalid call, replay acceptance, changed issued forecast or an
earlier law changed by future-outcome flips.

## Primary Research

[Joulani, Gyorgy & Szepesvari (ICML 2013)](https://proceedings.mlr.press/v28/joulani13.html)
formalize timestamped delayed feedback, including out-of-order arrival,
and distinguish stochastic/adversarial delay costs. This motivates the
explicit issue/arrival boundary; their regret bounds are NOT inherited by
this finite Bayesian adapter. Our stationary order-invariance follows from
its own integrated likelihood and must be checked directly.

## Prospective Verification

Test multiple pending trials/member, reverse/random arrival, independent
batch integrals, original score binding, invalid/owner/replay/cancel/epoch/
time/cap atomicity, issue no-learning and future-outcome flips. Measure
Issue/Resolve/epoch reset and bounded memory separately. Broader noisy,
shifted/delayed controlled cohorts remain required for Goal1; no synthetic
adapter test alone completes any whole goal. Production is untouched.
