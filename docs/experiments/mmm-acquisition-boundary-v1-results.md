# Boundary alignment result

Status: mechanism works, quality rescue FAIL.669/672 non-harm checks and0/64
required gain checks pass. All2688 consumed runs retained; no fresh confirmation.

Disagreement acquisition gives new/earlier training evidence on58.14% of queries,
up from40.77% at original timing. Mean paid-reveal-to-first-fit wait falls from
18.56 to11.25 frames among included queries.97.26% of the9408 shifted boundary
queries supply evidence not yet naturally available at their immediate next fit.
All shifted queries enter that fit; remaining naturally known-by-fit cases reflect
labels whose natural arrival occurs on that same clock. Total41664 queries,
unchanged. No-fit-in-horizon cases fall from5376 to4032.

Phase1 delayed terminal64 served Brier:

| Case | Natural | Aligned random | Aligned entropy | Original disagreement | Aligned disagreement |
| --- | ---: | ---: | ---: | ---: | ---: |
| Additive stationary | .221755 | .221620 | .220334 | .220710 | .220085 |
| Parity4 | .049066 | .048837 | .048584 | .048812 | .048707 |
| Null | .258525 | .257035 | .257204 | .258398 | .257492 |
| Majority to parity | .060262 | .054548 | .056047 | .056180 | .054164 |
| Parity to majority | .102272 | .099485 | .096966 | .096202 | .096574 |

Non-harm failures: phase0/case1 delayed terminal against aligned random and
entropy (lower gains -.010412/-.012797); phase1/case20 delayed terminal against
aligned entropy (lower -.010861). These are uncertainty-bound failures, not proof
of uniformly negative mean effects. All gain gates fail. Do not adopt timing
alignment as a validated accuracy upgrade solely because its mechanism improved.

## Verification

Boundary and original-mode contract suites passed under race (45.02s/44.56s).
Full collector341.96s test time,344.64s shell wall, four workers, no dropped errors.
Independent scorer validates13,762,560 probabilities, training-origin lists,
equal actual costs versus original timing and immediate-delivery identity.
Maximum metric discrepancy3.33e-16. Full scoring replay byte-identical; not a
second full collection or a loaded serving benchmark.

Original helper/collector snapshots match prior artifact hashes and live in
`research/acquisition-training-v1-source`; archived original scorers retained too.
Default research mode still uses original timing. No production path changed.

Next diagnostic: compare fully immediate natural feedback with delayed feedback
on the SAME underlying trajectories. That can show whether feedback timeliness
is a substantial whole-learner limitation before more acquisition tuning. It is
not an oracle ceiling: additional labels can change fitted models adversely.
Verify paired X/Y/Q and initial samples rather than assuming schedule metadata
proves identical problems. Preserve every case and per-expert outputs.
All seven whole research directions remain open.

Artifacts: [protocol](mmm-acquisition-boundary-protocol.md),
[raw](mmm-acquisition-boundary-v1.jsonl), [summary](mmm-acquisition-boundary-v1-summary.json),
[scoring replay](mmm-acquisition-boundary-v1-summary-replay.json),
[training lead](mmm-acquisition-boundary-training-lead.json),
[race checks](mmm-acquisition-boundary-contracts.txt), [run log](mmm-acquisition-boundary-v1-run.txt).
