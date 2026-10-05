# Delay and missingness decomposition

Status: diagnostic completed. Both new feedback combinations refit the same
four experts and Markov mixer on all1344 consumed latent trajectories. No new
acquisition policy validated; all seven whole directions remain open.

Phase1 terminal64 served-mixer Brier, lower is better:

| Case | Complete immediate | Missing only | Delay only | Delay + missing |
| --- | ---: | ---: | ---: | ---: |
| Additive stationary | .220269 | .222063 | .222442 | .221755 |
| Additive gradual | .225874 | .229585 | .237185 | .239065 |
| Parity4 | .048973 | .049288 | .048633 | .049066 |
| Null | .255557 | .256423 | .259318 | .258525 |
| Majority to parity | .049675 | .049825 | .054097 | .060262 |
| Parity to majority | .077009 | .087174 | .096827 | .102272 |

Parity-to-majority delay cost with complete feedback is.019819, paired interval
[.006152,.033485]. Missingness cost at zero delay is.010166, interval
[-.001658,.021990]. Interaction is-.004721, interval[-.021363,.011922]. The sum
is the total.025264 penalty. Phase0 shows the same direction: delay.019988 with
positive lower bound; missingness.012859 with interval crossing zero. These are
exploratory intervals, not proof delay dominates missingness in every setting.

Positive terminal mixer delay-cost lower bounds occur in phase0 cases1,2,20 and
phase1 cases2,20. No terminal mixer missingness-only interval has positive lower
bound in either phase. This does NOT prove missing evidence is harmless; the
point effects and interactions can be substantial and estimates are uncertain.
More labels also alter sample ages under fixed count caps.

## Verification

Four-mode/two-control/eight-prefix tests passed under race40.54s. Full collection
88.08s shell wall (85.634s Go package), four workers.3,440,640 new probabilities
rescored, every fit-origin set checked against its intervened schedule, all1344
original trajectory signatures matched, and all delay/missingness/interaction
identities checked. Original complete/immediate and combined controls reused
from their independently verified paired artifact, not refitted unnecessarily.
Scorer replay is byte-identical (`cmp` exit0); no production throughput claim.

## Next direction

Deprioritize another missing-label recovery score on these fixtures. More relevant
is when the fixed acquisition budget is spent after evidence of change: the
previous policies selected WHICH label, mostly on a fixed eight-frame schedule.
An evidence-triggered burst policy could reduce waiting after the first revealing
label. It must use only arrived labels, retain the same total acquisition budget,
and face random/uncertainty and periodic controls plus stationary false-trigger
tests. This is a remaining lead, not an implemented or validated rescue. Do not
use known generator change times or hidden Q to trigger it.

Artifacts: [protocol](mmm-feedback-factorial-protocol.md),
[raw](mmm-feedback-factorial-v1.jsonl), [summary](mmm-feedback-factorial-v1-summary.json),
[scoring replay](mmm-feedback-factorial-v1-summary-replay.json),
[race tests](mmm-feedback-factorial-contracts.txt), [run log](mmm-feedback-factorial-v1-run.txt).
