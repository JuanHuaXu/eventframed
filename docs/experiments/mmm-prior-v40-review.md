# Prior V40 Mathematical And Lifecycle Review

2026-10-03. Research-only review, not production approval or a whole-goal
completion certificate. The frozen candidate and prospective gates are unchanged.

## Checked Invariants

- Shared and private candidates have identical initial marginal predictions.
  Only shared-family weights use other members' evidence; member rate chains
  remain conditional and distinct. No Anti-Pigeon authority follows.
- Prior, Bernoulli evidence likelihood, transition, and next prediction are
  marginals of one finite joint model. Discrete Beta-shaped masses do not have
  exactly their nominal continuous-Beta mean or pseudocount.
  Changing strength can therefore change the actual mean as well as variance;
  its factorial contrast isolates that contract parameter, not variance alone.
- Late arrivals enter their original nomination position. Replaying a suffix
  replaces the member's old evidence contribution in the global log likelihood;
  it does not multiply earlier labels twice.
- Originally issued forecasts are privately retained for receipts and scoring.
  Future labels cannot retroactively improve an issued score. Pending/cancelled
  slots have unit emissions, while nomination transitions still occur.
- Owner, epoch, ordinal, clock, replay and cap checks precede published mutation.
  All scratch row/evidence/weight normalizations validate before publication.
  Single-owner ownership is required; race-test success is not concurrent-use
  authorization. Epoch changes explicitly discard model state.
- Predict/Issue cost O(HQ), suffix replay O(HQL), and retained state O(MHQL).
  Frozen bounds are H=3, Q=21, M<=200, L<=64. Construction peak during an epoch
  replacement, acquisition, storage and full serving are separate costs.

Measurement erratum: `runPriorV40` allocates its issued/expert-output arrays
before starting the per-arm elapsed timer (the legacy controls do likewise).
Model construction and ticket allocation are timed, but that initial output
bookkeeping allocation is not. Read "complete collector work" in the frozen
protocol as its declared loop boundary, not every allocation in the function.
The 400ms gate is NOT changed; no unmeasured end-to-end cost is inferred.
Raw collector process wall time includes scoring/output serialization and is
reported separately, not divided into a fictitious request-latency estimate.

## Evidence And Falsifiers

The seven model race tests pass: whole-history delayed reference, identical
priors/cross-member borrowing, opaque issued forecasts/lifecycle, atomic
normalization failure, invalid contracts/caps, explicit 21^3 path enumeration,
and interleaved prefixes/censoring. The integration race run passes 13
non-identity semantic corruptions plus missing-allocation rejection and future
prefix invariance. Vet passes.

Diagnostic audit verifies 28 worlds, 924 arms and 15,400 snapshots against an
independent unnormalized whole-history reference, plus full arithmetic replay.
Separate JavaScript readback verifies diagnostic statistics and rejects altered
means/contrasts. Diagnostic n=1 cells have no confidence interval or adoption
verdict. Normal design/confirmation remain separately required.

Concrete falsifiers include: a private unobserved member changing after another
member's label; a replay disagreeing with whole-history integration; a changed
ticket copy changing its receipt forecast; future labels changing a served
prefix; or a failure publishing any row/weight/clock/pending-state mutation.
No new confirmed implementation bug was found in this scoped review. This does
not imply the model assumptions fit the external data-generating process.

## Assumptions And Viable Follow-Ups

The shared calibration family is static. Local rates can reset but family
uncertainty cannot itself switch within an epoch. Its three calibration shapes
also omit the control's richer rank-dependent mean family. These are competing
misspecification hypotheses, not proven explanations or grounds to retune V40.
The rate prior itself is Beta-shaped, whereas the control has additional
finite rate-shape kernels. A NEW study should ablate mean-family breadth and
rate-prior breadth separately; adding both at once would confound the rescue.
Retaining only current conditional rows and recomputing bounded histories is
one possible memory/work tradeoff for a larger family, not a tested speedup.
Noisy feedback is treated as an ordinary Bernoulli label; the true-utility risk
is only an evaluator quantity. Arbitrary selection-dependent delays/censoring
are not modeled or certified by this study.

[Herbster and Warmuth (1998), Section 3](https://mwarmuth.bitbucket.io/pubs/J39.pdf)
motivates tracking changing experts by loss updates followed by weight sharing.
This supports a future switching-expert experiment, not an inherited recovery
guarantee under delayed labels or a claim of ordinary hierarchical Bayes.
The implemented adaptive-window comparator already uses a fixed-share variant;
repeating that name alone is not a new research contribution.

[Knoblauch and Damoulas (2018)](https://proceedings.mlr.press/v80/knoblauch18a/knoblauch18a.pdf)
motivates joint online model uncertainty and changepoint prediction with a
declared model universe. A future richer or switching calibration family must
still define its own evidence/prediction relationship and measured costs.
Neither paper proves our selective sharing, provenance or Anti-Pigeon claims.

Follow-ups must freeze NEW cohorts, preserve V40 negatives, and compare issued
risk, stationary harm and recovery rather than optimizing a consumed endpoint.
The full seven-goal scope, untouched production and original serving deadlines
remain unchanged.
