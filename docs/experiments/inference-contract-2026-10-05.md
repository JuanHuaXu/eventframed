# Stream-level paired-loss inference contract

Date: 2026-10-05. This is an additive research utility and a prospective inference
contract, NOT scientific confirmation, an experiment rerun, production integration,
or a replacement verdict. All existing protocols, artifacts and verdicts remain
unchanged. No new experiment is preregistered merely by publishing this document.

## Estimand and independence unit

For comparison j, define the completed-stream observation

```text
X[j,i] = average over prespecified clocks t of
         ((p_control[j,i,t] - Y[i,t])^2 - (p_candidate[j,i,t] - Y[i,t])^2).
```

Both arms use the SAME Bernoulli outcome Y in {0,1}; forecasts are issued before
that outcome and probabilities lie in [0,1]. Every clock gain and stream average
lies in [-1,1]. Positive mean is candidate benefit; candidate harm reverses the
sign. The convention is control minus candidate, not candidate minus control.
For a benefit margin tau, require the lower endpoint above tau; for no more than
h harm, require the lower gain endpoint at least -h. These are statistical
conditions, not built-in experiment acceptance rules.

One `AddStream` call is ONE equally weighted paired stream mean, or one
prespecified fit-cluster mean when fits are the replication unit. It is not one
clock, label, arm, forecast, or fitting call. `AddPairedStream` averages the
within-stream pairs first. Unequal stream lengths still get equal stream weight;
this targets an average over streams, not a pooled per-clock average. Freeze the
window, inclusion, missingness, weighting and stream order before inspecting
their losses. Include declared warm-up, adaptation, failures and delayed outcomes.
Outcome-dependent omission, duplicate streams, post-hoc windows or ordering by
completion when completion reveals losses can break the assumptions. Availability
and missingness require their own justification; the utility cannot supply it.

The exact stochastic assumption, for EACH frozen comparison, is

```text
X[j,n] in [-1,1] almost surely;
E[X[j,n] | F[n-1]] = mu[j] for every n, for one constant mu[j].
```

F contains the previously observed information and any information used to
choose the next unit. Independent streams from an unchanged generating law
and frozen policy are a sufficient special case; independence alone without a
common mean is not enough for this estimand. Dependence within a completed stream
is allowed. A stream may contain prespecified shifts and adaptive learning:
the *stream-average* law must still have the same conditional mean across units.
General drift across streams is NOT supported. Sharing outcomes across comparisons
is allowed: the family union bound does not assume comparison independence.

If streams share one fitted model, conditional independent-stream inference
targets THAT fitted model's performance. It does not integrate training-seed
uncertainty. To target new random fits, independently sample fit clusters and
submit one prespecified bounded average per cluster, with the same constant-mean
condition across clusters. Never count all streams in shared-fit clusters as
independent random-fit evidence. Equal cluster averages estimate an average over
fits; pooling unequal cluster sizes changes the estimand. An inherited stream
SE is not a fit-cluster SE.

## Frozen family and proof

`NewConfidenceSequence(alpha, comparisonNames)` copies a nonempty, unique family
of m names and requires finite 0 < alpha < 1. Alpha and the family have no setters.
Allocate alpha/m to each two-sided comparison, even if some comparisons are not
observed. Names must encode the actual hypotheses: control, candidate, scenario,
window, split and any distinct endpoint to which a claim will attach. More
comparisons across candidates, scenarios, windows, cohorts or restarts need a
family budget allocated before inspecting outcomes. No automatic allocation for
the entire research history is provided. Constructing additional objects or
renaming comparisons does not create fresh error budget.

For n >= 1, conditional Hoeffding's lemma on an interval of width 2 gives

```text
E[exp(lambda*(X[j,i]-mu[j])) | F[i-1]] <= exp(lambda^2/2).
P(|mean[j,n]-mu[j]| >= r) <= 2*exp(-n*r^2/2).
delta[j,n] = alpha / (m*n*(n+1)).
r[j,n] = sqrt(2*log(2*m*n*(n+1)/alpha)/n).
C[j,n] = [mean[j,n]-r[j,n], mean[j,n]+r[j,n]] intersect [-1,1].
```

