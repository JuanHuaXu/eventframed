# Preserved-incumbent MMM v3 protocol

Frozen 2026-09-12 before evaluating v3. This is a new research test, not a
production change. v1 and v2 evidence remains immutable. No outcome-dependent
parameter changes or selection of a favorable confirmation run are permitted.

## Hypothesis and controls

Preserve the 4096-example original predictor. Maintain short (latest 64 complete
live audits) and long (latest 256 live audits) challengers; pooled long uses the
latest 128 live and 128 reference audits. Train once 32 audit pairs exist, then
every 16 pairs. Shared fitting produces immutable model versions available only
on the NEXT prediction. No labels may affect a current forecast. The original
model never changes. Prior weights on original/short/long/uniform experts are
.7/.1/.1/.1; use existing bayes.ForecastMix with .002 fixed share, probability
floor 1e-6 and unit outcome weight. Do not invoke its reset wrapper. Weights track
adaptive expert algorithms, not immutable-model posterior probabilities.

Arms: frozen_mmm, mix_breadth_ap, mix_random_ap, mix_mmm_no_ap, mix_mmm_local,
mix_mmm_ap. The no-AP arm keeps pooled long; local never pools; AP arms start
pooled and may permanently switch the long slot to local. All other parameters
are equal. After a split, reset only the long slot's influence to the minimum
of its old weight and .1, renormalizing; retain original and short models and
their relative weights. This slot-change rule is frozen, not a model wipe.

One shared foreground observation per arm has cap SIX coordinates, not six per
expert. Its guiding model is the largest previous-weight available non-uniform
expert (ties original first). MMM/breadth/random select views using that model.
All experts then predict using exactly the resulting observed mask/values.
Unobserved values are unavailable. Frozen uses its original MMM guide.
This controls exposure budget when measuring MMM's incremental contribution.

## Sharing gate

Existing observationrescue.Monitor / bayes.AssessRevision nominate divergence
using independent reference/live correctness of the ORIGINAL MMM policy. A
changepoint may request investigation but never erases a model or resets mixture
weights. Only a nominated split plus the following sequential evidence can stop
pooling. Correctness is an observed proxy: this does not certify the full target
transition-law diameter required by the general whitepaper.

Let D_t = reference_correct - live_correct in [-1,1]. For each of eight fixed
starts 0,64,...,448, initialize E+ = E- = 1 and update after paired outcomes:
E+ *= 1 + .25*(D_t-.15); E- *= 1 + .25*(-D_t-.15).
Reject that window if (E+ + E-)/2 >= 800. By the nonnegative-supermartingale
bound and union bound over eight windows, false rejection is at most .01 within
512 steps IF |E[D_t | previous information]| <= .15 throughout. This is an
explicit conditional-mean null, not a universal truth, independence or causal
certificate. The nominal .01 budget is per stream, not across 320 streams.
No estimated selection probability: both correctness streams are fully observed.
Record all nominations, evidence maxima, and actual sharing revocations. A shared
change may be handled by forecast weights without splitting compatible contexts.

## Data and cost

Same five v2 families: stable process XOR; live member shift at 256 to local bit;
common shift of both streams at 256; live recurring changes every 128; null fair
outcomes. Nine binary coordinates, 5% independent label flips except null.
Task views supplied as valid synthetic EventFrame envelopes, not learned text.
32 independent streams x 512 steps x two splits x five scenarios. Fitting seed
2026091701, design base 2026091702, confirmation base 2026091703. Evaluation RNG
seed = split_base*1000000 + scenario*100000 + stream*100 + role, where role=0
live,1 reference,2 audit,3 random-policy. These do not overlap v1/v2 seeds.

Independent Bernoulli(.25) audit nomination before each frame/outcome; every arm
receives the same opportunity. Selected audits read all nine live AND reference
coordinates after forecasting (18 units). Diagnostic old-live/reference MMM reads
are charged separately. All arms pay these common costs, including controls that
do not use them. Report actual foreground + diagnostic + audit cost and fitting
work; a foreground win is not an overall latency claim. Training is synchronous
in this isolated deterministic replay, not a serving implementation.

## Frozen criteria

Primary rescue: mix_mmm_ap post-change Brier gain over frozen_mmm >= .05 with
paired lower bound >0 on BOTH single shifts. Stable protection: full-stream
Brier harm upper bound <= .01 and empirical stable split fraction <= .05.
No overall pass if either fails. Report >.01 mean harm in every scenario.

Distinct questions, not substitutes for primary criteria:
- MMM increment over mix_breadth_ap: positive lower bound on full stable and on
  post member/common shift separately; report random control too.
- AP increment versus mix_mmm_no_ap and mix_mmm_local: report all paired losses;
  no claim of unique AP benefit unless a positive lower bound is observed.
- Report recurring recovery and null behavior, all warm-up, confident errors,
  proper log loss, observed-model support, weight histories, and split timing.

Use z=3.5 paired normal intervals on 32 per-stream metrics, at most 100 prespecified
contrasts (five controls x two windows x five scenarios x two splits). Intervals
are approximate fixed-sample, conditional on the one fitting model, not confidence
sequences. The sharing gate has the separate explicit sequential null above.
Keep all raw traces and source/protocol hashes. Run unit/race and deterministic
replay checks; serial microbenchmarks after experiments, excluding other workloads.

Research basis: https://www.jmlr.org/papers/v3/bousquet02b.html (expert tracking),
https://arxiv.org/abs/2210.01948 (anytime-valid inference). The existing mixer is
log-loss-based; this test measures Brier separately and claims no imported Brier
regret theorem. Fresh-data success remains necessary.
