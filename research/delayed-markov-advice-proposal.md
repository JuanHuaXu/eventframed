# Delayed Markov advice: prospective research after v111

Subsequent [v112 quality test](../docs/experiments/mmm-markov-v112-results.md)
passes protection but FAILS the complete gain screen. Preserve the original
proposal and component evidence below; neither supplies the missing gains.

Subsequently implemented with [component checks and paired benchmarks](delayed-markov-component-results.md).
The original prospective formulation follows.

V111 rejects compatibility
handoff as a complete rescue. Its pending-loss attenuation worsens both delayed
switches against publication-only transfer in both phases. Preserve those
results; do not tune its affinity exponent or weaken its protection gates.

## Different mechanism, not another discount factor

The existing arrival-log policy scores an immutable old forecast against the
current advice weights and applies a share transition when feedback arrives.
That is its declared policy, not a coding error. It is not the same operation
as conditioning a Markov sequence of advice states on evidence about a past
state. Handoff by geometric attenuation does not restore this distinction.

[Mourtada and Maillard (2017)](https://proceedings.mlr.press/v76/mourtada17a/mourtada17a.pdf),
Algorithm2 and Lemma4 in Section5.1, express MarkovHedge as exponential weighting
over state sequences, with likelihood reweighting followed by a stochastic
transition. The forward recursion avoids enumerating every path. Their protocol
observes each outcome before advancing; it does not provide our delayed,
censored, gated Brier guarantee. The proposed delayed message handling below
must be checked independently against explicit path enumeration.

Use the existing four fixed learner roles and original prior, not a growing
model bank. A role is a sequential prediction policy and may legitimately
publish a changed fitted model. Its issued forecast remains immutable.
This avoids adding a new expert population or inheriting old model-specific
rejection tests. The ordinary version-scoped comparative gate stays separate.

## Exact finite-state calculation to test

Let z_i be the latent advice role for issued event i and let p_i,j be that
role's actual issue-time probability at the acquired view. Predeclare:

```text
T(k | j) = (1-alpha) * 1[k=j] + alpha * prior_k
alpha = .001

emission_i(j) = p_i,j^y_i (1-p_i,j)^(1-y_i)  if label i is available
                1                          otherwise

posterior_i = normalize(prior_i * emission_i)
prior_(i+1) = T * posterior_i
```

Alpha deliberately starts at the existing .001 value; do not fit a forgetting
rate to consumed outcomes. Transitions occur once per issued event, not once
per label arrival. On complete immediate feedback this should recover the
existing log/share forecast sequence. With missing or delayed labels it is a
different, explicitly specified filter. Marginalize an unknown label with
emission1; never substitute a negative outcome or a .5-valued observed label.

When a delayed label arrives, replace its previously unit emission with its
immutable issue-time likelihood, then recompute the forward messages from that
origin through the current issued frontier. An old label changes beliefs about
the old state; the declared transitions carry that information to the present.
It is neither a score under the current fitted model nor a direct likelihood
penalty on today's role weights. Retain every label's ownership and single use.

Conditioning on the issued observation stream is part of this research model.
The common input distribution and frozen observation policy do not establish
correctness under arbitrary selective missingness or misspecified evidence.
This is an exact calculation for a declared finite model, not proof that its
latent states describe the real environment or that calibration improves.

## Bounded journal integration

Keep a committed forward checkpoint and the uncommitted issue suffix. In the
current frozen delay/expiry contract, unresolved feedback lives for at most32
event clocks and the ring capacity is64. A prefix may be committed only once
each included label is settled or explicitly censored under that contract.
No later delivery may alter a committed prefix. Already delivered labels remain
in the suffix until they can be committed; recomputing a message is not a second
admission of their evidence.

Track the issued-event frontier separately from wall-clock flush time. Do not
create hidden-state transitions for extra events that were never issued.
Out-of-order feedback, same-clock zero-delay feedback, expiry and publication
must have explicit ordering. The published comparative tests still consume only
their existing ordered, version-valid evidence. Refiltering advice must not
repeat a gate update or alter an already emitted forecast.

The structured transition takes O(M) operations rather than a dense O(M^2)
multiply. Refiltering D retained issues costs O(D*M), with M=4 and D bounded by
the explicit journal contract here. Memory stores the immutable issue forecasts,
availability flags and checkpoint. This added cost must be measured; it cannot
be hidden in the previous arrival-log O(M) update claim. No persistence change
or foreground production integration is authorized by this proposal.

## Required falsifiers and next steps

1. Prove implementation equality with literal enumeration of short state paths,
   including out-of-order and missing labels. Compare final conditional beliefs
   for different arrival orders with the same available evidence; preserve the
   actual issued probabilities when making that comparison.
2. Test immediate log/share forecast equality, zero-transition commutation,
   fully mixing transition limits, and evidence crossing model publications.
   Do not require intermediate states to match the arrival-order heuristic
   when the two algorithms intentionally implement different clocks.
3. Audit checkpoint expiry, duplicates, labels on either side of a publication,
   backpressure, reentrancy and atomic rollback. Check coherent acquisition and
   serving under the same gated mixture. Archived emitted scores never change.
4. Benchmark the complete bounded lifecycle against original arrival-log,
   including worst allowed reordered feedback. A unit-test pass is not quality.
5. Freeze a fresh complete quality comparison with the existing stationary,
   recovery and delayed-feedback gates. No history-based parameter sweep or
   retrospective favorable-window selection. Keep original log and Brier
   controls and the recorded failures of prior mechanisms.

Success is not assumed: alpha .001 may still adapt too slowly, this model may
be misspecified, or observation choice may dominate. If the mechanism fails,
retain the negative result and investigate those distinct causes. All seven
research directions remain open; this addresses only part of directions1/2/4.