The MGF inequality iterates across the n stream units; optimizing the Chernoff
bound in each tail gives the stated two-tail inequality. Union over all j and
all n, using sum(n>=1) 1/(n*(n+1)) = 1, yields

```text
P(for every j and every n>=1, mu[j] belongs to C[j,n]) >= 1-alpha.
```

Thus `Anytime` supports repeated looks and stopping at any stopping time under
these assumptions, without a predetermined horizon. Comparisons may have different
unit counts. At n=0 the interval is the vacuous [-1,1]; reported Mean=0 is only a
placeholder. Each displayed interval is the current mean plus/minus radius, not
a running intersection, and endpoints need not move monotonically.

This proof is a direct union-bound extension of
[Hoeffding (1963)](https://doi.org/10.1080/01621459.1963.10500830).
For the distinction between fixed-time inference and time-uniform coverage, see
[Howard et al. (2021)](https://arxiv.org/abs/1810.08240).
No empirical-Bernstein, betting, LIL-rate or variance-adaptive claim is imported.

The mapping to the parent Appendix C supplied for integration is exact: its K
is this implementation's frozen family size m, its j is this comparison's
completed-unit count n, and alphaFamily is alpha. Thus its
`alpha_j,c = alphaFamily/[K*j*(j+1)]` and
`sqrt(2*log(2/alpha_j,c)/j)` are precisely `Anytime`'s allocation and halfwidth.
The range is [-1,1], with one fixed conditional mean per comparison. The explicit
`TestAppendixCFormula` checks this notation directly. No stronger betting method
is implemented or substituted for Appendix C. Parent source integration remains
a separate task; this document does not claim to have inspected or edited it.

`FixedSample` instead uses radius sqrt(2*log(2*m/alpha)/n). Its coverage requires
each evaluated sample count to be fixed independently of evaluated outcomes;
it is NOT valid after outcome-driven stopping, repeated-look selection or selecting
the favorable count. Both methods expose the same accumulated mean, but represent
different analyses, not two independent opportunities to reject. Choose the
inferential analysis prospectively. Fixed-sample normal mean +/-3.5 SE summaries
are different again: approximate, estimated-variance intervals, not this exact
bounded fixed-sample inequality and not confidence sequences.

## Reuse and prospective registration

An anytime guarantee protects stopping within the declared process, not model,
window, hypothesis or alpha selection using its outcomes. Applying this new
contract after inspecting old tapes cannot retroactively preregister them or
replace their frozen verdicts. Any later analysis of consumed tapes must be
labeled retrospective/diagnostic, not fresh confirmation. Old design and
confirmation labels do not become independent evidence again after a reanalysis.

A future study must freeze its complete family and alpha, target population and
mean, comparison direction and substantive margins, fit conditioning or cluster
unit, fresh seed/data domains, forecasts and learning policy, stream windows,
missingness/failure rules, acquisition costs, order and stopping/reporting rule
before observing fresh evaluated outcomes. Optional stopping need not have a
fixed maximum horizon, but all supported and failed comparisons must be retained.
Allocate budgets across study restarts rather than resetting after an unfavorable
path. This package enforces numeric bounds and local family immutability, not
provenance, prediction timing, independence, deduplication or preregistration.

## Pilot planning, not confirmation

The archived [preserved-incumbent v3 results](mmm-preserved-v3-results.md) and
[summary](mmm-preserved-v3-summary.json) report, for confirmation/member_shift/post,
`mix_breadth_ap` minus `mix_mmm_ap`:

```text
pilot streams n0 = 32 (conditional on one original fitted model)
paired mean gain = 0.00859583394601159
old approximate interval = [0.0023387244853875296, 0.01485294340663565]
old frozen critical multiplier = 3.5
observed SE = (upper-lower)/7 = 0.00178774556017830
observed stream SD = SE*sqrt(32) = approximately 0.01011301
```

These numbers are read-only planning inputs, not a new measurement or a verdict
update. This incremental matched-observation contrast is distinct from the much
larger member-shift rescue gain versus frozen MMM (0.1981758), and from the
unresolved AP ablation (0.0005635). Do not substitute one claim for another.

`FixedSampleNormalPlan` implements the approximate directional-power formula

```text
fresh units = ceil(n0 * ((z_critical+z_power)*observed_SE/target_gain)^2),
             with a minimum of 2 units.
```

This is the real-valued formula of an approximate statistical planning model,
not a claim of formally certified floating-point upward rounding. The helper
first evaluates `observed_SE/target_gain`, so equal subnormal inputs cancel before
multiplication: n0=32, SE=target=SmallestNonzeroFloat64, z_critical=3.5 and
z_power=.84162123357 now give 604, not the former erroneous 512. Tests include
an exact rational reference for this regression and adjacent unequal subnormal
ratios, but do not certify every floating-point ceiling near integer boundaries.

Nonfinite, zero or subnormal dimensionless SE/target or scaled ratios, a
nonfinite critical sum, and counts exceeding 2^53 are conservatively rejected.
This can reject extreme combinations whose final real-arithmetic formula would
be representable after a different rescaling. Input SE and target themselves may
be subnormal if their ratio is supported. A small supported scaled ratio safely
below the two-unit floor returns 2 without squaring into underflow. No log-domain
fallback, interval arithmetic, or formal rounded-ceiling guarantee is claimed.

Use target_gain = hypothesized mean minus the null/benefit threshold, not raw gain
when testing a nonzero margin. Supply an independently chosen, multiplicity-aware
critical multiplier. The illustration uses the old 3.5 and z_power=0.841621234
(80% directional power for a positive alternative with a two-sided cutoff).
This gives the following algebraic TOTAL fresh-stream budgets, not counts to add
to or subtract from the consumed 32-stream pilot:

| Assumed future gain over zero | Approximate fresh units |
| --- | ---: |
| Full observed gain, 0.008595834 | 27 |
| Half observed gain, 0.004297917 | 105 |
| Quarter observed gain, 0.002148958 | 418 |

At 90% directional power (z_power=1.281551566), the full-gain estimate is 32.
These estimates do not recommend reducing a study to 27 units. They assume stable
variance, a correct normal approximation, independent identical units, an effect
that replicates, and a chosen z justified for the future family. Small n, selected
pilot effects, variance estimation uncertainty, changed policies/populations and
fit-seed variability can substantially inflate needs. Doubling the SE multiplies
the unrounded budget by four. A near-zero estimated pilot SE is not evidence of
certain power; this helper rejects zero SE. Pilot reuse and retrospective selection
make this planning, not an attained-power certificate or new significance test.
Use cluster-level pilot variability when targeting independent random fits.

### Reviewer target: unresolved AP member-shift ablation

The breadth-versus-MMM example above is not the AP ablation. The exact consumed
summary contrast for confirmation/member_shift/post, `mix_mmm_no_ap` minus
`mix_mmm_ap`, is:

```text
pilot streams n0 = 32 (conditional on the same single original fitted model)
paired mean gain = 0.0005635187818819783
old approximate interval = [-0.0003752428971604797, 0.0015022804609244361]
observed SE = (upper-lower)/7 = 0.0002682176225835594
observed stream SD = approximately 0.001517268
z_critical = 3.5; z_power = 0.8416212335729143 (80% directional power)
unrounded planning count = 136.6506399536487
approximate TOTAL fresh independent streams = 137
```

Using the rounded mean .000564 and interval [-.000375,.001502] also yields 137
after rounding upward. `TestAPMemberShiftNormalPlan` checks both. This is an
optimistic fixed-sample normal planning calculation under the same assumptions
and caveats as above, not attained power, a CS stopping budget, a fresh test,
or evidence resolving AP's incremental benefit. The old interval crosses zero;
the AP ablation remains unresolved. These units cannot be interpreted as 137
independent fits or as 105 additional streams after recycling the pilot 32.

Keep the separate confirmation/recurring/post AP contrast visible:
mean -0.0003482400977983611 with interval
[-0.0018459539459258038, 0.0011494737503290817]. It is negative and unresolved,
not the positive member-shift effect. `TestRecurringAPIsNotPositivePlanningTarget`
rejects passing that negative mean as a positive-benefit planning target; neither
its absolute value nor the breadth contrast substitutes for the AP member target.

This normal calculation is NOT a power estimate for `Anytime`. For illustration
only, alpha=.05 and m=100 give the reference Hoeffding radii at n=32:
fixed-sample 0.719985 and anytime 0.976483. The anytime radius first falls below
the pilot gain at about 970,812 units under the stated formula. Even that is just
a width comparison, not a probability of rejection or a guaranteed stopping
budget: the sample mean fluctuates. This deliberately simple bounded reference
is very conservative for small effects. A tighter future implementation would
need its own proved assumptions and tests before claiming variance adaptation.

## Implementation and verification boundary

New files are confined to `internal/researchstats/` plus this document; no existing
packages, experiment runners or verdict logic are edited or imported. There are
no external dependencies, file writes, scientific runs, fitted models or real
data accesses in the utility or its tests.

Patch-reasoning gate: the requested upgrade is additive, not a confirmed bug in
an old verdict. Candidate failure mechanisms are clock pseudoreplication,
optional-stopping misuse, and shared-fit/reused-data overclaiming. The invariant
is bounded completed-unit averaging with one immutable family budget; the concrete
falsifier is a radius whose two-tail Hoeffding error exceeds its allocated look
budget. Tests check that identity, telescoping spending, fixed/anytime separation,
family widening, signed and null synthetic controls, endpoints, zero observations,
invalid inputs, state atomicity, configuration copies, cancellation and count
saturation. A two-stream unequal-length test rejects accidental clock pooling.
This is not a modification of an existing upstream bug, so related-fix/PR lookup
is not a precondition; no branch or dependency changes are needed.

Construction uses O(m) memory; `AddStream` and each bound lookup take expected
O(1) time with O(1) state per comparison. `PairedBrierMean` uses O(T) time and
O(1) extra space for T supplied pairs. No input history is retained. A single
goroutine owns the sequence; copying/restarting it is unsupported. Counts above
2^53 are rejected. Compensated sums reduce cancellation and log-space spending
avoids alpha/count underflow and product overflow. Arithmetic is IEEE-754
float64, not formally certified outward-rounded interval arithmetic; the proof
is in real arithmetic. Tests of formulas and degenerate synthetic streams do
not empirically prove coverage for arbitrary laws or verify study assumptions.

Focused verification commands (no broader tests or experiment commands):

```sh
go test -mod=readonly ./internal/researchstats -count=1
go test -mod=readonly -race ./internal/researchstats -count=1
go test -mod=readonly -race ./internal/researchstats -count=1 -cover
go vet -mod=readonly ./internal/researchstats
go test -mod=readonly ./internal/researchstats -run '^$' -bench . -benchmem -count=3 -cpu=1
git diff --check -- internal/researchstats docs/experiments/inference-contract-2026-10-05.md
```

The scoped tracked-file diff check does not include untracked additions. For
these new files, also check each actual addition without staging it:

```sh
git diff --no-index --check -- /dev/null <new-file-path>
```

Local results on 2026-10-05, Go 1.27.1, darwin/arm64, Apple M4: focused tests,
race tests, example and vet passed; race/coverage run reported 100.0% statement
coverage. Three serial one-CPU benchmark repetitions reported 29.59--29.75 ns/op
for update plus anytime bound and 2561--2579 ns/op for a 512-pair Brier mean.
Both measured 0 B/op and 0 allocs/op, excluding construction and input creation.

Benchmarks measure only local stream updates plus interval calculation, and
averaging 512 already-materialized paired Brier observations. They exclude
forecasting, training, acquisition, I/O, scheduling and end-to-end service costs.
They are not scientific evidence or a production latency certificate. No commit,
push, dependency install, scientific confirmation or historical re-verdict is part
of this task.

## Generated paper handoff

The short [utility report](inference-utility-2026-10-05-report.json) contains
generated synthetic interval output, the exact AP planning contrast and recurring
context, and fresh utility-only benchmark values. Raw local command records are
in `evidence/researchstats-2026-10-05/`, with source digests for reproduction.
These records contain deterministic synthetic inputs and public/archived aggregate
planning values only, no raw private data. Parent integration/copying into paper
source or its evidence tree is separate; old verdicts remain unchanged.
