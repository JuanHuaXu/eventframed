# Full-input learner v118: no standalone replacement passes

Completed2688 fresh runs and exact replay. None of the four candidates passes
the frozen broad replacement rule or its complete structural-target screen.
Ridge64 has useful additive-task results, but not robust recovery. No incumbent
is removed and no result is promoted to the daemon or whitepaper.

| Candidate | Non-harm | Gain | Total | Structural target | Broad verdict |
| --- | --- | --- | --- | --- | --- |
| Ridge64 | 70/336 | 7/36 | 77/372 | 7/12 | FAIL |
| Ridge32 | 58/336 | 0/36 | 58/372 | 0/12 | FAIL |
| Context-tree64 | 226/336 | 2/36 | 228/372 | 2/12 | FAIL |
| Context-tree32 | 214/336 | 0/36 | 214/372 | 0/12 | FAIL |
| Combined | 568/1344 | 9/144 | 577/1488 | 9/48 | FAIL |

Non-harm compares each candidate against both matched-window generic and Boolean
controls across all21 cases, phases, schedules and two score windows. The
structural-target column is a predeclared subset of gain tests, not extra gates
or a substitute for the broad verdict. Even it does not fully pass.

## What improved, and what did not

Ridge64 passes stationary additive all-frame gain in BOTH phases and schedules.
Confirmation gains over generic64 are .008571 with paired interval
[.003773,.013369] for immediate feedback and .010110 [.004952,.015269] for delayed.
This is finite evidence for the additive learner hypothesis, not general learning.

Its three other gain passes are design/immediate additive abrupt and gradual,
and confirmation/immediate additive abrupt. All four delayed additive recovery
tests fail. In confirmation/delayed abrupt, mean gain .012073 has interval
[-.001160,.025306]; gradual has .002453 [-.010055,.014962], below the .005 mean
floor as well as crossing zero. A larger sample cannot be presumed to repair
that insufficient observed mean.

Context-tree64 passes only design/delayed hierarchy abrupt and gradual gain;
those successes do not survive the complete two-phase target requirement.
Both32-example candidates pass no required gain. Shortening the window is not
an established rescue for these learners.

Selected confirmation/delayed terminal64 Brier scores (lower is better):

| Case | Generic64 | Boolean64 | Ridge64 | Tree64 |
| --- | ---: | ---: | ---: | ---: |
| Additive/stationary | .217400 | .233843 | .202374 | .218284 |
| Additive/abrupt | .229750 | .233983 | .217677 | .227673 |
| Additive/gradual | .240664 | .247732 | .238211 | .237305 |
| Hierarchy/gradual | .254588 | .250793 | .267812 | .248030 |
| Local table/gradual | .260645 | .253286 | .283248 | .253852 |
| Parity4 | .070408 | .048262 | .288078 | .257275 |
| Null | .266249 | .253626 | .285759 | .255727 |
| Majority to parity | .131128 | .074073 | .298608 | .262136 |
| Parity to majority | .104675 | .206473 | .113641 | .120540 |

An additive logit does not represent higher-order parity; a depth-three tree
does not represent a general four-bit parity rule. Their poor standalone parity
results are therefore not a numerical solver bug. Null-task risk also shows
estimation/forecast dispersion: with true q=.5, Brier equals .25+(p-.5)^2.
That does not by itself establish posterior-variance miscalibration or prove a
Bayesian approximation will rescue performance.

All scores, paired bounds and cases, including the omitted32-example columns,
remain in [the complete summary](mmm-soft-learners-v118-summary.json). Intervals
are mean +/-3.5SE over32 trajectories: approximate fixed-sample screens, not
anytime or adaptive-research-history-wide guarantees.

## Scope and audit

- [Frozen protocol](mmm-soft-learners-v118-protocol.md) and
  [raw artifact](mmm-soft-learners-v118.json).
- All9 transfer cells and all12 original Boolean regimes,2 phases,32 indices:
 1344 latent trajectories,2688 schedule runs,688128 steps and5505024 forecasts.
- Every learner gets all9 query coordinates and identical eligible labels.
  This isolates fitting; it is NOT the six-coordinate MMM controller experiment.
  Query/initial reads total2448 per trajectory, with no duplicate coordinate
  charge when its already-observed packet's label arrives.
- Source/seed/as-of QA passed under the race detector (package7.692s). Scoped
  vet passed. No solver failure or outcome-driven adjustment occurred during
  quality generation; the earlier roundoff repair preceded this frozen run.
- Generation completed in561.27s and full exact replay in577.07s. These elapsed
  experiment durations are not serving latency or a comparative runtime claim.
- Independent verification passed all162 source hashes,6720 effective seed
  identities, teacher probabilities, pairing, evidence windows, scores and gates.
-1376256 ridge forecasts reconstruct directly from saved coefficients;
 43008 fitted ridge states independently satisfy objective/stationarity checks
  against their eligible labels. This checks numerical fitting, not calibration.
- The independent summary reproduced byte-for-byte after replay.
- Raw JSON size179286369 bytes, exclusive mode0600. The frozen pre-quality
  component report remains unchanged; this report supersedes its untested status.

Artifact SHA-256:
`b458fa8e11bc20a45d2380fb93726dd61acb3909c7586769e04d951b37380d78`.

## Next research step

Keep the existing specialists. The cheap MAP learner is a plausible additive
specialist but an unsafe broad replacement. The separately proposed
[variational predictive comparison](../../research/variational-logistic-component-plan.md)
will test uncertainty averaging under the SAME prior, not tune the ridge penalty
against these outputs. It still cannot make a linear model represent parity.

Any later composite needs pre-outcome evidence-based admission, retained incumbent
controls and its own fresh integration experiment; teacher family IDs and fitted
training accuracy are not legitimate admission signals. The present study does
not validate that composite. All seven research directions remain open.
