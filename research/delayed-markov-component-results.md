# Delayed Markov advice component

Subsequent [v112 quality comparison](../docs/experiments/mmm-markov-v112-results.md)
passes all672 non-harm gates but FAILS13 gain gates. The component-only report
below predates that experiment and does not establish quality success.

2026-09-13. Component and lifecycle checks PASS; quality UNTESTED. This is a
research-only implementation of the [prospective filter](delayed-markov-advice-proposal.md),
not a production change or a rescue claim. V111's failures remain unchanged.

## Reasoning and boundary

The confirmed v111 finding is a failed policy comparison: compatibility-based
pending transfer worsens both delayed switches relative to publication-only
in both phases. It does not prove a programming error in arrival-log weighting.
Possible remaining causes include stale advice, insufficiently fast role
switching and observation selection. This component isolates a different
evidence-timing interpretation; only a fresh quality experiment can establish
whether that interpretation helps.

No upstream released fix is being patched. New private research code leaves
the frozen predecessor kernels, model fits, evidence-routing authority and
artifact history untouched. No data repair, production calls or new private
data are involved.

## Implemented state model

Four role states use the original prior and fixed .001 transition-to-prior
probability. Each issued event has one transition, regardless of when its
label arrives. Unavailable/censored outcomes have unit emission. An arriving
label replaces its own unit emission with its immutable issued likelihood,
then forward messages are recomputed through the retained suffix.

The filter stores a forward checkpoint and up to64 ring entries. A known or
censored contiguous prefix is committed exactly once; later delivery to it
is rejected. Known evidence behind an unresolved head remains available for
refiltering without being admitted twice. After all issues settle, the
checkpoint equals the complete forward filter. Wall-clock flushes after the
256 issued events do not invent more state transitions.

The adapter uses the existing coherent acquisition and gated law at issuance,
then stores the actual raw forecasts from that acquisition. Refiltering cannot
change an archived output or read a current model in place of an old forecast.
The unchanged comparative journal still controls origin ordering, expiry and
version-specific test updates. Prediction, delivery and expiry use atomic
copies; the outer owner also rejects reentrant reader callbacks.

This is numerical evaluation of a declared finite-state model, not proof that
its state transitions describe the environment, that selective missingness is
ignorable, or that calibration improves. Immediate complete-feedback parity
does not establish delayed quality.

## Verification

- Literal enumeration of every short latent-state path agrees with the filter
  for three arrival orders and transition rates0/.001/.2/1. Partial availability
  and fully mixed/zero-transition limits are included. These rates are unit
  boundary checks, not a quality sweep; the runtime candidate remains .001.
- A separate dense full-history recurrence checks all256 issues across four
  ring wraps, reverse-order batches, missing labels, prefix expiry and every
  intermediate update. Committing a checkpoint agrees with retaining history.
- No-movement is not needed for immediate parity: all256 forecasts and complete
  underlying journal states match original log/share exactly with fitted tables
  changing at every publication.
- Delay31 plus missing labels crosses publications and matches the independent
  dense filter. The underlying gate journal equals its unchanged reference
  after each delivered label. Final filter frontiers are256, not288.
- Invalid transitions/probabilities/future labels, duplicate/expired deliveries,
  capacity pressure and reentrant operations fail without partial state changes.
  An injected downstream filter inconsistency also rolls back the gate update.
- Log-domain underflow remains recoverable when transition probability is zero.

Focused tests under race PASS1.465s. Combined Markov/log/compatibility tests
under race PASS1.649s; package vet PASS. No substantive defect was found during
this component round. These deterministic tests do not exercise concurrent
multi-owner use, which is outside this single-owner contract.

## Isolated bounded-lifecycle benchmark

```sh
go test ./internal/observationlearners -run '^$' -bench '^BenchmarkMarkovAdviceJournal$' -benchmem -benchtime=300ms -count=3
```

Apple M4, darwin/arm64, Go1.27.1, suffix10. No other research process was running
during the batch. All18 repeats are retained in the
[raw output](../docs/experiments/mmm-markov-component-benchmarks.txt).
Each operation includes256 adaptive forecasts, eight alternating prebuilt model
publications, expiry scans and full feedback settlement. Fitting, persistence,
network calls and serving queues are excluded.

| Schedule | Arrival log, ms/lifecycle | Markov, ms/lifecycle | Arrival log B/op | Markov B/op |
| --- | ---: | ---: | ---: | ---: |
| Immediate | 1.686-1.717 | 2.057-2.069 | 158200-158202 | 174577 |
| Delay31 | 1.314-1.320 | 3.089-3.111 | 146041-146042 | 162416-162417 |
| Reversed batch, delays0..31 | 1.549-1.560 | 3.207-3.239 | 146040-146041 | 162416 |

Allocation counts remain1276 for immediate and1191 for delayed schedules in
both policies. Immediate forecast paths are identical; delayed paths can differ,
so delayed totals are not pure same-output overhead. These are per256-event
lifecycle measurements, not per-event latency or production p99 measurements.

Recomputation costs O(D*M), with at most64 retained issues and four active role
states here (a fifth zero-prior neutral slot follows the existing representation).
The frozen delay/expiry experiment normally retains at most32 issues. Each
arrival and expiry currently rebuilds the retained suffix from its checkpoint;
it does not cache every interior message. Atomic copies also copy the fixed
capacity buffer. No claim of unchanged O(M) arrival-update work is made.

Source SHA256:

```text
278e100b02382e93c4225d19328eeba029fd009df29afd0388a1f83815930d85 markov_advice.go
d5f6e25610b7b2ef424194c324c167f40f5dec5ddbc77ed6f2a51fb03838e8f6 markov_advice_test.go
212ec7d0c6242d09821474bfdf78a50848afe8c0fa1c72eb02ec7b24c6d45e83 markov_advice_benchmark_test.go
```

Next: fresh complete comparison against the seven original generic/Brier/log
controls, retaining all stationary and recovery gates. Do not trade a passed
unit contract for missing quality evidence, or retune alpha on consumed data.
All seven research directions remain open. No commits, pushes or whitepaper
promotion.
