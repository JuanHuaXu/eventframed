# Conditional authorization: fresh-seed forecast component

## Decision

The [predeclared fresh-seed test](mmm-conditional-gate-v2-protocol.md) passes
all four toy screens. With the *same* reference, live frames, selected audits,
arrivals, and local evidence cache in both arms, only the context-conditioned
gate authorizes a useful update after the declared conditional swap. Stable
arms never split and emit identical forecasts. This is a genuine prequential
forecast benefit in this fixed two-context toy family, not a pass of the
existing EventFrame Anti-Pigeon gate or Goal 3 as a whole.

| Schedule | Regime | Scalar / conditional flags | Scalar / conditional mean Brier | Paired realized gain, mean [3.5-SE interval] |
| --- | --- | --- | --- | --- |
| Complete immediate | Stable | 0 / 0 of 1,000 | .090101 / .090101 | 0 [0,0] |
| Complete immediate | Swapped | 0 / 1,000 of 1,000 | .725593 / .133106 | .592487 [.589034,.595939] |
| 25% audit, 20% missing, delay 0..31 | Stable | 0 / 0 of 1,000 | .090329 / .090329 | 0 [0,0] |
| 25% audit, 20% missing, delay 0..31 | Swapped | 0 / 1,000 of 1,000 | .725795 / .325383 | .400412 [.393675,.407149] |

The hidden-law expected-Brier calculation, used **only by the evaluator**, gives
gains .592111 and .400288 in the two swapped schedules. This agreement with
realized Brier is an arithmetic cross-check, not an independent sample. In
the sparse delayed swapped schedule, the conditional gate flags at clock
median183/p95 263, with about99.005 labels actually delivered by horizon on
average out of127.724 nominations;25.651 are missing and3.068 are still
pending. Complete-immediate detection median33/p95 46. The reference itself
costs512 labels per trial and is not included in those live audit counts.

## What the experiment tests

The fixed old context law is `(0.9,0.1)`, and the new law swaps those values.
The scalar correctness stream has exactly the same Bernoulli(1/2) law in both
regimes. Both arms retain all available `(X,Y)` audits, but one authorization
test discards `X`. Before authorization, each issues its reference Beta(1,1)
mean. After authorization, it issues the per-context Beta(1,1) mean of its
own **already arrived** live audits. Forecast probabilities are computed before
the same frame's outcome is generated or any zero-delay label is delivered.
No inference code receives the synthetic target probabilities.

This is why the large gain is unsurprising: the old law is confidently wrong
after a deliberately complete conditional swap, and the conditional gate
permits switching to directly observed local estimates. It is still useful
evidence that the v1 identifiability improvement reaches a properly scored
output under honest timing, not merely an alarm count. It does not establish
benefit for subtle shifts, sparse or learned contexts, source dependence,
adaptive audits, calibrated general Bayes, or any real agent task.

## Reproducibility and checks

- [Raw 4,000-trial artifact](mmm-conditional-gate-v2.jsonl), SHA-256
  `fdf920205ac0c1f768938fd7889b4d2905aa613a661dbb936f4f3e34e6976e29`.
- [Simulator](../../research/conditional-gate-v2.mjs), SHA-256
  `e2ff19456bd28809f1b3a7cdb27c57d36a3c0ac3779da76c8a0b6a0cbe9ad60d`.
- [Independent row/summary verifier](../../research/conditional-gate-v2-verify.mjs)
  passes all4,000 unique records, evidence budgets, score bounds and grouped
  paired statistics. A separately generated replay matches all trial rows and
  the summary apart from elapsed runtime.
- Negative controls reject future, missing, un-nominated and duplicate audit
  delivery. The scalar and conditional arms' retained labels, counts and
  successes agree after every clock. Input, outcome, nomination, missingness
  and delay use separately seeded deterministic streams.

The whole four-cell simulator took about0.58s for two million toy frames on
this host, including random generation and assertion work. This is not a Go
daemon benchmark, request p99 or end-to-end learning cost. The declared gate
and per-context estimator perform constant work for two contexts, but an
unbounded set of learned contexts would not retain that bound.

The false-revocation guarantee is only the finite-horizon two-cell union
bound in the protocol. It assumes independent bounded labels, a fixed context
partition and outcome-independent audit selection/arrival. No certificate is
claimed for the unknown external-law diameter of an EventFrame abstraction.
The present toy gives split authorization after a new regime begins at the
first live frame; it has no overlapping event graph, sheaf-inspired neighbor
selection, source-authentication, or later reversal. Prior Goal 3 failures
with the actual low-weight pooled/local slot remain unchanged.

## Next boundary

An actual MMM research integration would have to bind a small *predeclared*
context partition to independently nominated full audits, allocate error
across all active buckets and repeated decisions, and compare the changed
forecast law against the existing gate on fresh shifted/stationary trajectories.
Context creation, provenance and informative feedback cannot be skipped by
copying this toy threshold. Keep the original 0.005 post-Brier gain, false-
revocation and cost screens; do not infer them from this stronger toy swap.
This turn changed no production code, whitepaper or remote repository.
