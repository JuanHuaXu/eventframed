# Retained window bank v15

Frozen before evaluation. Preserve v14's count models and Bayesian subset prior.
Three modes, changing only the adaptive-retained arm2:

- subset64: original inner [short-count,subset64,subset64,subset64];
- window_bank: [short-count,subset16,subset64,subset64];
- retained_bank: [short-count,forest64,subset16,subset64].

Use existing ForecastMix priors/share and journaled outcome updates. The winning
inner guide aggregates duplicated identical experts before comparison; ties
prefer the earlier expert. Only if the outer inner-slot wins does it control
acquisition. Other outer arms0,1,3 are identical across modes. Existing incumbent,
long256 model and uniform outer component remain untouched. No destructive reset.

All learners use the same admitted outcomes, with32-first/16-refit cadence.
Subset16 consumes the last16 of that existing audit history; no additional labels
or earlier first fit. Shared empirical input distribution uses last<=256 audited
inputs plus one uniform pseudo-observation. Forest retains its existing uniform
missing-coordinate contract, explicitly distinct from subset joint integration.

Use v14 families and generators/scenarios, six fits and two streams per fit,
two splits, three modes =1728 streams. Fresh fit2026108201, design2026108202,
confirmation2026108203; same collision-free seed encoding. No tuning between
splits. Stream sources and full records as compressed JSON-lines.

For each bank versus subset64 arm2: every family/generator/split full/post mean
harm<=.01; every clustered shift128 post gain>=.005. Also require every shift128
post gain>=.005 versus fixed counts and no >.01 harm versus fixed in any group.
Retain all failures and descriptive six-fitting-group95% bootstrap intervals
(20000 resamples, seed2026108299). Intervals are not simultaneous certificates.

Verify exact subset64 control, unchanged arms0/1/3, paired inputs/feedback,
pre-outcome inner journals, duplicate-weight guide selection, source hashes and
complete replay. This is a research window selector, not ADWIN or an ordinary
Bayesian posterior over nonstationary regimes. No production modification.
