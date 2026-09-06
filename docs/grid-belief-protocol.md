# Bounded grid belief: frozen protocol

## Critique and scope

Confirmed: the optional two-hypothesis working law is confined to approximately
[0.201484, 0.798516]. Its posterior odds may reverse, but evidence cannot produce
a usefulness probability outside that interval. This is a modeling limitation,
not a claim that every downstream calibrated forecast has that range.
Needs investigation: real signed-feedback accuracy, source dependence, certificate
coverage, and high-concurrency storage latency. This experiment cannot settle them.

Implement an opt-in fixed-share Bernoulli grid, leaving defaults and the old
two-hypothesis mode unchanged. Twenty-one hypotheses: .01, .05, .10, ...,
.95, .99; uniform reset prior; reset share .02 per admitted observation.
Prediction uses the next-observation mixture, before seeing its outcome. Update
uses that prior and a likelihood raised to the admitted weight clamped to [0,1].
Invalid/mismatched state resets to uniform. No imported Beta certainty. No scans
of historical evidence; fixed-size value storage. Existing trust, replay,
Anti-Pigeon, scoring and residual gates remain authoritative.

Research basis: Herbster and Warmuth (1998), Tracking the Best Expert,
https://doi.org/10.1023/A:1007424614876 ; Bousquet and Warmuth (2002),
https://jmlr.org/papers/v3/bousquet02b.html (fixed share to start vector).
This is a finite reset-HMM instantiation, not a new inference algorithm or a
claim to implement full run-length Bayesian changepoint detection.

## Frozen test before first execution

No parameter search. Design seed base 2026090601; confirmation 2026090602.
64 independent trajectories per scenario, 1000 observations each. Predict before
update. Scenarios: stationary .01, .20, .50, .80, .99; abrupt .99 to .01 halfway;
recurring .9/.1 every 100 observations; linear drift .1 to .9.
Controls: existing two-hypothesis filter and ordinary Beta(1,1) sequential mean.
Report per-trajectory raw Brier and log loss, plus Brier of the declared
noncontextual composition .9*.5 + .1*belief (identity calibration, no residual).
The latter is a scoring-contract simulation, not a retrieval or LLM benchmark.
Paired gain is old minus new. Report trajectory-normal 95% intervals and
Bonferroni simultaneous intervals over 8 scenarios x 3 metrics. Keep all results.
Success for the narrow range-bias claim: positive simultaneous Brier gain on
both stationary extremes. No universal superiority hypothesis. Report losses to
Beta on stationary streams rather than hiding that control.

Mechanism tests: exact one-step Bayes, probability normalization, invalid state,
zero/invalid/oversized weights, policy mismatch, reversal, reset, persistence,
replay, served law and score integration. Audit prediction/update timing and
copy semantics. Run full suite, vet, race checks. Benchmark old and grid with
authentication held constant: primitive updates and existing isolated disk
internal requests, 1/4 workers, recall-only/mixed, 500 operations x 3 repetitions.
Report mean throughput cost separately from per-request p99; no SLA conclusion.
Only synthetic data and public documentation may be published.
