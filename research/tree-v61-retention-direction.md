# V61: member-specific retention selection

PROPOSAL AND COMPONENT STUDY, not adoption or a completed seven-goal rescue.
V60's full controlled experiment fails quality/recovery. Post-collection probes
show more evidence helps many stationary cases but harms late/recurring changes;
the independent-only branch is worse, so neither change is adopted blindly.

## Related work and patch-reasoning gate

Prior [v97 window bank](../docs/experiments/mmm-window-bank-v97-results.md)
passes 102/106 gates but fails parity-to-majority recovery: global weights
underweight the useful short expert then retain it after its useful period.
[Delayed fixed-share](../docs/experiments/mmm-spike-fixed-share-v1-results.md)
reduces some harms but still has 27 locally harmful windows. No prior theorem
or earlier partial result is treated as a complete rescue.

Primary sources rechecked 2026-10-04:
[Herbster and Warmuth, Tracking the Best Expert (1998)](https://mwarmuth.bitbucket.io/pubs/J39.pdf)
and [Korotin et al., Adaptive Hedging under Delayed Feedback (2019), sections 2 and 4.2](https://arxiv.org/html/1902.10433).
The latter's origin-aware filtering and prior-reset transition motivate this
component. Its setting assumes losses eventually revealed; our missing/noisy,
paired and selective observations do not inherit its regret guarantee.

Classification: V60 accuracy/recovery failure CONFIRMED; retention tradeoff
CONFIRMED as a fixed-stream final-state diagnostic; adaptive selector benefit
NEEDS INVESTIGATION. Possible causes include limited retained samples, stale
evidence, prior/source misspecification and pooling geometry. Probes do not
prove any unique cause. This is isolated research, not an upstream runtime bug
fix; no tracked/production code or earlier frozen inputs are changed.

Falsifiers: a wrong forward/backward result versus explicit latent-path sums,
arrival-order dependence at equal visible evidence, double-counting W1/W2,
retroactive recomputation of issued forecasts, use of unreceived labels,
continued stationary harm, unchanged slow recovery, or excess full-loop cost.

## Bounded candidate

Keep V60's model/priors and unchanged controls. Candidate experts would retain
600/1,200/2,400 global issued positions, not that many labels per member. A
uniform three-expert selector runs independently for each member. Its clock is
that member's original trial ordinal, not response arrival or global packet
count. Before ordinal n>1 use transition

$$T_n(j,k)=(1-1/n)\mathbf 1[j=k]+(1/n)/3.$$

No hidden regime, truth rate or availability enters selection. The neutral
first-trial prior is uniform. Retention lengths come from the bounded diagnostic
and original 2,400-position fixture, not a grid selected to erase failures;
this remains consumed-data development, not fresh statistical confirmation.

## Issued evidence contract

Each expert must issue a clean Y forecast and a NORMALIZED joint forecast
J[n,k](a,b)=P(W1=a,W2=b) BEFORE either measurement arrives. That joint must
integrate ONE Y and that expert's observation-noise law; it must not multiply
two independent Y predictions. Store the joint and clean forecast immutably.
Do not refit those predictions when an old label arrives. The selector is a
working forecast aggregation, not an ordinary posterior over the world's
common noise variable or a new Anti-Pigeon certificate.

At wall time t the original ordinal's factor is 1 if first evidence is absent,
sum_b J[n,k](a,b) after W1=a, or J[n,k](a,b) after requested W2=b. W2 REPLACES
the first factor at the same origin, never appends another trial. Missing replies
add no factor and cannot be silently imputed. Selection depends on already
visible history; no unseen second outcome or source availability is exposed.

Normalize forward messages in ordinal order. When an old factor changes,
recompute that member's bounded suffix from its original ordinal. For a W2
request at an old trial, smooth that origin's expert label using forward/backward
messages from all currently observed factors, then mix each stored joint's
P(W2=1|W1=a). Updating today's weights by an arriving old factor is the wrong
negative control. Issued forecasts themselves are never rewritten.

The component holds at most 200 members and 64 trials/member, with owner,
epoch, monotone-time, request, duplicate/missing and cap fences. Selector ledgers
retain issued grading predictions across base-expert suffix expiry; grading an
old issued forecast DOES NOT restore expired evidence in the base model. These
are distinct lifecycles and must remain distinct in integration.

## Tests and unresolved integration

Before broad collection, implement the standalone Go selector and independent
small-path enumeration. Cover noisy shared-Y joints, delayed/reordered/missing
packets, zero support, cancellation, epochs, invalid joint/clean inputs, cap
atomicity, hidden-future forks, origin versus naive-arrival controls, expert
permutation and repeated request fences. Benchmark issue/replay/request and
constructor allocations with all caps; no same-instance concurrency promise.

Integration must expose actual issue-time joint laws from each base expert,
publish the scored clean mixture, and define acquisition value for that SAME
mixture; merely changing rank or scoring an old observation policy is insufficient.
Measure every expert setup/update, selector replay, nomination and drain. Three
unshared V60 constructors plus selector may exceed 8 MiB: do not hide those costs
or relax that gate. Shared immutable templates are a separate optimization only
after equivalence and lifecycle proofs. Keep .01 gain/harm gates, recovery gates,
400 ms complete-loop cap and all original seven WHOLE success criteria intact.

Component success cannot prove practical benefit. A controlled prospective
candidate run on the consumed diagnostic, independent replay, broader seeds,
error-controlled useful Anti-Pigeon outcomes, untouched agent outcomes, loaded
freshness and equal TOTAL-cost observation comparisons remain necessary.
