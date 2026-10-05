# Available evidence with unchanged forecast weights

## Result: Reject Blind View Expansion

The fixed-state diagnostic passes0/4 reverse-recovery gain screens and12/16
nonharm screens. Richer available observations can help substantially, but using
them with the existing narrow-view weights is not a safe general improvement.

Delayed post-Brier, requested view versus already-paid available view:

| Cohort / case | Requested | Available |
| --- | ---: | ---: |
| 1 / stable majority | .048535 | .054847 |
| 1 / stable parity | .144882 | .086316 |
| 1 / majority to parity | .273706 | .269874 |
| 1 / parity to majority | .259748 | .266852 |
| 2 / stable majority | .051261 | .058518 |
| 2 / stable parity | .183083 | .099076 |
| 2 / majority to parity | .270235 | .266394 |
| 2 / parity to majority | .252457 | .269253 |

Stable parity improves markedly: cohort2 accuracy rises65.55% to87.40%.
Stable majority accuracy is nearly unchanged, yet Brier worsens, demonstrating
that classification correctness alone misses probability-quality harm. Both
cohorts' immediate and delayed reverse cases fail the declared nonharm screen.

The delayed reverse mean gains are -.007103 and -.016796; descriptive intervals
are[-.026186,.011979] and[-.033659,.000068]. These are16-trajectory mean +/-3.5SE
screens, not confidence sequences or coverage over the research history.

## What Was Isolated

All256 original credit-learning schedules replay with exactly unchanged issued
forecasts, states, fit histories, feedback, acquisition and credit ledgers. The
diagnostic captures the monitor's cache BEFORE any of the experimental learner
arms request coordinates, then unions that mask with arm2's own requested mask.
It does not reuse bits obtained only by a different experimental arm.

The diagnostic reconstructs the original four experts and weighted forecast,
then reevaluates the same incumbent/short/local-or-pooled/subset models on the
available union. Both inner and outer weights remain frozen at issue time.
No Reader calls, current labels, new fits, hindsight weights or feedback updates
enter this alternate prediction. Every consumed coordinate is independently
checked against the original shadow-monitor masks and actual frame values.

The foreground still ACQUIRES at most six coordinates. A diagnostic forecast
may CONSUME up to nine coordinates because monitoring already paid for them.
This explicit mask distinction is essential; the experiment does not silently
relax a read cap or count a new nine-coordinate query as a six-coordinate one.

## Verification And Numerical Correction

An initial exact-equality unit assertion failed at1.11e-16 difference after
reconstructing default mixture weights. A tooling sequencing mistake launched
collection despite that failure; `mmm-available-evidence-v1.jsonl` is retained
as provisional, not the authoritative artifact. No outcome metrics were used
to change the predictor or experimental thresholds.

The unit check was corrected to use the existing1e-14 reconstruction tolerance,
while retaining exact same-mask recomputation and exact learner-state replay.
Race contracts then passed (1.57s test time), followed by vet and a fresh full
collection. The captured planned evaluator had the analogous bitwise expert
comparison; a separately named verified evaluator applies its existing numeric
tolerance to those reconstructed experts. Maximum expert reconstruction error
was4.44e-16. No gain/nonharm criterion was changed.

Authoritative raw: `mmm-available-evidence-v1-verified.jsonl`, SHA-256
`3a5fb1fd119c7c14aaeb1864a95e3895a18781bc3e21e75ef5f7dcef485cb76e`.
Summary: `mmm-available-evidence-v1-summary.json`.
Verified evaluator: `research/available-evidence-verified-summary.mjs`, SHA-256
`3ca4056e0b391560db88050849e72f80829aed96f4d4ce0b11f7d9f20856c4c5`.

All882 captured source hashes verify. The verified evaluator was added after
collection and is separately hashed above. All128 records /256 schedules match
between the provisional and verified collections despite different unit-test
source metadata. Both replay the original learner exactly. Independent masks,
mixture sums, scores and summaries verify; summary replay is byte-identical.
Collection38.85s provisional,38.60s verified. No full-collection race is claimed.

## Computational Cost

Apple M4, three500ms repeats of one fixed-state law evaluation with count and
subset models present:

- Requested view:37.18/36.84/37.13ns,0 allocations.
- Available view:36.11/35.95/36.10ns,0 allocations.

This is cache-hot immutable-model arithmetic, not acquisition, fitting, queueing,
storage or serving latency. Small timing differences do not establish a speedup,
and cheap evaluation does not rescue the failed quality screen.

## Next Distinct Test

Treat requested-view and available-view forecasts as different prediction
experts whose losses must be learned from their own issued, subsequently
arrived outcomes. Do not assume that a selector trained on narrow-view scores
is calibrated for an expanded view. First inspect prior view/selector experiments
to avoid repeating an already failed method; a new comparator must preserve
origin identity, missing-label handling, split boundaries and actual costs.

The next bounded prequential diagnostic can test whether separate view-loss
accounting captures the stable-parity benefit without reverse-shift harm. It
must freeze its update rule before scoring and must not be described as an
ordinary Bayesian posterior without a coherent model. Success on fixed tapes
would still require closed-loop and untouched confirmation. These results also
leave open whether the local models themselves need better generalization.

All seven goals remain open. Production, whitepaper, remote repositories and
private data were untouched.
