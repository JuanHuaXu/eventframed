# Paired current-context certificate v3: frozen protocol

The v2 historical-reference window can flag a live change even when the
reference member changes in the same way. It is not a current-pair diameter
certificate without a stationary reference assumption. This research-only
study tests a narrower valid design: two contemporaneous observations with
the *same fixed context bit* and independently generated outcomes. It is
not a method for unmatched event streams.

For 512 clocks, draw one fair context `X_t` and independent reference/live
Bernoulli outcomes conditional on that shared context. Both start with
`P(Y=1|X=0)=0.9`, `P(Y=1|X=1)=0.1`. Cases are `stable`, `live_shift256`,
`reference_shift256`, `common_shift256`, and `live_shift0`; a shift swaps the
two context probabilities for the named side(s). Use independent seed bases
`2026100501` design and `2026100502` confirmation, 1,000 trajectories per
case/schedule/split. Random roles are independent and paired across arms.

Schedules are `complete` (all pairs immediate) and `sparse_delayed` (one
outcome-independent 25% nomination per pair, independent 20% missingness per
member, independent uniform delay 0..31 per member). A nominated pair is
eligible only after both member outcomes arrive; missing or still-pending
pairs never update a gate. Each nomination acquires two context readings;
each delivered pair exposes two labels. All costs are reported.

For an eligible pair, define `D_t=Y_ref,t-Y_live,t` in `[-1,1]`. The candidate
gate keeps only origin clocks `[clock-127, clock]`, and in each context tests
the mean of its arrived `D` values. Its finite-horizon radius at `n>0` is

`r_cond(n)=sqrt(2*log(4*(T+1)/delta)/n)`,

with `T=512`, `delta=0.02`. Flag once if for either context
`abs(mean D) > epsilon + r_cond(n)` with `epsilon=0.10`. This union-bounds
two-sided bounded-difference deviations over two fixed cells and at most
`T+1` check times, assuming the paired differences have conditional mean
within `[-epsilon,epsilon]`, independent/mean-zero noise across origin clocks,
and nomination/arrival independent of outcomes. A common change preserves
that null at each clock. Batch baselines and hidden change times are not used.

Controls see the exact same eligible pairs: a cumulative contextual gate
with the same radius, and a 128-clock *scalar* gate that discards `X` and
uses `r_scalar(n)=sqrt(2*log(2*(T+1)/delta)/n)`. Each arm has its own
`delta=0.02` false-flag statement; no joint family-wise guarantee across
arms or multiple buckets is claimed. Every forecast or split downstream
would still need a separately frozen test.

Report flags/1,000, prechange flags, median/p95 flag clock, nominated,
delivered, missing and pending pairs, acquired context readings, usable
labels, and experiment wall time. Candidate success requires <=20/1,000
flags in both `stable` and `common_shift256`, >=800/1,000 flags in both
member-specific clock-256 shift cases, no prechange flags for those cases,
and median detection delay <=192 clocks in the complete schedule. The
`live_shift0` case must flag >=800/1,000. Apply these screens to both
schedules; a sparse-schedule failure is a full-study failure, not a reason
to relax the audit budget. Preserve every row, independent summary
verification and exact replay. Passing would establish only fixed-context
paired-evidence power, not actual MMM Anti-Pigeon authority, downstream
Brier, unmatched-event validity or serving latency.
