# Next repair: shared reads with an explicit credit ledger

Status: implemented in an isolated test harness and passed the consumed shadow
screen; see docs/experiments/mmm-monitor-credit-v1-report.md. Closed-loop quality
and fresh confirmation remain untested. The original proposal follows.

The alternating full-monitor
policy improves reverse detection but fails forward retention and exceeds the
original per-trajectory cost on16/128 trajectories (32 paired schedules).

Preserve original bounded monitoring on every frame. Where a random learning
audit already requires a full frame, reuse the bounded observation rather than
charging it again. Bank only the resulting measured coordinate savings. Spend
available savings on completing BOTH reference/live views on later non-audit
frames; otherwise retain the original bounded observations. No borrowing from
future audits, outcomes, or projected average savings is allowed.

For each origin, let b be the actual bounded reference/live coordinate cost
(0<=b<=18), a the pre-outcome audit indicator, and c>=0 the existing credit.
The old charged allowance is b+18a. Start by charging b for bounded reads.
On an audit, completing the pair costs18-b, yielding total18 and adding b to
credit. On a non-audit, complete only if c>=18-b; subtract that additional
cost from credit. Otherwise retain bounded monitoring and leave credit unchanged.
An origin uses exactly one correctness pair, full if completed, bounded otherwise.

Inductively c equals cumulative old allowance minus cumulative candidate charge,
and c remains nonnegative. This gives a prefix coordinate-budget invariant,
unlike the rejected alternating schedule's lower average cost. Missing outcomes
do not refund acquisition cost. Learning-audit evidence is unchanged.

Required caveats and tests before collection:

- Actual reader masks/values must justify reuse. Unique-coordinate accounting
  is not proof of fewer backend requests, bytes, synchronization, or latency.
- Selection based on pair cost must be symmetric under exchanging reference
  and live. Under identical independent joint laws this preserves a zero-mean
  comparison. It does not automatically preserve the whole .15 tolerated null
  under arbitrary covariate/selection changes; state the conditional null.
- Exhaustively test budget boundaries b=0..18, a=0/1, and credits around18-b;
  then adversarial sequences, exact original-output fallback, audit preservation,
  and outcome/missing-label independence.
- Keep the external gate threshold and Bayesian policy fixed. Do not OR arms.
- Repeat all stable/shifted scenarios, record all costs, and require forward
  retention plus reverse improvement before closed-loop forecast testing.

The repair may still fail: a better observed view need not improve every sample's
decision, and the scalar tolerance may still conceal target-law differences.
Do not call this a rescue until those tests support it.
