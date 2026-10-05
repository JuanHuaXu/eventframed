# Whole-learner feedback comparison

Status: diagnostic completed; no new acquisition policy validated.1344 paired
trajectories have matching SHA256 signatures of Initial and all256 X/Y/Q values.
All immediate runs have Delay0 and Missing=false.10,321,920 original forecast
values rescored for15 arms; all metric checks pass. Both schedules are one pair,
not independent evidence. All21 cases retained in each consumed phase.

Phase1 terminal64 original four-expert Markov Brier:

| Case | Delayed/missing | Immediate/complete | Paired gain interval |
| --- | ---: | ---: | --- |
| Additive stationary | .221755 | .220269 | [-.009171,.012144] |
| Additive gradual | .239065 | .225874 | [.004551,.021831] |
| Parity4 | .049066 | .048973 | [-.001390,.001576] |
| Null | .258525 | .255557 | [-.001806,.007741] |
| Majority to parity | .060262 | .049675 | [.000623,.020552] |
| Parity to majority | .102272 | .077009 | [.003365,.047163] |

Each phase has3/21 terminal mixer cells with lower gain>0 and18 intervals crossing
zero; none has upper gain<0. Positive cells differ: phase0 cases1,2,20;
phase1 cases2,19,20. These exploratory intervals are not simultaneous inference.

Parity-to-majority improves.025264 in phase1 and.024801 in phase0. That establishes
feedback sensitivity on these tests, not that31 selected queries can match full
feedback. Delay, missingness and changed sample ages under count caps remain
confounded. Immediate/complete is not an oracle bound: more evidence can change
fitted models adversely. Next use the frozen2x2 delay/missingness diagnostic.
All seven full directions remain open.

Artifacts: [protocol](mmm-feedback-pairs-protocol.md),
[results](mmm-feedback-pairs-v120.json), [replay](mmm-feedback-pairs-v120-replay.json).
Full diagnostic replay is byte-identical (`cmp` exit0).
