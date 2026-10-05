# Learned contrast v3: null and out-of-family transfer protocol

This isolated Goal 1/7 study follows the fresh v2 known-hypothesis component.
It does not retune v2's consumed streams. The new question is whether an
explicit unknown/null branch improves absolute calibration without erasing
the earlier equal-cost acquisition gain, and whether the gain reflects
earlier *recovery*, not merely a lower postchange average.

## Frozen population and arms

- Seven 512-tick cases: stable parity(6,7,8); parity-to-bit2; the same shift
  with 16-tick label delay and 25% missingness; parity-to-bit0;
  parity-to-parity(0,1,2); parity-to-majority(0,1,2), which is outside the
  declared known-rule family; and a Bernoulli(1/2) null. Other cases have
  20% independent missingness and zero delay. All nonnull rules have .05
  independent flip noise. Change onset is clock256.
- Use new design/confirmation stream seed bases 2026102301/2026102302 and
  separate frozen-baseline fit indices 600..615/700..715. Per split and case,
  run 16 independently fitted baselines times 16 independent streams. Within
  each trajectory, all six arms share context, outcome, missingness and delay.
- The first three arms preserve v2's 11-rule, 32-arrived-label working
  likelihood and random/uncertainty/learned-disagreement selectors. The next
  three use the same policies with a 12th constant-Bernoulli(.5) rule. Frozen
  priors are .85 baseline, .01 for each of ten known alternatives and .05
  null. If the null weight exceeds .5, the new learned selector falls back
  to quota-rate random nomination; otherwise it nominates
  current contexts that disagree with its largest-weight known alternative.
- Each arm reads 512 contexts, nominates exactly 128, requests one label per
  nomination and updates only from labels delivered after its current
  forecast. Horizon-censored and missing labels are recorded separately.
  The null branch is a *working* likelihood term, not an external truth
  certificate. Neither arm can invent the majority rule.

## Measures and frozen decision

Report realized Brier on full512, first64 after onset and post256, plus
expected Brier against the declared generator's true conditional probability.
The latter is an evaluator-only diagnostic, never available to any selector.
For each changed case, define recovery as the first clock after onset at
which trailing-32 expected Brier is at most .12 for 16 consecutive clocks.
Unrecovered streams receive delay256. Compare restricted mean delay and miss
fraction. Use paired means and mean +/- 3.5 standard errors at both the 256
stream and 16 independent-fit-cluster levels; neither is a simultaneous
population certificate.

The new 12-rule *learned-disagreement* arm passes this component screen only
if **both** design and confirmation meet all conditions:

1. On the null case, each 12-rule policy's full realized Brier is at most
   .27 and improves its matched 11-rule policy by at least .05, with a
   positive fit-cluster lower gain endpoint. On stable parity, the 12-rule
   learned arm's full harm versus the 11-rule learned arm has an upper
   fit-cluster endpoint below .01.
2. On immediate and delayed bit2 shifts, 12-rule learned post Brier gains
   over both 12-rule random and uncertainty are at least .01 with positive
   fit-cluster lower endpoints; first64 mean harm is at most .005. Its
   post-Brier harm versus the 11-rule learned arm has an upper fit-cluster
   endpoint below .01 on all four known-rule shifts.
3. For both bit2 cases, 12-rule learned restricted mean recovery delay is at
   least 10 clocks shorter than **each** 12-rule control, and its miss
   fraction exceeds neither control by more than .05. The majority case is
   an explicit out-of-family falsifier: report its score and recovery, and
   require 12-rule learned post *expected* Brier gain over 11-rule learned
   of at least .005 with a positive fit-cluster lower endpoint.
4. All arms request exactly128 labels; 12-rule learned arrived-label mean
   gaps against its controls are at most two per case. No invalid state,
   future-label use or replay mismatch. Isolated rolling update p99 is under
   10 us, forecast plus selection p99 under 1 us, and state under 4 KiB.

The known-rule alternatives are declared, not discovered. A pass is only a
synthetic component result; no Anti-Pigeon authority, external target-law
bound, loaded request latency or agent-outcome claim follows. A failure is
retained without retuning these cohorts.
