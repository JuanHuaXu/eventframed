# Learned contrast v2: bounded known-hypothesis component passes

The [frozen protocol](mmm-learned-contrast-v2-protocol.md) **passes** on
independent design and confirmation streams. Unlike the earlier oracle-supplied
contrast study, the selector names an alternative from an eleven-rule family
using only its own arrived labels. It improves the same rolling-likelihood
forecaster's post-change Brier at exactly 128 requested labels per trajectory,
compared with random and baseline-uncertainty acquisition. This is a bounded
known-hypothesis component result. Goal 7 and all other whole goals remain
**OPEN**; no serving or Anti-Pigeon change follows.

## Confirmation

Lower Brier is better. Each case has 256 paired streams from 16 independent
baseline fits, 16 streams per fit. Forecasts precede requested-label delivery.
Post means cover clocks 256..511. All policies read 512 contexts and request
128 labels per trajectory; missingness and horizon-censored delay affect the
arrived count.

| Case, post256 Brier | Random | Uncertainty | Learned disagreement |
| --- | ---: | ---: | ---: |
| Bit2 shift | .18595 | .18174 | **.16208** |
| Bit2 shift, delay16 + 25% missing | .21945 | .21904 | **.20380** |
| Bit0 shift | .18552 | .18316 | **.16464** |
| Parity(0,1,2) shift | .18422 | .18749 | **.16227** |

The immediate bit2 paired gains are .02387 versus random and .01966 versus
uncertainty; delayed bit2 gains are .01565 and .01524. Mean-minus-3.5-SE
lower endpoints calculated across the **16 fit clusters** remain positive:
.01533/.01075 immediate and .00592/.00795 delayed. These cluster intervals
are descriptive, not simultaneous population coverage. The learned policy
names the correct alternative on 85.60% of its 16,450 postchange bit2
nominations and 74.70% of 16,370 delayed-bit2 nominations. Bit0 is 84.40%
of 16,407. The design split shows the same direction and passes every frozen
gate; its immediate/delayed bit2 gains against random are .02494/.01921.

First64 after the delayed shift is nearly unchanged: confirmation Brier
.38408/.38409/.38370 for random/uncertainty/learned. The result supports
lower average postchange error, **not** a demonstrated early detection-time
advantage. The stable full512 Brier is .06234/.06233/.06229, but the null
full512 Brier is .37605/.37434/**.37976**. Null harm to the controls is
small enough for the frozen relative non-harm gate (upper endpoints .00723
and .00863), yet the absolute score is poor. The model family has no
Bernoulli(1/2) or unknown-rule hypothesis, so this study does not establish
calibration under misspecification. Its null post256 harm versus uncertainty
is .00642, and the fit-cluster lower gain endpoint is -.01019. That warning
must travel with the pass result.

## Audit and cost

- [Design records](mmm-learned-contrast-v2-design.jsonl.gz): 1,536
  trajectories, SHA256
  `7cd22ddb3c66e00b0844958a123b3a894ec27d6e704ca00556cdd9c0eeded288`.
- [Confirmation records](mmm-learned-contrast-v2-confirmation.jsonl.gz):
  1,536 trajectories, SHA256
  `ce9958473834bcdb8cd129f45ea958312d8b0c0310056796508b355f0965c6d8`.
- The [independent verifier](../../research/learned-contrast-v2-verify.mjs)
  checks source hashes, unique trial identities, all 4,718,592 scored
  policy-clock forecasts, requests, no future/unrequested/duplicate delivery,
  exact score and count reconstruction, paired stream/fit-cluster intervals,
  identification denominators and the frozen decision. It reports an overall
  pass for both splits. A fresh confirmation recollection is byte-identical
  to the compressed original. Focused `-race` tests and `go vet` pass.
- On this Apple M4, the isolated 32-label rolling update takes about
  1.44 us with zero allocations; forecast plus selection takes about
  22.5 ns with zero allocations. These are single-threaded component
  benchmarks, excluding baseline fitting, I/O, queuing, persistence, agent
  execution, and loaded tail latency.

The hypothesis family explicitly includes the shifted bit2, bit0 and
parity(0,1,2) rules. The experiment shows evidence-driven selection **among
known alternatives**, not invention of a missing rule or causal discovery.
It uses synthetic nine-bit contexts and paid labels, not agent outcomes or
real-world provenance. There is no reference member, external target-law
diameter, paired-gate certificate or Anti-Pigeon split. A next frozen study
should include an explicit unknown/null branch, non-family shifts and a
predeclared time-to-recovery measure before attempting transfer to Goal 7's
full falsification-oriented observation criterion.
