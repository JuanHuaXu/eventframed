# v97 short-window expert bank (frozen before outcomes)

Retain the v96 six arms and all106 primary gates. Add generic32, Boolean32,
and the primary four-expert Brier bank; nine reported arms in total. The bank
experts are generic64,Boolean64,generic32,Boolean32 with prior
[.95,.05/3,.05/3,.05/3], eta.5 and rho.001. Fixed generic prior retains the
comparator budget; other experts share the remaining mass symmetrically.
Replacing the three challenger forecasts by identical Boolean64 forecasts
must reproduce the two-expert predecessor, tested independently.

Only new model windows change. Preserve the fixed32-step fit cadence,
full-frame evidence and the separate full/mask63 weight states. The first
fit has16 samples for both windows; thereafter use at most32 and64 previously
received samples respectively. Do not reset weights on a fit or use the
simulator's change time. Record fit steps, sample counts, oldest/newest origin
indices and pre-outcome bank weights for every publication and view.

Two phases,32 paired streams each of the same12 cases,16 initial labels,
256 scored steps:768 streams. Fresh seed = 2054119700 + phase*1000000 +
case*10000 + index*10, roles0/1/2 rule/input/outcome. v93 disjoint phase rule
pools and case generators are reused, never their outcome streams.

All forecasts exist before sampling the current outcome. Oracle probabilities
are evaluation-only. Include all adaptation costs in full-stream and late-half
expected Brier/accuracy, and record realized loss. Generic64 remains comparator.
Bank prefix bound uses its actual normalized generic prior and sharing rate:
[-log(prior0)-(N-1)log(1-rho)]/.5, tolerance1e-8. This controls realized complete
feedback loss, not pointwise future risk, delayed feedback or a change interval.

Primary bank must pass all106 frozen v96 quality gates:48 full-stream and48
late non-harm upper bounds<=.01,6 stationary interaction gains and4 recovery
gains with mean>=.005 and lower>0, using paired32-stream z=3.5 normal intervals.
These are approximate screening intervals, not anytime confidence sequences.
No hyperparameter sweep, no choosing another arm as winner after seeing data.

Unit-check explicit normalization, singleton/equal and duplicated experts,
the two-expert predecessor, lifecycle and comparator bound. Run race smoke,
vet, exact full replay and independent hash/summary verification. Benchmark
all new fitting costs and the bank arithmetic separately. No production change.
