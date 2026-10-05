# Mixed-outcome scored-law v22: learning is not exposed in the packet

Date: 2026-10-02. The frozen [service component protocol](mmm-mixed-outcome-v22-protocol.md)
passes its narrow journal/epoch guard conditions in both design and
confirmation. The serving-facing observation is a **failure**: after
feedback, zero of the ten initially packed labelled events remains in
the packet in every world. Do not count the journaled forecast gain as
an answer-quality or Goal 6 success.

**Fixture erratum, 2026-10-02:** the subsequent v23 audit found that this
runner stored `denseRowV6`'s length-two vectors without the unit normalization
required by the research SQL cosine route. Its clamped similarities flattened
most nominee baselines. The raw journal, Beta update, epoch and arithmetic
observations remain reproducible for those stored rows, but the study does
not establish behavior on the intended normalized cosine fixture. Its 0/10
packet-exposure result is therefore a diagnostic on a mis-specified geometry,
not a confirmed normal-cosine serving failure. The separate v24 study repairs
the fixture with durable-vector and service-score oracle checks and fresh seeds.

## Evidence

Each split used four fresh LibraVDB/SQLite instances, 16 distinct
selected nominees per world, one pre-feedback outcome per nominee,
and an independent training-label seed per world. Hidden event-level
Bernoulli probabilities alternate 0.8/0.2 in pre-feedback nomination
order. The service sees only the sampled outcome, never that
probability. Expected Brier below is evaluated against a hypothetical
independent next outcome under the declared probability, not against
the training label. There are 64 selected event laws per split, but
only four independent fixed-geometry worlds; the intervals are not
population or cross-generator guarantees.

| Split | Positive / negative training labels | Journaled base -> learned expected Brier | Paired world gain [mean +/-3.5 SE] | Initially packed / learned packed / postwrite packed selected items |
| --- | ---: | ---: | ---: | ---: |
| Design | 29 / 35 | .430625 -> .390076 | .040549 [.038529,.042569] | 10 / **0** / 10 per world |
| Confirmation | 36 / 28 | .430625 -> .394428 | .036197 [.026807,.045587] | 10 / **0** / 10 per world |

All eight worlds show 16/16 selected nominees still present in the
post-feedback and postwrite **nomination** reports. At the learned
Recall, all 16 have an updated `BeliefLaw`; after 16 visible backfilled
writes and synthetic certificate refresh, none has one. All three
certificate checks are true in every world; evidence epoch moves
217 -> 217 -> 233. The postwrite expected Brier for these nominees
returns exactly to the same-as-of base value .430625. Thus the current
service respects its epoch guard, but the retained learning does not
reach the packed context.

The mechanism is visible in the recorded forecasts. Near-identical
nominees start around 0.925 baseline usefulness; a single positive or
negative observation yields belief-conditioned corrected probabilities
around 0.899 or 0.866. Unlabelled candidates retain the high baseline
and displace every trained candidate from the ten packed slots. The
study does **not** assign a hidden law to those replacement candidates,
so it cannot say whether the replacement packet is better or worse.
It does show that a journaled proper-score improvement can be wholly
decoupled from agent-visible exposure. This is not a failure of the
posterior mean calculation by itself, and no rank fix is justified
without a fresh all-candidate utility law and selection-aware control.

The [design tape](mmm-mixed-outcome-v22-design.jsonl) SHA256 is
`11bda22ba4aa9b319c4206404e29940f5374a7854ad28f7b9d3b2fc3cb830a91`;
the [confirmation tape](mmm-mixed-outcome-v22-confirmation.jsonl) SHA256
is `fa7d10d5e510d7de71b225afb062872d5a7d52444fcd650969c3e10ceacd90bd`.
The [independent verifier](../../research/mixed-outcome-v22-verify.mjs)
checks five source/protocol hashes, all eight identities and epoch
transitions, mixed-label counts, probability bounds, 128 expected-risk
calculations, belief gating, packet counts and data hashes. It passes.
The design split also passes the full opt-in test under `-race`; package
`go vet` passes. Maximum sequential Recall was 12.70/10.95 ms and
maximum outcome publication 5.09/5.23 ms in design/confirmation.
These are not loaded p99 or freshness measurements.

## Decision

The v22 engineering guard component passes; the user-facing learning
objective remains unachieved. Synthetic selection and omitted-influence
certificates are assumed inputs and prove no external-law coverage.
The fixed geometry, rank-order hidden law, and one training outcome
per event cannot establish real-task transfer or general calibration.
Before considering a rank or posterior-reuse rescue, freeze a new
evaluation law for **all** nominees, include held-out answer/packet
utility and false-promotion costs, then compare the current ranking,
an incumbent-preserving correction and any proposed candidate-specific
reuse under concurrent writes. Keep the actual service epoch guard on.
No production, OpenClaw or whitepaper change follows; all seven whole
goals remain open.

Reproduce the artifact check with `node research/mixed-outcome-v22-verify.mjs`.
