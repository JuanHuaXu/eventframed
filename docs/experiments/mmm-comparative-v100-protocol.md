# v100 comparative forecast falsification

Frozen before learned-stream and family-null outcomes. Component results are
already consumed; they establish numerical checks and cost, not quality.

## Matched learned streams

Retain all ten v99 arms and append the comparative-gated bank as primary arm10.
Fresh seed2066110000+phase*1000000+case*10000+index*10; roles0,1,2 select rules,
inputs and outcomes. Two phases,12 cases,32 trajectories each,16 initial labels,
256 scored steps, refit before steps0,32,...,224 on latest64 and32 labels.
Both full511 and partial63 views train on complete frames. Same models, model
priors, uniform compiled input law, raw-bank priors and sharing as v99. Partial
view expected Brier is scored against full simulator truth, not used in fitting.
All forecasts precede current outcome generation; feedback is immediate/full.
Independent rule pools across phases do not turn these into real-world evidence.

All106 quality gates remain unchanged:48 whole-stream and48 late-half non-harm
checks require an approximate paired z=3.5 upper bound on Brier harm <=0.01;
six full-view stationary interaction gains and four full-view late-half shift
gains require positive paired lower bounds and mean gain >=0.005. There are32
independent trajectories per cell. These approximate intervals are not anytime
confidence sequences. Report both phases, all gates and all eleven arms; never
select a winner after seeing outcomes. Overall PASS requires all106 gates and
all four finite null checks below. This narrow PASS cannot complete the roadmap.

## Comparative gate

Each target expert uses a fixed mixture: neutral alternative mass1/2 and each
other raw expert mass1/6. All forecasts have strict full support and exist before
the outcome. For each alternative, mix likelihood-ratio products over all32
predeclared starts, dormant terms remaining1. No maximum over alternatives.
Use the same threshold6400,64 tests, eight scheduled versions, four experts and
two views, with conditional family error budget0.01. Unrejected does not mean
true. Other experts need not be correct or independent for an alternative to
be a normalized predictable bet. Its evidence weight is not an extra vote.

Reject only the tested published version; mask from NEXT forecast, renormalize
the unchanged raw-bank weights, fall back to0.5 if all mass is removed. Reopen
only on the scheduled next refit. Retain rejected raw forecasts as alternatives.
Both gated policies observe all outcomes; neither changes training or raw-bank
weight updates. Record each first rejection and reconstruct masking/fallback
counts from its timing. Raw-bank cumulative-loss bounds do not cover either
gated law. Neutral dilution costs log2 evidence; no assumption of improved power.

## Paired family-null simulations

4096 families per mode. Seed3090110000+mode*10000000+family*1000+test.
Mode0:64 independent target-null streams,32 outcomes each. Conditional p cycles
0.05,0.2,0.5,0.8,0.95 using(test+step)%5. Alternatives are0.5, an adaptive0.2
(0.8 if previous outcome true; previous initially false),1-p,and p.
All alternatives are chosen before sampling Y from Bernoulli(p).
Mode1:eight independent version streams, each logically copied to all four
experts and both views. All target laws then equal p, so alternatives are
0.5,p,p,p; the64 tests are perfectly dependent within each version.

Run neutral and comparative monitors on the same outcomes. Count families with
any rejection for each method/mode. Require the95% Wilson upper rejection bound
<=0.01 for all four counts. This checks a known calibrated null only; learned
model rejections cannot be labeled false positives using this simulation.
No seed, boundary, mass or threshold sweep after generation.

## Verification and cost

Freeze19 transitive source/protocol hashes; reproduce the complete learned and
null artifact, validate the summary independently, run race reference/lifecycle
and learned smoke checks, and vet. Preserve failed artifacts. Measure the whole
eleven-arm fitting fixture separately from the already measured monitor kernel;
neither is a production serving benchmark. No production, model installation,
private-data access or external publication is part of this experiment.
