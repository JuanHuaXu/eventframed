# v101 evidence-attributed routing

Frozen before fresh learned-stream and null outcomes. Component tests are already
consumed; no quality results were used to choose the routing coefficients.

## Paired design and success criteria

Retain all eleven v100 arms and add evidence routing as primary arm11. Fresh
learned seed2078110100+phase*1000000+case*10000+index*10, roles0/1/2 for rules,
inputs and outcomes. Two phases,12 cases,32 trajectories each;16 initial labels,
256 scored outcomes, full511 and partial63 views, complete training frames,
32-step refits on latest64/32 labels. Retain v100's model/prior/weight settings
and immediate full feedback. Predictions precede current outcome generation.

Keep all106 gates:48 whole-stream and48 late-half non-harm checks have approximate
paired z=3.5 upper Brier-harm bound <=0.01; six full-view whole-stream stationary
interaction gains and four full-view late-half shift gains need mean >=0.005
and positive paired lower bound. There are32 independent trajectories per cell.
No anytime/exact interval coverage is claimed. Overall PASS requires all106 gates
and all four null screens below. No retrospective winner selection or threshold
tuning. A finite PASS does not complete the seven-direction roadmap.

## Routing, not new rejection evidence

Use exactly v100's comparative tests: half neutral alternative mass, one sixth
for each other raw expert,32 equal starts,6400 boundary,64 tests and conditional
0.01 family error budget. At the first crossing for target i, freeze log-normalized
alternative contributions including dormant starts. At subsequent predictions,
redistribute current raw-bank weight of each rejected target to its neutral and
currently unrejected alternative recipients, proportional to those frozen credits.
Keep the original raw weight of every unrejected expert. Normalize final weights
for floating-point summation. Non-finite or unnormalizable credit falls to neutral.

Artifact effective-weight index0 is neutrality,1..4 are raw expert indices0..3.
No mass is routed recursively or to a rejected recipient. Credits are neither
truth probabilities nor future-Brier guarantees. All-rejected outputs0.5. Clear
credits only at the same scheduled version reset. The revealing outcome cannot
change its already-issued prediction. Ordinary raw-bank learning is unchanged.
Require the routed gate's complete comparative-test state to equal the paired
comparative control after EVERY outcome, not just aggregate rejection counts.

## Diagnostics, nulls and cost

Record all arms' block Brier/accuracy over each32-step version, plus block means
of raw and effective weights and neutral-fallback counts. Reconstruct whole/late
metrics independently from block summaries. These diagnostics explain results;
they do not feed training, determine success or change publication timing.

Repeat v100's paired neutral/comparative null screens with seed3110110100+
mode*10000000+family*1000+test.4096 families per mode:64 independent target-null
tests or eight shared version streams copied to four experts and two views.
Each32-step stream samples Bernoulli(p), cycling p=.05,.2,.5,.8,.95 by(test+step)%5.
Independent-mode alternatives are.5, previous-outcome predictor(.8 if true else
.2, initially false),1-p,p. Shared-mode alternatives are.5,p,p,p. All are issued
before the current outcome. Require all four95% Wilson upper rejection bounds
<=0.01. Routing changes no tests and receives no extra null budget. This is not
a calibrated-law guarantee for learned models or a routed-loss bound.

Freeze21 source/protocol hashes before generation. Require complete artifact
replay, independent summary reconstruction, race component/learned smoke checks,
vet and new-file whitespace checks. Measure capped K=4 O(K^2) routing separately
from the full twelve-arm fitting fixture. Neither measures serving latency.
No production changes, private-data acquisition, installs or external publication.
