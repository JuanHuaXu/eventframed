# Learned acquisition: source check and next distinction

Primary source read: Konyushkova, Sznitman and Fua, NeurIPS2017,
[Learning Active Learning from Data](https://papers.nips.cc/paper/2017/file/8ca8da41fe1ebc8d3ca31dc14f5fc56c-Paper.pdf),
sections4 and5. Their acquisition regressor uses classifier-state and candidate
descriptors to predict error reduction. Their implementation mostly uses random
forests and includes validation performance and learner-state statistics. Their
iterative variant constructs training states under the learned acquisition
policy to represent selection bias. Offline supervision can be expensive.

The general learned-error-reduction idea is prior art, not a novelty of these
EventFrame critics. Our single-decision linear experiments do not reproduce
their regressor, state descriptors or iterative training process; their reported
results cannot be imported as evidence for our failed screens.

## Proposed diagnostic

Our source tape records forecasts before reading the current label
(`soft_learners_v120_test.go`, prediction loop preceding the evaluator-truth
access). Some of those labels have naturally arrived by decision160. This gives
a possible source of prequential error evidence that the current critic lacks.
It must use the issued forecast, never a hindsight refit or source Q.

Before implementing a new critic, build and test a strict projection of earlier
issued forecasts and labels whose delivery is observable by160. Exclude current
frame160 labels and every unresolved outcome. Do not expose actual future delay
or permanent-missingness flags to the learned policy. The projection must
survive poisoning of future/unarrived Y, all Q, and all case/seed metadata.

Then freeze a small factorial comparison: original versus error-informed state
descriptors, crossed with linear versus a bounded nonlinear regressor. This
separates missing information from insufficient functional flexibility. Keep
all controls, query costs, gates and fresh-data requirements. If OOB estimates
or resampling are used, respect whole-trajectory groups rather than treating
sibling candidate rows as independent samples. No thresholds are selected here.

This is an adaptation hypothesis from a source check, not evidence of a rescue.
