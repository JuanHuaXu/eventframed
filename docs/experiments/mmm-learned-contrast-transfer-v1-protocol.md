# Known-hypothesis acquisition: independent stress transfer

Freeze before collection. This tests whether the existing v2 11-rule
random/uncertainty/learned-disagreement comparison transfers beyond its
original uniform-context, clock-256, 5%-noise generator. No policy, prior,
budget or rule is changed. Design/confirmation seed bases are
2026102501/2026102502; baseline fits are indexed 1200..1215/1300..1315.
For each case/split, run 16 fits x 16 streams, 512 clocks and exactly 128
label requests per policy on a shared potential-outcome tape. Labels may be
used only when selected, nonmissing and due. Outcome truth is evaluator-only.

Seven fixed cases:

| Case | Change | Noise | Missing | Delay | Contexts | Post-change rule |
| --- | ---: | ---: | ---: | ---: | --- | --- |
| early_bit2 | 128 | .05 | .20 | 0 | uniform | bit2 |
| late_bit2 | 384 | .05 | .20 | 0 | uniform | bit2 |
| noisy_bit2 | 256 | .20 | .20 | 0 | uniform | bit2 |
| delayed_bit2 | 256 | .05 | .25 | 32 | uniform | bit2 |
| skew_bit2 | 256 | .05 | .20 | 0 | bit2=1 with probability .10 | bit2 |
| stable20 | none | .20 | .20 | 0 | uniform | no change |
| majority_ood | 256 | .05 | .20 | 0 | uniform | majority(bits0,1,2) |

All other context bits are uniform. Outcomes are Bernoulli with probability
`noise` for false rules and `1-noise` for true rules. The pre-change rule is
parity(bits6,7,8). Each baseline is independently fitted on 4096 pre-change
samples with the matching noise; the learner's known alternatives remain
hardcoded .05/.95, deliberately misspecified for `noisy_bit2`. The skewed
case alters only input frequency, not the law conditional on its context.

Primary measure is expected Brier averaged over clocks from the declared
change to 511, computed from the generator's conditional probability and
recorded forecasts; realized post Brier, first 64 post-change Brier,
arrived-label count, and selected-clock count are also reported. For
5%-noise known cases, recovery is the first 16 consecutive trailing-32
expected-Brier means <= .12, observed no earlier than 31 post-change
clocks; unrecovered trajectories are censored at the remaining horizon.
The 20%-noise case has Bayes risk .16, so this recovery threshold is not
applicable there. OOD and stable cases have no recovery claim.

Use paired fit-cluster intervals `mean +/- 3.5 SE` over 16 independent fits.
The v2 known-family transfer screen passes only if **both** fresh splits meet:

1. In at least four of five known bit2 cases, learned post expected-Brier
   gain versus *each* of random and uncertainty is >= .01 and its fit-cluster
   lower endpoint is positive. In none of the five may the learned post
   expected-Brier harm upper endpoint versus either control exceed .01.
2. Across the four 5%-noise known cases, learned first-64 post expected-Brier
   harm versus either control is <= .005. On `early_bit2`, `late_bit2`, and
   `skew_bit2`, learned restricted mean recovery lead versus both controls
   is >= 10 clocks; on `delayed_bit2`, report the lead and misses without
   requiring an impossible pre-arrival response.
3. On `stable20`, learned full expected-Brier harm upper fit-cluster endpoint
   versus both controls is < .01. On `majority_ood`, learned post expected-
   Brier harm upper fit-cluster endpoint versus both controls is < .01.
4. Every arm requests exactly 128 labels; learned's mean arrived-label gap
   from each control is <= 2 per case. Source hashes, replay, no-future-label
   accounting, and independent per-tick score reconstruction all pass.

Isolated update and forecast/select cost and state size are measured separately;
they cannot establish full service latency. A failed screen is retained, not
retuned on either split. Even a pass would validate only this bounded synthetic
known-hypothesis family, not automatic hypothesis invention, Anti-Pigeon,
agent answer quality, or all of Goal 1/7.
