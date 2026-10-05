# Shared-versus-local freshness: fixed-trace diagnostic

Frozen before quality scoring. Consume ALL v13 traces without changing their
actions. The new model has three fixed components with prior weights(.25,.25,.5):
globally copied renewals, globally independent renewals, and independent .5
freshness modes per type. Each component retains ordinary mode and shared-root
semantics. Components receive their own likelihood update; their posterior
weights update by predictive evidence exactly once per report. An impossible
component receives zero weight without discarding the observation.

This is exact Bayesian model averaging for a declared finite family, not a
certificate that sources are independent. The local component gives mixed
source mechanisms prior support; a correct component is not guaranteed to get
enough evidence within the budget. Unit negative control must demonstrate that
evidence of fresh behavior in one type does not force all other types fresh,
and that mixed evidence can favor the local component.

For all640 episodes and six original action traces, compare hierarchical and
local uncertain inference on identical acquired reports. Primary screen uses
uncertain_mixed traces: each split requires lower paired gain>=-.01 for all five
environments, and mean gain>=.005 plus lower>0 on copied20 and mixed20 genuine
renewal cases. All14 gates required. Intervals mean +/-3.3 SE over64 episodes,
descriptive not simultaneous. Retain all cases, means, intervals and failures.

This pilot contains all-genuine and all-counterfeit renewal environments, not
a generated mixed-counterfeit environment. The structural unit control alone
cannot fill that gap. No quality pass here authorizes pooling, closed-loop
adoption or real-world claims; fresh mixed-counterfeit acquisition tests remain
required even if all fourteen gates pass. No prior tuning on results.

Verify Bayesian evidence chain rule, all six-report binary sequences in two
channel orders, component marginalization and exact full replay. No production,
whitepaper or private-data changes. Forecast costs scale linearly with three
components at fixed type/hypothesis bounds; loaded timing is not measured here.
