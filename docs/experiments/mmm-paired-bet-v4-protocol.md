# Paired betting certificate v4: frozen protocol

This research-only successor to [v3](mmm-paired-context-v3-results.md) tests
whether a fixed bounded betting e-process can recover power at the same
sparse paired-evidence budget. No v3 seed, observation, or threshold is
reused for fitting. The null is a *per-clock conditional-mean* statement,
not an externally identified target-law diameter:

`|E[Y_ref,t-Y_live,t | X_t=x, pre-outcome information]| <= epsilon`

for both `x in {0,1}` and every clock, with `epsilon=0.10`. Pairs have a
shared known context and independent Bernoulli outcomes conditional on it.
The gate receives both labels atomically only when both have arrived;
one-sided labels are hidden from all gate state even though they count as
acquisition cost. Nomination, missingness and delays are independent of
outcomes. If those assumptions fail, no e-process validity is claimed.

Generate 512 clocks with a fair context bit; initial probabilities are
`P(Y=1|X=0)=0.9`, `P(Y=1|X=1)=0.1` on both sides. Frozen cases are
`stable`, `boundary` (reference probabilities 0.55/0.45, live 0.45/0.55
throughout), `common_shift256`, `live_shift256`, `reference_shift256`,
`live_shift224`, and `live_shift0`. A shift swaps initial probabilities
on the named side at or after its clock; common shift swaps both sides.
The `boundary` case tests both signs at the null's `epsilon` boundary.
Use independent seed bases `2026100601` design and `2026100602`
confirmation, 1,000 trajectories per case and schedule per split. Random
roles are independent and arms see the same realized pairs.

Schedules are `complete` (all pairs immediate) and `sparse_delayed`:
outcome-independent 25% pair nomination, independent 20% missingness
per side, and independent integer-uniform delays 0..31 per side. A pair
is eligible only at the maximum of its two arrival clocks and only if both
arrive before clock 512. Record nominations, context readings (two per
nomination), missing and pending pairs, observed labels on either side,
and usable paired labels. Partial labels never enter any gate or future
nomination decision.

For `D=Y_ref-Y_live in {-1,0,1}`, let starts be the fixed origin clocks
`S={0,64,128,192,256,320,384,448}`. For each start `s`, context `x`,
and sign `a in {-1,+1}`, hold wealth initially one. At clock `t`, for
each newly eligible pair with origin `o>=s` and context `x`, multiply by

`1 + lambda*(a*D-epsilon)`, with frozen `lambda=0.80`.

Future starts retain wealth one. No wealth is reset, multiplied across
starts, or recomputed using future labels. The candidate flags once when
the arithmetic mean of all 8x2x2 wealth processes is at least
`1/delta=50`, with `delta=0.02`. Under the stated null each factor is
positive (at least 0.12) and has conditional expectation at most one;
the mean is a nonnegative supermartingale. Ville's inequality gives an
anytime `delta` bound for *one* pair of members under this null. This
does not grant simultaneous error control over buckets, model-selected
contexts, or a current-law certificate after a historical divergence.

Controls consume the exact same eligible pairs: v3's 128-origin-clock
conditional Hoeffding gate with `r(n)=sqrt(2*log(4*513/delta)/n)`, and a
context-blind 128-clock Hoeffding gate with
`r(n)=sqrt(2*log(2*513/delta)/n)`. Each control has its own `delta` claim;
there is no joint family-wise statement across arms. Record flags,
prechange flags, median/p95 flag clock, resources, isolated experiment
wall time, and update counts. A conditional candidate pass requires
at most 20/1,000 flags in each null case (`stable`, `boundary`, common
co-drift) and at least 800/1,000 flags in each member-specific case in
*both* schedules. Clock-224 and clock-256 cases must have no prechange
flags. Median delay must be at most 192 clocks in complete and 224 in
sparse-delayed schedules for those cases. Full study fails if any cell
fails; do not relax thresholds after design or confirmation inspection.

Preserve every trial row, independent summary verification, source and
protocol hashes, exact confirmation replay, and negative controls for
future, unnominated, missing, duplicate, and partial-pair updates.
Passing would establish only a bounded, matched-pair component under
the declared observation model. Actual MMM nomination, causal or target-law
diameter, downstream Brier, loaded latency, and multi-bucket error remain
open. Method inspiration: [Waudby-Smith and Ramdas](https://arxiv.org/abs/2010.09686),
with the finite start/mean construction derived explicitly here.
