# Faster forecast bank: selection and bank limitations

Diagnostic completed on2688 consumed trajectories, both cadences and both
scoring windows. No deployable selector or new learner has been validated.

Phase1 delayed terminal64, faster8-clock bank:

| Case | Served | Best fixed | One switch | Hull oracle | Brier floor |
| --- | ---: | ---: | ---: | ---: | ---: |
| Additive stationary | .221658 | .219149 | .214627 | .194089 | .173275 |
| Additive gradual | .234161 | .229188 | .222548 | .198771 | .173021 |
| Case4 | .245446 | .241744 | .237207 | .218214 | .198756 |
| Case5 | .242155 | .236600 | .231955 | .213192 | .191753 |
| Case7 | .244145 | .240463 | .235191 | .216535 | .196194 |
| Case8 | .248307 | .242021 | .237232 | .218900 | .196820 |
| Parity4 | .050053 | .048065 | .047956 | .047798 | .047500 |
| Null | .257049 | .252389 | .251686 | .250485 | .250000 |
| Majority to parity | .053252 | .049566 | .048864 | .048481 | .047500 |
| Parity to majority | .098235 | .079682 | .068195 | .062037 | .047500 |

The reverse Boolean transition retains.036198 served-minus-hull selection gap
and.014537 hull-minus-floor bank gap. A single hindsight switch has.030039
gain over the served mixture, but this assumes the correct initial expert and
switch clock are known from Q. It does not prove an online selector can learn
that choice from sparse/delayed observations.

Cases4,5,7,8 retain approximately.0195-.0221 bank gap and.0272-.0294 selection
gap. Neither faster publication nor selection alone can eliminate all error
in these issued banks. Bank gap is conditional on these particular predictions,
not a theorem that their model families can never improve after different fits.

For majority-to-parity, the remaining selection gap is only.004771; cadence
already closes much of the old.011113 gap by point estimate. The paired reduction
interval crosses zero, so do not claim an established causal decomposition from
these means. The full summary retains uncertainty for both gap reductions.

## Verification

For each frame the hull forecast is Q projected onto[min expert,max expert].
The served mixture is checked inside that hull. All10752 trajectory/cadence/
window decompositions satisfy served-floor = selection gap + bank gap, with
nonnegative components and floor <= hull <= framewise <= two-switch <=
one-switch <= best-fixed risk. Losses match the audited cadence summary and
both source hashes. The dynamic program passed7776 exhaustive small paths and
160 checks again. Scoring replay is byte-identical (`cmp` exit0).

The oracle restarts independently for the terminal window. All four expert
roles are matched across cadences; earlier six-expert oracles are not silently
substituted. These analyses use unavailable Q and are not implementation results.

Artifacts: [protocol](mmm-cadence-bank-protocol.md), [diagnostic](mmm-cadence-bank-v1.json),
[replay](mmm-cadence-bank-v1-replay.json).

## Next lead and prior-work check

The largest reverse-transition opportunity remains selection. The existing
segment-posterior implementation already computes generic/Boolean marginal
likelihoods from eligible labels; this is not a new invention. Its previous
full segmentation and static-atom rescue failed broad protection screens.
Before a new combined policy, inspect that evidence and test whether a simpler
same-window family-evidence update can address stale performance weighting
without importing all the segmentation machinery. Comparing likelihoods from
different sample counts directly would be invalid; window handling must be
explicit. No new candidate or success claim is made in this diagnostic.
All seven full goals remain open.
