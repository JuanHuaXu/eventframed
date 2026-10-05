# Hazard-mixture component

Component checks PASS; quality UNTESTED. The research-only implementation uses
the exact fixed family in the [proposal](hazard-mixture-advice-proposal.md):
11 rate hypotheses, prior mass.5 on.001 and.05 on each other rate, four role
states, and the existing delayed checkpoint/gate contract. No rate was selected
by scoring consumed v112 quality outcomes. No production integration or promotion.

## Reasoning gate

V112 establishes all672 non-harm checks but not the required recovery gains.
An unsuitable fixed rate is a hypothesis, not a confirmed cause. Model capacity,
state misspecification and observation choice remain alternatives. This is a
new experimental inference rule, not a repair to an upstream released bug.

The chosen invariant is joint inference over rate and role: preserve each
rate's evidence mass while propagating its conditional role distribution.
Normalizing every rate independently would make the rate posterior incorrect.
The implementation marginalizes rates before passing the same four raw-role
weights through the unchanged evidence gate and observation controller.

The rate remains constant along a latent path. The model does not infer a
piecewise-varying hazard or certify a changepoint. Label availability and the
observation stream are conditioned upon under the declared research model;
arbitrary selection bias or real-world calibration is not proved by exact
finite-state calculation.

## Verification and bug hunt

- Literal enumeration across all rate/state paths agrees for three feedback
  orders, including intermediate partial availability and rate posteriors.
- A point mass at every one of the11 rates agrees with its fixed-rate Markov
  filter. The.001 point-mass journal also matches256 immediate forecast traces
  and role weights with model changes at each publication.
- Missing or role-independent likelihoods preserve rate priors. A contrasting
  fixture actually changes them, ruling out a vacuous non-learning implementation.
- Dense full-history filtering checks four ring wraps, reversed deliveries,
  missing labels, prefix checkpoints and expiry, including the rate marginal.
- Delay31 journal integration matches the dense reference at every clock and
  preserves the unchanged gate's delivery result. Wall-clock flushes do not
  invent issued events. Duplicate, expired and future labels are rejected.
- Invalid priors/emissions and capacity violations are atomic. Failed reader
  callbacks and an injected downstream filter failure cannot partially commit
  the gate. Reentrant calls are rejected by the original owner as well as its copy.
- Zero-rate log evidence remains recoverable after probability underflow.

Initial component race checks pass1.794s; journal plus component checks pass
2.026s. Combined hazard/Markov/log tests pass7.808s before the arithmetic
optimization and7.083s afterward. The added point-mass journal test passes
under race1.320s. Package vet passes. No substantive behavior bug was found
in this round; no concurrent multi-owner API is claimed.

## Equivalent arithmetic optimization

The initial kernel performed all transition sums in log space. Its benchmark
exposed substantial cost in repeated transition calculations. The optimized
kernel computes a rate's conditional role transition and then restores its
log evidence mass. That restoration is essential; it is not independent
normalization of competing rate models. The zero-rate branch retains original
log values to avoid permanently erasing underflowed states.

The original all-log kernel is retained as a test reference. All256 steps of
an extreme-likelihood/missing-label fixture agree in normalized log weights
within1e-9. Independent path and dense-filter tests still pass. This was an
algebraic optimization before any fresh quality evaluation, not a rate/prior
change or a result-driven tuning operation.

## Isolated performance

Both complete batches used:

```sh
go test ./internal/observationlearners -run '^$' -bench '^BenchmarkHazardAdviceJournal$' -benchmem -benchtime=300ms -count=3
```

Apple M4, darwin/arm64, Go1.27.1, suffix10. Each operation has256 forecasts,
eight alternating prebuilt publications, expiry scans and full settlement.
No other research process ran during either batch. Fitting, persistence,
network and serving queues are excluded. All27 repeats in each batch are
retained: [initial](../docs/experiments/mmm-hazard-component-initial-benchmarks.txt)
and [optimized](../docs/experiments/mmm-hazard-component-benchmarks.txt).

| Schedule | Original log | Fixed-rate Markov | Initial mixture | Optimized mixture |
| --- | ---: | ---: | ---: | ---: |
| Immediate | 1.729-1.734ms | 2.096-2.165ms | 5.186-5.207ms | 4.870-4.900ms |
| Delay31 | 1.338-1.344ms | 3.150-3.165ms | 47.829-48.027ms | 39.491-39.604ms |
| Reverse batch, delays0..31 | 1.601-1.604ms | 3.273-3.305ms | 38.365-38.678ms | 31.638-31.713ms |

Control columns above come from the optimized batch; the original batch's
control repeats are also retained. Optimization reduces mixture lifecycle time
by about17-18% in these delayed schedules, but the delay31 total remains about
12.5 times the fixed-rate total. No quality gain has yet justified that cost.
Unlike the fixed-rate immediate identity case, default-mixture observation paths
can differ from controls; totals are not pure same-output overhead estimates.

Optimized mixture allocations are about174576B/1276 allocations immediate and
162416B/1191 allocations delayed, essentially the same benchmark allocation
class as fixed Markov. That does not imply equal computational work. Refiltering
is O(D*H*M), H=11 and M=4 (stored with an excluded neutral slot), with D bounded
by the explicit64-entry ring and normally32 under the frozen expiry contract.
These are per256-frame lifecycle times, not per-event latency or production p99.

## Frozen component sources

```text
718526e5bb17770a7af931db7e55a6db1d52839e8e2ed0bc5b8161a4d0ad3d2f hazard_advice.go
f9cc776b3758b180e98f7f270e0d4a4d04874bc97a9c25fad33afab4f098b8a8 hazard_advice_test.go
565d978104f9b821eabd00627d702559b3b70492af9c0096bd4fd89b5124e799 hazard_advice_journal_test.go
0c4ef1a14f237aed98f7997176d880e79ac61d5c30fc0aee6165d2c0888ab151 hazard_advice_reference_test.go
0a8ef843dbe582ab88a9f877e0cd34234e8e4d8715e1dfa44c36f84ce1606529 hazard_advice_identity_test.go
142ab6705b8ebd32782ffe3a5d8c99d756278b777ee9cc6cdaeb2a9a95636c29 hazard_advice_benchmark_test.go
```

Subsequent [v113 quality test](../docs/experiments/mmm-hazard-v113-results.md)
completed generation and full replay: FAIL, 804/818 gates (767/768 non-harm,
37/50 gain). Component correctness does not establish a useful rescue.

Original next step: freeze a fresh complete quality comparison with the original controls
and fixed-rate Markov. Preserve all stationary and recovery requirements; do
not substitute a unit-contract pass for missing forecast-quality evidence.
All seven research directions remain open. No private-data expansion, commits,
pushes or whitepaper promotion.
