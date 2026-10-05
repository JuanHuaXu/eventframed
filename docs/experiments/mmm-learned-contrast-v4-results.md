# Learned contrast v4: evidence-gated null forecast

**Frozen screen: FAIL on design and untouched confirmation.** This is an
isolated Goal 1/7 component experiment, not an EventFrame serving result.
The protocol was frozen before collection. A pre-collection matched-selection
harness bug was traced to floating-point ties under rescaled known priors;
the protocol and harness then used the exact v2 selector state alongside the
new forecast state. The aborted first run emitted no trial rows. No thresholds
or model rules changed after either completed split was collected.

## Artifacts and method

- [Protocol](mmm-learned-contrast-v4-protocol.md),
  [independent verifier](../../research/learned-contrast-v4-verify.mjs),
  [harness](../../internal/observationgate/learned_contrast_v4_test.go).
- Design archive `mmm-learned-contrast-v4-design.jsonl.gz` SHA-256
  `d4629358601b368aa9dde15247ae753100c16a66f6f035dd2213902992fc27a9`;
  confirmation archive `mmm-learned-contrast-v4-confirmation.jsonl.gz` SHA-256
  `ca53a02353cf33af4448080186b1df666ceb0ef831088e7e2270be8c122a2980`.
- Each split has 7 cases x 16 independently fitted baselines x 16 streams =
  1,792 trajectories, each 512 ticks. Each arm requests 128 labels. The
  independent verifier reconstructed 14,336 arm trajectories per split,
  all per-tick scores and generator probabilities, delivery clocks, gate
  counts, recovery, source hashes and matched v2/v4 selection clocks.
  Confirmation replay was byte-identical to its archived gzip output.
- Paired intervals use 16 fit-cluster means and `mean +/- 3.5 SE`, matching
  the earlier v3 analysis. They describe variability across these synthetic
  fits, not population or deployment guarantees.

## Efficacy

Arm 2 is v2 learned without null; arm 3 is v3 always-on null with its
original selector; arm 7 is v4 gated null with the v2 selector. Lower Brier
and recovery delay are better.

| Case and measure | Design v2 -> v4 | Confirmation v2 -> v4 | Frozen verdict |
| --- | ---: | ---: | --- |
| Stable full Brier | .06197 -> .06197 | .06240 -> .06240 | Pass; no post-100 gates |
| Immediate bit2 post Brier | .16140 -> .16772 | .16328 -> .16928 | Known recovery condition fails |
| Immediate bit2 recovery clocks | 127.96 -> 135.95 | 129.11 -> 136.31 | Harm exceeds 5-clock upper screen |
| Delayed bit2 post Brier | .20017 -> .20441 | .19452 -> .19867 | Known recovery condition fails |
| Delayed bit2 recovery clocks | 157.27 -> 163.43 | 153.55 -> 159.05 | Harm exceeds 5-clock upper screen |
| Null full Brier | .37895 -> .29643 | .38146 -> .29461 | Improves, but misses <= .27 |
| Majority post expected Brier | .29299 -> .29474 | .29613 -> .29683 | Fails <= .26 and positive gain |

The null full-Brier gain versus matched v2 learned is .08252
`[.07689,.08816]` on design and .08684 `[.08126,.09243]` on confirmation;
all three gated policies gain at least .05 with positive fit-cluster lower
endpoints. However, their null full Brier ranges .29176-.30511 on design
and .29091-.30213 on confirmation, above the frozen .27 cap. Always-on null
gets .26398/.26308, showing how much the conservative gate gives back.

Matched selection and arrived-label parity hold. V4 learned's bit2 post
Brier gains over gated random and uncertainty are positive with positive
fit-cluster lower endpoints. Its delayed recovery lead is only 7.95/8.73
clocks on design, below the required 10; immediate lead versus uncertainty
is 9.81 on design. The paired old-minus-new recovery-delay interval is
`[-10.42,-5.56]` immediate and `[-9.20,-3.13]` delayed on design, and
`[-9.14,-5.25]` / `[-7.56,-3.43]` on confirmation. Thus the v4 gate
reliably loses known-rule recovery time on these streams. First64 harm is
within the frozen bound, and bit0/interaction post-Brier non-harm bounds pass.

Gate activation after clock 100 is zero on stable parity, about 2.6-4.2%
on known-shift cases, 76-79% on null, and 14-15% on majority OOD for the
learned arm. Because the selection schedule is exactly matched and the
non-gated forecast is the conditional v2 known-rule law, these comparisons
isolate the forecast gate's effect. The few known-case gate activations
explain the observed degradation without requiring a selection change.
The majority case still misses recovery in about 99% of trajectories and
gets no positive expected-Brier gain over v2.

## Cost and scope

The two states total 472 bytes by Go `unsafe.Sizeof`, below 4 KiB. In a
20,000-sample isolated component test, forecast plus selection p99 ranged
125-250 ns across four runs. Paired model-update p99 ranged 3.29-10.71 us;
one of four runs exceeded the frozen 10 us screen, so that cost condition
is not robustly established. This includes both working-state updates but
excludes request handling, storage, scheduling, and model fit. Package tests,
`go vet`, and the targeted race test passed.

**Interpretation:** a high null-evidence threshold protects stable behavior
but is not a free unknown-class rescue. This v4 candidate is rejected on its
predeclared joint efficacy screen and must not be promoted or retuned on
these cohorts. Next research should treat null/OOD detection as a separate
diagnostic or obtain new evidence features; merely tuning the same threshold
on these archived tapes would overfit. All seven whole research goals remain
open. Production and the whitepaper were not changed.
