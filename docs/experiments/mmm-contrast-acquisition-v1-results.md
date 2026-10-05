# MMM contrast acquisition v1: faster evidence, full screen fails

The [frozen protocol](mmm-contrast-acquisition-v1-protocol.md) **fails** in
both design and untouched confirmation. A declared baseline-versus-local
contrast policy uses the same 128 label nominations, 256 label requests,
512 live-context reads and 512 reference-context proposals as random and
model-uncertainty controls. It greatly improves paired-gate detection
and early forecast Brier, but its post-shift Brier gain over random is
smaller than the predeclared 0.02. The null-case arrived-label parity
screen also fails in design. Keep this as a promising *conditional*
component, not a Goal 3 or Goal 7 pass.

## Method and audit

The opt-in [Go harness](../../internal/observationgate/contrast_acquisition_v1_test.go)
uses the existing MMM `integrationLabel` generator and a frozen fitted
baseline. The second hypothesis is the generator's bit-2 local rule,
declared to the study in advance. This is an oracle-supplied hypothesis
family, not a discovered EventFrame abstraction. Each policy chooses
online from the current context before the revealing outcome. The
selector function receives context, frozen baseline probability, quota
and its own RNG; it receives no labels, delays or future contexts.

All arms see a shared potential-outcome tape within a trajectory. Four
reference-context proposals and two label requests are charged for every
nomination whether or not the pair matches or arrives. The gate only
receives a complete, same-origin pair in a predeclared contrast class.
The forecast law is a common discounted-log-odds *working* mixture of
the two hypotheses; it is scored on every live clock before due feedback.
The gate does not alter this law, so score gains here come from acquisition
and the common learner, not from an actual Anti-Pigeon split.

- [Design JSONL](mmm-contrast-acquisition-v1-design.jsonl): 5,000 trajectories,
  SHA256 `c1f9bcbb77f691ada76c58d1a80618aec8d2b1458c27c09457d21699687bcf45`.
- [Confirmation JSONL](mmm-contrast-acquisition-v1-confirmation.jsonl):
  5,000 fresh trajectories, SHA256 `5c2b145d7a23340e702c32a81dce32f532dc07a034651f49c83ce1fcd045d35a`.
- The [independent verifier](../../research/contrast-acquisition-v1-verify.mjs)
  checks source hashes, unique rows, exact requested-cost invariants,
  reported cells and paired contrast summaries. A fresh confirmation
  replay matches all 5,002 entries excluding fit and experiment timing.
  The verifier reconstructs summaries from per-trajectory metrics; it
  does **not** independently reconstruct per-clock forecasts or prove
  the hypothesis model.

## Confirmation

| Member shift at 256 | Random | Uncertainty | Contrast |
| --- | ---: | ---: | ---: |
| Paired-gate flags / 1,000 | 639 | 627 | **996** |
| Median flag clock, detected only | 458 | 463 | **391** |
| Post-256 Brier | 0.11231 | 0.11863 | **0.09865** |
| First-64-after-shift Brier | 0.28679 | 0.30274 | **0.24978** |
| Complete contrast pairs / trajectory | 27.12 | 27.17 | **54.00** |
| Arrived labels / trajectory | 199.07 | 200.16 | 199.02 |

Contrast-minus-control post-256 Brier gains are 0.01366 against random
(paired mean-minus-3.5-SE lower bound 0.01094) and 0.01998 against
uncertainty (lower bound 0.01714). Both miss the frozen *magnitude*
floor of 0.02, despite positive lower bounds. First-64 gains are
0.03701 and 0.05296, both above the floor with positive lower bounds.
No member-shift gate flagged before clock 256. These Brier gains are
absolute score differences, not percentages of accuracy.

Design was consistent but also failed: contrast flags 998/1,000 versus
607/1,000 random and 614/1,000 uncertainty, while its post-256 gain
over random is 0.01547. Design's null-case uncertainty comparison has
an arrived-label mean gap 2.261, exceeding the frozen allowance of 2;
it is not erased because confirmation's corresponding gap is smaller.
Every arm still requested exactly the same 256 labels and 512 reference
context proposals per trajectory; the difference comes from selection
times, missingness and horizon-censored arrivals.

Contrast-gate false flags in confirmation are 0/1,000 stable,
0/1,000 common shift and 2/1,000 null, below the frozen ceiling.
The common-shift forecast still improves with contrast acquisition:
both members change, so the pair gate appropriately does **not** split,
while the separate live learner adapts. Recurring contrast flags
940/1,000, but a latched historical flag cannot certify divergence
after the member reverts. Null full-stream Brier is about 0.2734
for contrast; the predeclared screen did not establish a general
nonharm guarantee outside its stable case.

Baseline fits took about 2.3 ms per split; the isolated 5,000-trajectory
experiment took about 0.73 s per split on this host. Those figures
include simulated generation and research bookkeeping, exclude
storage, queues, concurrency and agent serving, and are not a
sub-100-ms request claim.

## Decision

Goal 3 and Goal 7 remain open. The result supports the mechanism that
observing contexts where two declared hypotheses disagree can collect
more discriminating evidence at the same requested acquisition cost.
It does not meet the frozen scored-learning rule, and the alternative
rule was supplied by the experimenter. A successor should test a
predeclared, non-oracle method for generating and validating candidate
hypotheses and improve long-horizon adaptation without changing this
study's thresholds. It must retain random and uncertainty controls,
arrived-label parity, false-split checks, delayed feedback and proper
scores. No production or whitepaper change follows.

This selector is inspired by discrimination among competing hypotheses
in [Golovin, Krause and Ray](https://arxiv.org/abs/1010.3091), but is
not EC2 and inherits none of that paper's guarantees.
