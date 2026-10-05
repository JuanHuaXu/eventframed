# Same-window family evidence candidate

Frozen before efficacy testing. This reuses the generic/Boolean evidence model
already present in segment_posterior.go, not a new inference invention. Earlier
full-segmentation and static-atom rescues failed broad screens; retain them.

Expanded prior-work check before any new efficacy run: family.go and
family_prior.go already implement this same-window model, including the
skeptical prior. V91/V92 static-data failures and V93's mixed evidence remain
valid. This experiment is a streaming/cadence re-evaluation, not a novel model.
Reuse familyLogEvidence and check predictions against fitPriorConditional;
the research adapter merely records evidence and avoids constructing unused
partial-view tables. Do not describe the existing implementation as a gap.
V92 used generic prior0.9; this protocol had already frozen0.95 before that
historical comparison. Keep that distinction explicit and do not retune it
from the upcoming outcomes. A consumed-data pass would not erase V92's failure.

Given one eligible ordered sample set S, compute family marginal likelihoods
Z_G(S), Z_B(S) under the existing normalized subset prior and Beta(1/2,1/2)
rates. Freeze prior generic mass pi=.95. Posterior generic mass is
w_G=pi Z_G / (pi Z_G + (1-pi) Z_B), computed in log space. Predict using
w_G p_G(y|x,S)+(1-w_G) p_B(y|x,S). Both likelihoods and both predictors use
EXACTLY S. Comparing Z values across32 versus64 samples is prohibited.

This is ordinary model averaging conditional on a declared window/model family,
not a calibrated truth certificate or posterior over the whole changing stream.
Overlapping refits rebuild the window posterior rather than multiplying past
window likelihoods again. No changepoint partition, old issued-performance
journal, current outcome or hidden Q enters the family weights.

First stage implements a research-only fitter and tests it against independent
beta integrals and the existing segment likelihood's no-boundary component.
Check prior endpoints, invalid inputs, predictive ratios, evidence identity,
sample-order invariance, input ownership and fit/aggregation cost.

After component validation, collect four variants:64/32-sample windows at
32/8-frame fit cadences, using the same as-of origins already frozen in v120
and the cadence diagnostic. Primary is64 samples/cadence32. Compare each with
the same-cadence served Markov control across all2688 consumed trajectories.
No paid labels. Keep all variants and all forecast components; do not relabel
a favorable secondary variant as the original primary.

Primary screens: all168 served Brier cells have paired gain lower>=-.01;
delayed terminal cases1,2,4,5,7,8,19,20 in both phases need mean gain>=.005
and positive lower bound. Intervals are mean +/-3.5SE across32 trajectories,
exploratory and not simultaneous/anytime coverage. All seven full goals remain
open even on a consumed-data pass pending required replication/integration.

The primary retains the original publication cadence and requires no extra
expert fits if it replaces the corresponding two-family window forecast.
Marginal aggregation is extra O(512) arithmetic on fitted evidence and tables;
measure it. Rebuilding models is NOT O(1); do not confuse serving a frozen
table lookup with training cost. The faster8-frame secondary still costs more
fits. No production defaults or paper claims change in this stage.
