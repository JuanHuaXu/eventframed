# Learned contrast v3: null rescue works, joint screen fails

The [frozen v3 protocol](mmm-learned-contrast-v3-protocol.md) **fails** on
independent design and confirmation cohorts. Adding a constant-probability
null hypothesis greatly improves absolute null-stream calibration and lowers
error on a shifted rule outside the declared family. It also slows recovery
on known shifts and misses the predeclared 10-clock time-to-recovery gain.
Do not replace the v2 component or promote v3 into serving. All seven whole
research goals remain **OPEN**.

## Confirmation

Each case has 256 paired trajectories from 16 independently fitted baselines,
16 streams per fit. Every arm reads 512 contexts, requests exactly 128 labels,
forecasts before feedback, and shares the same potential-outcome and delivery
tape. The first three arms use v2's 11-rule family; the last three add a
Bernoulli(.5) null rule. Random and baseline-uncertainty arms have identical
selection clocks across model families, giving a direct paired model ablation.

| Case and metric | 11-rule learned | 12-rule random | 12-rule uncertainty | 12-rule learned |
| --- | ---: | ---: | ---: | ---: |
| Null, full512 realized Brier | .37991 | .26384 | .26630 | **.26496** |
| Bit2 shift, post256 realized Brier | **.16599** | .18336 | .18483 | .17197 |
| Delayed bit2, post256 realized Brier | **.20108** | .21594 | .21357 | .20370 |
| Bit0 shift, post256 realized Brier | **.16112** | .18608 | .18110 | .16996 |
| Majority shift, post256 expected Brier | .29737 | .29209 | .29270 | **.28023** |

Null full-stream Brier improves .11494 for the learned arm and clears the
.27 ceiling in all three 12-rule policies, with positive independent-fit
lower gain endpoints. In the out-of-family majority case, 12-rule learned
improves post *expected* Brier by .01715 over 11-rule learned (fit-cluster
mean-minus-3.5-SE lower endpoint .01242). This is a hedge, not successful
rule learning: .28023 remains worse than a constant .5 forecast's expected
.25, and **255/256** confirmation majority streams miss the .12 recovery
threshold (all 256 miss in design). A mixture of the known rules may
approximate majority, but no individual declared rule is majority.

The 12-rule learned policy still beats its own random and uncertainty
controls on known-rule post Brier in several cells, but does not preserve
v2's speed. On immediate bit2, its mean restricted recovery delay is
145.69 clocks, versus 148.34/148.73 for its controls and **130.32** for
11-rule learned. On delayed bit2, it is 168.77 versus 173.91/169.81 and
**156.70** for 11-rule learned. Gains over its own controls are only
2.66/3.04 and 5.14/1.05 clocks, below the frozen 10-clock floor. The
first64 delayed Brier remains near .384 in all 12-rule arms. The 12-rule
bit0 post score also worsens by .00885 versus its paired 11-rule learned
arm in confirmation; the fit-cluster upper harm endpoint .01357 exceeds
the frozen .01 ceiling. Design fails the same known-rule protection.

## Audit and cost

- [Design records](mmm-learned-contrast-v3-design.jsonl.gz): 1,792
  trajectories, SHA256
  `0dcd2eae7c3ed64ee43844e731fba7d07c850f77cdfd7a165e59f4e30ba5fb59`.
- [Confirmation records](mmm-learned-contrast-v3-confirmation.jsonl.gz):
  1,792 trajectories, SHA256
  `856b8fb095aca8b374be471ea6034d630640b3cf12e280c6837c38a839efc2ac`.
- The [independent verifier](../../research/learned-contrast-v3-verify.mjs)
  reconstructs source hashes, the generator's conditional probabilities,
  11,010,048 issued forecasts and scores, all requests/deliveries, expected
  losses, recovery clocks, paired ablation selections and fit-cluster
  comparisons. The fresh confirmation archive replays byte-identically.
  Focused `-race`, `go vet`, and the full observation-gate suite pass.
- State is 240 bytes. Across three 100,000-update timing repetitions on
  this Apple M4, worst isolated rolling-update p99 was 5.375 us;
  forecast-plus-selection p99 was 84 ns in both known-rule and null-fallback
  modes. These are component measurements, not loaded request latency.

The result isolates a real tradeoff: a fixed null branch protects against
overconfident wrong-family forecasts, but its probability mass and fallback
behavior postpone commitment when an in-family new rule is actually present.
That mechanism is consistent with the measured regression; this experiment
does not separately identify the mass versus fallback contribution. Any next
rescue must freeze a new cohort and test that separation explicitly. Neither
v3's synthetic generator nor its evaluator-only true probabilities establish
external-law truth, Anti-Pigeon authority, real-agent usefulness or full
falsification-oriented observation.
