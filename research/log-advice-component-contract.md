# Log-advice component contract

Research only. Retain all frozen v108 Brier/age files unchanged. The log adapter
embeds the verified no-age journal and replaces only selector delivery. The
original carrier bank still supplies journal lifecycle bookkeeping; its internal
Brier updates do not choose this policy's advice weights.

For captured issue-time raw probabilities p_j in(0,1), append neutral p_0=.5.
Use the same optional .05 neutral prior and original raw prior proportions.
At each validated arrival, add the Bernoulli log likelihood to log weights,
normalize, and mix .001 of the declared prior. No age discount, adaptive rate,
new threshold, private data or change-point oracle is used. Acquisition and
issued forecasts share the same evidence-routed mixture.

The immediate ungated interpretation is checked by enumerating finite paths
with transition T(next|current)=(1-share)I+share*pi. The recurrence matches
the next-state posterior after each observation and transition. This numerical
check is not a theorem about delayed filtering. At nonzero share, arrival order
changes the transitions' order relative to events; this is role-advice weighting,
not exact event-clock inference under an HMM. Gate/censoring guarantees do not
transfer automatically.

Tests cover nonuniform priors, neutral enabled/disabled, zero/.001/.2 share,
literal path enumeration, zero-share arrival permutation, strict support,
invalid input atomicity, origin/version/censoring, duplicate and reentrant
delivery, and failed acquisition. At zero share, log weights are retained even
when exp underflows: avoid irreversible log(0) feedback from temporary displayed
zero weights. The deployed research variant always uses .001 share.

The prior Brier regret statement is not claimed for this selector. Evaluate
expected Brier as the primary quality criterion and log score only secondarily.
The raw selector's model-path interpretation does not prove a regret guarantee
for the evidence-gated served mixture.
