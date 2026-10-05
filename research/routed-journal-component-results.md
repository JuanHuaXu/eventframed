# Routed delayed-feedback journal: component checkpoint

Status: lifecycle and retention checks PASS. No new learned-stream quality pass
is claimed. This advances the delayed-feedback boundary for directions 1, 2
and 4 after [v103](../docs/experiments/mmm-adaptive-v103-results.md). It does not
fix the older delayed-learning or member-split failures by itself.

## Two distinct meanings of stale evidence

The conservative adapter retires a delivered forecast from a replaced model
version without updating either its selector or the current evidence gate.
The role-carry variant instead retains that forecast's loss for the selector
over four stable learner roles: generic64, Boolean64, generic32 and Boolean32.
Neither variant lets a stale forecast update the newly published model's tests
or routing credits. Roles must keep their meanings; this is not permission to
carry statistics across an arbitrary abstraction split, tenant or task change.

For captured raw forecasts p_i,j and revealed outcome y_i, the selector uses
the original losses (p_i,j-y_i)^2, with the unchanged learning rate .5 and fixed
share .001. It never recomputes those probabilities using today's models.
Current-version records update selector and gate; old-version records update
only the selector in the role-carry variant. This is delayed expert-advice
learning, not an ordinary posterior for one unchanged statistical model.

Keeping a late loss is not necessarily beneficial: stale regime losses can
pull the selector away from what is useful now. The new variant must pass a
fresh quality comparison rather than being adopted on retention counts alone.

## Lifecycle contract

- A single owner serializes issuance, feedback, publication and expiry.
- A bounded 64-entry ring stores original raw forecasts, served probability,
  origin and model version. Slot pressure rejects before reading more evidence.
- Results may arrive out of order. Updates drain only a resolved origin prefix;
  receiving a later result cannot silently train ahead of a missing head.
- Explicit deadline expiry censors unresolved labels, never invents outcomes.
  Already delivered buffered labels are preserved and released when possible.
- Publications occur at origins 0,32,...,224, with a 256-issue horizon. They reset
  version-specific tests before late feedback can reach a new version.
- Failed readers, epoch changes, duplicate/overwritten origins, reentrant calls,
  off-cadence publication and future expiry fail without partial state changes.
- Accounting is issued = pending + applied + bank-only + stale + censored.

Available labels can separately enter a training audit at their arrival time,
even if their old selector record is stale. That audit path is not implemented
inside this journal. The next experiment must enforce as-of training eligibility
rather than using the simulator's already-generated but unrevealed outcomes.

## Verification and retention experiment

Immediate feedback matches the original observer's entire forecast trace and
bank/gate state exactly for all 256 steps, including publication boundaries.
Out-of-order updates match an independently assembled origin-ordered sequence
of captured forecasts. Stale role updates leave the current gate and routing
credits bit-identical. Reader/backpressure/expiry tests cover adjacent failure
paths. Race-enabled component and adjacent tests pass; vet passes.

The deterministic delay grid covers all fixed delays 0..31, both complete
feedback and every fifth label missing, with a 32-step expiry. It checks all
64 schedules, not just the rows below. Each schedule has 256 issued records;
all records are settled at the end. This is a lifecycle fixture, not fresh
outcome-based confirmation or a random missingness model.

| Delay | Missing | Delivered labels | Conservative applied | Stale discarded | Role-carry selector updates |
| ---: | --- | ---: | ---: | ---: | ---: |
| 0 | None | 256 | 256 | 0 | 256 |
| 8 | None | 256 | 200 | 56 | 256 |
| 16 | None | 256 | 144 | 112 | 256 |
| 31 | None | 256 | 39 | 217 | 256 |
| 0 | Every fifth | 204 | 38 | 166 | 204 |
| 8 | Every fifth | 204 | 38 | 166 | 204 |
| 16 | Every fifth | 204 | 38 | 166 | 204 |
| 31 | Every fifth | 204 | 30 | 174 | 204 |

Without missing labels the stale count is exactly 7*delay, matching the seven
publication crossings. With missing labels, origin-order waiting makes even
promptly delivered later labels stale before they can be applied. The carry
variant retains their selector losses, but deliberately does not add those
losses to new-version certificates. This exposes a genuine availability versus
versioning tradeoff that a simple delayed-arrival count would miss.

[Raw test output](routed-journal-component-tests.txt) records the grid and test
results. No claim of Brier improvement, calibration, member-split recovery or
real-agent performance follows from these counts.

## Cost

[Raw benchmark](routed-journal-component-benchmarks.txt), Apple M4 darwin/arm64,
Go benchmark suffix -10, three 500ms repetitions after tests completed:

| Operation: one full 256-forecast lifetime | Time | Bytes | Allocations |
| --- | --- | ---: | ---: |
| Original immediate observer/update | 1.280-1.311 ms | 125,488 | 1,230 |
| Conservative journal, immediate feedback | 1.375-1.536 ms | about 137,776 | 1,231 |
| Conservative journal, delay8 | 1.244-1.252 ms | 135,280 | 1,215 |

These are lifetime timings, not per-request latency. They include acquisition,
trace allocation and updates but exclude fitting, model-copy publication,
storage, retrieval and network service. Delay8 changes forecasts, traces and
the number of applied updates, so its lower time is not an optimization claim.
The role-carry variant has not yet received a dedicated performance benchmark.

Source SHA-256:

```text
faf5b123eab500d91c9b2585504fa2b6f85ec3384a10a6c14f05b2c9d73fcde0  routed_feedback_journal.go
8e8362ec0fad11c2663db73cf0fbe2049ed0c0c14f81f4931c6f73695ef3de24  routed_feedback_journal_test.go
```

## Statistical boundary

[Joulani, Gyorgy and Szepesvari (2013)](https://proceedings.mlr.press/v28/joulani13.html)
analyze delayed online learning and reductions from non-delayed algorithms.
That motivates separating issue and reveal clocks; their guarantees are not
automatically guarantees for this adapter, censoring rule or dynamic expert bank.

[Shafer et al., revised working paper (2010)](https://www.probabilityandfinance.com/articles/33.pdf)
connect nonnegative test martingales and their running maxima to probabilistic
evidence. A likelihood-ratio test needs its null under the actual conditioning
information. Merely preserving eight test allocations does not establish that
condition for delayed or selectively revealed labels.

In particular, the old issued probability need not equal the null probability
conditional on later arrival information, correlated outcomes, missingness or
deadline selection. Origin-order draining preserves the algorithm's order but
does not repair that statistical mismatch. Treat informative delays/censoring
as outside the inherited certificate until a selection model or valid envelope
is established. The next synthetic test can use independently randomized delay
and missingness, but still must not advertise a general delayed-stream theorem.

## Next

Follow [the delayed-routing integration plan](delayed-routing-proposal.md):
freeze a paired quality test of conservative versus role-carry feedback,
retraining only from arrived labels, before returning to member-split audits.
No production changes, whitepaper edits, commits or pushes were made. All seven
research directions remain open.
