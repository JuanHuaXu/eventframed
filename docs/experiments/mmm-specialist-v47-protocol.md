# V47 task-local specialist mixture

Frozen before candidate quality output. Research only; no production adoption.
All seven whole goals remain OPEN. V43's failure is confirmed; its attribution
to global weight coupling is a hypothesis, not a proven bug. Alternatives are
inadequate child experts and too little evidence for rapid local recovery.

## Source and adaptation

[Freund et al. (1997), Using and combining predictors that specialize](https://cseweb.ucsd.edu/~yfreund/papers/SpecializedExperts.pdf),
Figure 1, updates awake experts and preserves their total weight; asleep
weights remain unchanged. With disjoint observable-task blocks, normalized
SBayes weights are ordinary Bayes within each block. This is V47's alpha=0
component. Positive reset hazards and delayed original-position replay are
explicit adaptations; the source's log-loss comparison is NOT a Brier,
stationarity, delayed-feedback, or Anti-Pigeon guarantee.

[Herbster et al. (2020), Online Multitask Learning with Long-Term Memory](https://arxiv.org/abs/2008.07055)
Section 2.1 supplies the observed task-index/local-clock formulation. V47 does
not implement or inherit its multitask-memory algorithm or regret bound.

For observable member m, retain its nomination sequence j=1,...,16 and
original three-expert advice q[m,j,h]. On this LOCAL clock only, use
T[g,h]=(1-alpha)1[g=h]+alpha*pi[h], pi=(.8,.1,.1). The first position
starts at pi. Subsequent positions transition once, including canceled and
not-yet-revealed positions. Revealed Bernoulli emissions use original advice;
unknown/canceled emissions are one. Delayed evidence re-filters that task in
original nomination order. Other tasks do NOT advance its mixer clock.
Actual scored prediction is sum_h predictive_weight[m,h]*current_advice[m,h].
Full, Adaptive, and rich-moment2 child learners remain shared ALL-visible-data
learners. Their advice must match separately audited control tapes exactly.
No regime identity, true rate, unseen outcome, or future nomination enters
candidate construction, routing, prediction or update.

## Fixed study

Three prospectively fixed hazards: static=0, local=1/16, reset=1. No output
selects/changes prior, hazard, routing or gates. Global cap2400, pending2400,
150 observable members, local cap16; original child caps64 stay unchanged.
The alpha=1 prior-only mixture is a negative control for evidence benefit.
Constructor allocation must stay <=8MiB; whole collected loop <=400ms.
The entire constructor, issue/resolve/inspection/snapshot/scheduling loop is
timed; independent audit and fixture generation are separate.

Fresh seed bases diagnostic2026104707 (n1/cell), design2026104709 and
confirmation2026104711 (n16/cell). Check ACTUAL world-seed disjointness against
V39/V40/V41/V43 and all three new splits. Reuse unchanged V39/V41 generators,
both geometries, all14 regimes, all3 delays: original84 cells, 28diagnostic
worlds or448worlds per normal split. Diagnostic has no intervals or adoption.
Preserve the original V43 quality screens: paired n16 mean+/-3.5SE; stationary
and final loss/usefulness lower gain >=-.01; shifted issued Brier/priority
versus Full mean gain>=.01 and lower>0; all Adaptive protection>=-.01;
declared phase recovery lower gain>0 and mean>=10% Full. One SAME fixed policy
must pass EVERY cell in BOTH cohorts. No cross-cohort policy substitution,
retuning, resampling or weakened thresholds. These screens are exploratory,
not simultaneous certificates, agent-task results or loaded latency proof.

Independent dense matrix reference must visit all ORIGINAL task positions,
bind each forecast/advice/receipt/snapshot and reconstruct per-member weights.
Audit complete arrival sets, cost conservation, caps, exact source hashes and
world coverage. Fault controls: foreign/old/duplicate/backward tickets,
partial-child fence, canceled unit emissions, inactive-task no-aging,
global/local cap rejection, future-prefix invariance, and semantic corrupted
advice/forecast/identity/weights/cost/coverage rejection. No candidate cache
may be used by the reference. Preserve negative raw results and exact sources.

Scope is an isolated research candidate, not repair of an invalid production
record. The falsifier is local weighting failing unchanged broad gates despite
valid reconstruction; this rejects the rescue, not the audit requirements.
