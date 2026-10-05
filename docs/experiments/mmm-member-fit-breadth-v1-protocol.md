# Member fitting-sample breadth v1

Question: do the member speed and finite false-revocation results survive
independent base fitting samples? This is not a proposed quality rescue.
Preserve every gate from member-coverage-v1, including the failed >=.005
post-Brier improvement gate. No algorithm or threshold tuning is permitted.

Use512 independent trajectories per scenario per phase, five scenarios and two
phases (5120 total). Phase seed bases2026091511 and2026091512 were searched in
source/protocols before creation with no matches. Keep memberRun unchanged.
For each trajectory independently fit4096 samples with its original generating
law and observation.Fit. The fit RNG uses observationpreserved.Seed with role4;
live/reference/audit/random streams retain roles0-3. Record fit seed and sample
SHA256 with each original record. No shared fitted base across trajectories.
Before collection, exact legacy-seed parity must match observationpreserved.Base
for both null and non-null laws. Distinct fit samples must have distinct hashes.

The same five paired arms, six-coordinate foreground cap, audit budget,512-step
horizon, fitting cadence and scoring windows remain unchanged. All twelve
stable/common/null false-revocation bounds use exact Clopper-Pearson alpha=.05/12
and must be <=.02. These now bound the joint fitted-sample/stream population of
the stated generators, not every possible fixed fitted model. They remain
finite-horizon bounds, not anytime target-law certificates. Paired intervals
remain mean +/-3.5SE. Member delay must improve>=10% without more premature
splits; post-Brier gain>=.005 with positive lower bound; other full/post gains
must have lower bounds>=-.01. Retain the conjunction, not just passing parts.

Require focused race contracts, immutable source hashes, complete deterministic
replay, summary replay and independent scoring audit. This experiment changes
no production code, sources of actual evidence, dependencies or model endpoint.
Training and trajectory work both count toward total experiment wall time;
neither is a serving-latency claim. No whitepaper edits, commits or pushes.
