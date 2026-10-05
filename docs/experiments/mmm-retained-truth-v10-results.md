# Retained challenger under a new outcome family

2026-10-01. The [frozen protocol](mmm-retained-truth-v10-protocol.md)
substituted a majority-rule outcome family for v8's parity-to-bit family,
while preserving the four arms and their original gates. A research-only Go
overlay changed exactly two input draws, two outcome calls and two seed
expressions. Design and confirmation now have separate base-fit and
evaluation seeds. Both uniform and latent-correlated input modes ran 480
streams of 512 frames each. The original v8/v9 sources and results remain
unchanged.

## Frozen verdict

**FAIL in both input modes for all three challenger arms.** Stationary
protection and the all-scenario mean-harm floor pass, but neither primary
shift meets the required gain/lower-bound combination. Shift256 also has a
negative fit-group sign for every arm in both modes. The earlier v9 static
success therefore does not transfer to this new synthetic outcome family.

Confirmation post-change Brier gain versus fixed MMM, with frozen paired
3.6-SE lower bound in parentheses:

| Input | Arm | Shift128 | Shift256 |
| --- | --- | ---: | ---: |
| Uniform | Replacement | 0.00404 (-0.00062) | 0.00028 (-0.00158) |
| Uniform | Adaptive retained | 0.00356 (-0.00041) | 0.00008 (-0.00105) |
| Uniform | Static retained | 0.00254 (-0.00036) | 0.00027 (-0.00069) |
| Latent | Replacement | 0.00195 (-0.00240) | -0.00015 (-0.00213) |
| Latent | Adaptive retained | 0.00223 (-0.00181) | 0.00017 (-0.00099) |
| Latent | Static retained | 0.00333 (-0.00010) | 0.00051 (-0.00105) |

This is not an implementation error that may be discarded: independent
arithmetic reproduced all 120 paired comparisons and all three verdicts for
each mode. The new outcome family was fixed before the run. We do not tune
it, weaken the v8 gate, or relabel v9 as a general result.

## Hindsight diagnosis

The separate [oracle diagnostic protocol](mmm-retained-truth-v10-oracle-protocol.md)
was frozen after the failure and cannot count as new confirmation. It
reconstructed all 49,152 archived outcomes across both modes from the
declared role-0 RNG and verified all 196,608 final selected-view mask/value
pairs against the reconstructed inputs. It then enumerated the exact conditional law on each
arm's actual acquired bits, plus an unattainable full-input law.

Post-change realized Brier, rounded:

| Input/case | Fixed model | Fixed-view oracle | Replacement model | Replacement-view oracle | Full-input oracle |
| --- | ---: | ---: | ---: | ---: | ---: |
| Uniform shift128 | 0.2503 | 0.1989 | 0.2463 | 0.1707 | 0.0506 |
| Uniform shift256 | 0.2576 | 0.2196 | 0.2573 | 0.1769 | 0.0503 |
| Latent shift128 | 0.2473 | 0.1951 | 0.2453 | 0.1699 | 0.0449 |
| Latent shift256 | 0.2561 | 0.2174 | 0.2562 | 0.1932 | 0.0450 |

The replacement arm acquires more bits (about 5.4-5.8 versus 3.9-4.8 for
fixed on these post windows), and its selected-view oracle is better, yet its
actual forecast barely improves. Thus observation scope alone cannot explain
the failure. There is substantial predictive headroom even on the views
already acquired, and further headroom between those views and full input.
These are hindsight, finite-sample comparisons, not a decomposition theorem
or evidence that either proposed improvement will work. The first 64 frames
after each shift show the same qualitative gap; complete values are in the
oracle artifacts.
The first oracle outputs checked views only after each shift; that verifier
scope was corrected and rerun to `-oracle-allviews.json` artifacts. Their
numeric records and cells match the earlier outputs exactly. Only the
all-view artifacts support the complete view-consistency claim.

## Audit and cost

The opt-in Go runner verified 983,040 finite pre-outcome forecasts per input
mode, all delivered-origin chronology, audit/availability accounting, and
every arm's Brier/log/accuracy from the full journal. The independent summary
checker verified all paired gains and gate flags. The oracle's selected-view
calculation passed full-mask/prior and latent-correlation unit controls under
the race detector, and a second script recomputed 64 aggregate fields per
mode. `go vet` and the focused test suite passed; a full-package race run was
not attempted for this experiment.

Offline mean per-512-frame stream costs were about 4.23 ms for count fits,
0.146-0.148 ms for tree rebuilds, 356 tree updates and at most 15 current
tree nodes. Core runs took 11.37 and 11.22 seconds per 480 streams, excluding
some artifact serialization. The streaming oracle scans took 3.82 and 3.68
seconds. None of these times measures daemon p95/p99, persistence, or
OpenClaw latency.

This study provides a concrete next hypothesis: test a model that estimates
the conditional outcome law on the *acquired mask/value* without relying on
the current fitted mix, and separately test a task-aware observation policy.
Both must be evaluated on fresh outcome families and untouched agent tasks,
not selected using these consumed journals. The current result is a failed
generalization test, not a rescued algorithm. All seven research directions
remain open.

Artifacts:

- [Uniform comparison check](mmm-retained-truth-v10-uniform-check.json) and
  [latent comparison check](mmm-retained-truth-v10-latent-check.json)
- [Uniform all-view oracle check](mmm-retained-truth-v10-uniform-oracle-allviews-check.json)
  and [latent all-view oracle check](mmm-retained-truth-v10-latent-oracle-allviews-check.json)
- Full journals and summaries share the `mmm-retained-truth-v10-{uniform,latent}`
  prefix; source overlay manifests are under
  `research/retained-truth-v10/{uniform,latent}/`.
