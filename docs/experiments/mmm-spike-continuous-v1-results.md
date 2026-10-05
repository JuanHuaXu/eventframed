# Continuous expert weights and budget: late-recovery rescue rejected

Frozen contract: `mmm-spike-continuous-v1-contract.md`. Source indices0-7,
21scenarios, two phases and schedules:672 trajectories,172,032 forecasts.
All data were previously consumed. No prospective or deployment claim.

| Method | Whole-stream expected Brier |
| --- | ---: |
| Markov incumbent | .156448819 |
| Reset32 guarded mixture | .154441468 |
| Continuous guarded mixture | .155041378 |
| Continuous unguarded mixture | .154615613 |
| Fixed half-mixture | .155699195 |
| Arrival-refitted challenger alone | .164420902 |

Continuous guarding improves whole-stream Brier by .001407442 versus Markov,
exploratory trajectory-cluster interval [.001120266,.001673438]. However, it
is worse than the matched reset32 control by .000599910, interval
[.000448962,.000753331]. This does not support continuity as the missing rescue.

## Late behavior matters

On terminal64 forecasts, continuous guarding is worse than Markov by
.000621136, interval [.000066299,.001150321], and worse than reset32 by
.000435407, interval [.000094585,.000804277]. Twenty-four trajectories have
terminal64 expected harm>.01. Only one whole-stream trajectory crosses that
threshold (phase0/parity-to-majority/index6/immediate, +.011161883): whole-stream
averaging hides much of the recovery problem.

For delayed parity-to-majority, terminal64 scenario mean harms are .020648610
(phase0) and .021623878 (phase1), with pointwise intervals
[.011436510,.032268026] and [.015624949,.027687938]. These are exploratory
eight-trajectory intervals, not simultaneous guarantees. They nevertheless
provide a clear reason not to promote this candidate on its average gain.

The global realized prefix budget holds to1.18e-16 rounding error. That does
not contradict the late harm: earlier gains can finance later excess loss.
A constructed unit negative control also demonstrates terminal-window harm
greater than .1 while preserving the whole-stream budget. Do not reinterpret
the global invariant as an interval-local or conditional expected-risk bound.

## Verification and cost

There are132,592 total fits, including84,378 newly computed and48,214 reused
from audited0/128/224 windows. New collection took835.90s, including source,
cached-state reads and raw-state writes. Reused inference is not free: its
earlier collection times were132.28s,164.23s and172.10s. These are artifact
collection measurements, not an isolated inference or serving benchmark.
Repeated unchanged-window fits at block boundaries are charged and checked
for exact state equality. Mixture weights and ledger do not reset there.

Independent source/origin/moment audits pass for all five new raw files;
every new fit converges (maximum iterations by clock355/376/525/275/364).
One earlier startup capped state remains included; it serves two forecasts.
The full132,592-fit computation has not been replayed. Numerical audits do
not independently integrate the full255-factor predictive mixture.

The budget core was extracted without changing its operations. All2,016
cached-window reset outputs exactly match earlier guarded artifacts. Continuous
aggregation replays exactly except elapsed time; postprocessing took1.118s
and1.095s, including input reads but excluding output serialization. The
offline guard scans history, O(T^2) per256-step trajectory; this is not an
unbounded production implementation. Prefix, delayed-label poisoning, missing
outcome and continuous-vs-reset controls pass. All-clock as-of tests and the
full spike race suite pass (83.679s). No running sessions remain.

SHA256:

- Compact predictions: `b54941e358dc9703b6cac3c988b969e98c620cfa27317d0e6a2555d871a1fbbe`
- Clock32 raw: `c5f86970edbe936578a413765099debd3ec00029053f2c45083569213e79aa95`
- Clock64 raw: `bf3fa164e4f2db90b419ee87721dd58503b3efefdcb40a0af8ff3b8b8ad7bb04`
- Clock96 raw: `4c8eaca78d788aa4606ae1f83b62b212e00f2927717e513119b05c2278be5a06`
- Clock160 raw: `69ba67385248c77745c170ef2a5e8a9461e2d6c267f58a51f140deb0c21f5dc5`
- Clock192 raw: `bcd5cb1466d00a8728cbeabb303c4c72c7b81160868a9c9c882af5bf9a255cfb`
- Scored result: `d24d8c2764ccfcad451dc576bf5784eda40c7ba8daefb949605a4e194a41efa6`
- Summary: `c3040f71ba7854b5a512a30916d17f463fe37780cb2fa8ced6295a46115497f5`
- Budget core: `aae61e98faecd98265d3e3e4b6aa043148c9d02b12496dee16a705e7d2001953`
- Scorer: `7194eb1573e41a8be3e6797c940c81d091e7fb5fd869aa3d134d3b873e3d2b25`

## Next lead

Static cumulative expert preferences can become inappropriate after shifts;
continuity alone did not fix the tested behavior. Separate expert switching
from the budget's ability to spend old gains. The source note
`research/delayed-fixed-share-source-note.md` records primary research for a
delayed switching filter, including permanent-missingness and regret caveats.
This is a proposed new experiment, not an already demonstrated rescue. Keep
the current negative result and the reset32 control. All seven goals remain
OPEN; production and the whitepaper are untouched.
