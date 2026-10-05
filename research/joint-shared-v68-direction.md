# V68 Lead: Coherent Shared Context With Non-Sharing Controls

PROVISIONAL. V66 is coherent and cheap but scientifically fails; V67 shows
removing the reset is not sufficient. Do not adopt either failed learner, tune
another hazard on these labels or call the next reused cohort confirmation.

## Research Grounding

[Adams and MacKay](https://arxiv.org/html/0710.3742v1), sections 2/2.1, explicitly
integrate a predictive law over change/run-length uncertainty; a constant hazard
corresponds to a geometric gap prior. That supports joint sequence accounting,
not the empirical suitability of our 1/16 prior or a guarantee of fast recovery.

[Bonilla, Chai and Williams](https://proceedings.neurips.cc/paper/2007/file/66368270ffd51418ec58bd793f2d9b1b-Paper.pdf),
sections 2/3, induce cross-task covariance and discuss low-rank approximations.
They also flag negative transfer and a noiseless-block-design transfer
cancellation. We borrow the question of controlled cross-context dependence,
NOT their Gaussian observation assumptions, complexity or guarantees. EventFrame
has binary outcomes, paired same-Y measurements and delayed selective receipts.

## Small Mathematical Preflight

`joint-shared-v68-preflight.mjs` instantiates four members and 30 joint states:
ten finite shared calibration atoms x three STATIC shared-noise states. One
atom is the supplied baseline (.8 prior). Three simple context loadings use
states {-4,0,4}, weights {.3,.4,.3}, total family masses {.08,.06,.06}. Each
loading/member has an intercept solved BEFORE observations so its weighted
logistic mean equals the supplied baseline. Hence initial forecasts preserve
every baseline while evidence can affect other members through a shared state.
This is an original finite toy adaptation, not a full GP implementation.

Within the toy episode the calibration state is fixed. The future target takes
one 1/16 reset transition, keeping noise fixed; it is the mean Brier risk over
four NEW Y outcomes. W2 still measures the original Y at its nominated member.
All 16 first-observation patterns x four possible queries are independently
enumerated over the 30 states AND 16 original Y vectors. 5,285 scalar checks
pass; max difference 6.66e-16. Joint evidence, baseline means and all-target tower
identities agree. Maximum nonlocal prediction value .001131853 is nonzero.

Therefore V66's local-only `/member_count` prediction-value shortcut CANNOT be
carried unchanged into a shared-context model. Compute the law change for every
affected target and account for that computation. This preflight proves only
finite algebra; no empirical quality, Anti-Pigeon certificate, runtime, online
global transitions or member-local escape implementation is yet validated.

## Next Implementation Tests

1. Define a bounded joint sequence model for shared contextual calibration and
   paired evidence, with an explicit independent-member alternative. No fitted
   future labels, retrospective similarity authority or oracle geometry.
2. Preserve baseline means and full-rate support. Enumerate short dynamic paths
   against an independent reference, including delayed original-position factor
   replacement, all-target acquisition values, atomic faults and future forks.
3. Measure the actual computational price of sharing: global/member clocks,
   affected target sets, prefix/suffix recomputation, request delay and memory at
   150 AND 200 members. Do not silently shrink the workload to obtain a pass.
4. Run existing complete consumed-cohort quality and recovery gates, including
   non-sharing/shared-noise/shared-rate ablations, static/misspecified-noise and
   unrelated-member negative controls. Preserve failed variants. No equal-total-
   cost observation win from equal request counts alone.
5. Only after diagnostic survival, freeze independent generator/noise/shift/
   delay replication and matched total-cost observers before reserved confirmation.
   Anti-Pigeon remains external final authority for actual posterior sharing;
   this modeled dependence is a challenger hypothesis, not proof of safe merges.

## Whole Goal Boundary

1/2/4/7 still need external quality, recovery, sample-efficiency and equal-total-
cost evidence. 3 still needs useful downstream certified splits; 5 untouched
labeled agent utility; 6 loaded durable freshness/serving. These are not waived
or made conditional on choosing an easier synthetic target. They can progress
independently; a consistent toy is not a completed goal. Production, private
sessions, sealed task labels, whitepaper and publication remain untouched.
