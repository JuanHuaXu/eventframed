# Evidence gate refinement v1 results

2026-09-12. [Frozen protocol](mmm-gate-v1-protocol.md),
[summary](mmm-gate-v1-summary.json), [raw outcomes](mmm-gate-v1.json.gz).

10,240 streams, 512 steps each, four paired gates. Design and confirmation used
disjoint seed domains with no tuning between splits. All inputs are generated
paired correctness differences, not private conversations or real outcomes.

## Frozen verdict

Grid mixture PASSED; adaptive and fixed/adaptive hedge FAILED. This is an
isolated evidence-gate result, not evidence that forecast accuracy or target-law
diameter estimation improved. Production and existing MMM experiments are intact.

Confirmation restricted mean delay (lower is better; premature alerts and
misses are charged the full remaining horizon):

| Scenario | Fixed | Grid | Adaptive | Hedge |
| --- | ---: | ---: | ---: | ---: |
| Moderate shift128 | 154.53 | 137.17 | 147.25 | 145.30 |
| Moderate shift256 | 155.30 | 137.73 | 144.75 | 145.30 |
| Strong shift256 | 61.50 | 35.24 | 33.98 | 36.47 |
| Negative shift256 | 159.60 | 142.12 | 149.94 | 151.66 |
| Weak shift256 | 255.93 | 255.20 | 254.70 | 255.34 |

Grid gains on the two primary moderate shifts are 17.36 and 17.57 steps, about
11% each, with paired z=3.3 lower bounds 12.59 and 13.56. Both exceed the frozen
10% improvement rule. Adaptive/hedge fall short of that minimum even where
their improvement is positive. No premature alerts occurred in the primary
confirmation shifts for any arm.

## Important tradeoff: missed deadlines

| Scenario | Fixed detected /512 | Grid detected /512 |
| --- | ---: | ---: |
| Moderate shift128 | 511 | 510 |
| Moderate shift256 | 500 | 488 |
| Strong shift256 | 512 | 512 |
| Negative shift256 | 494 | 479 |
| Weak shift256 | 1 | 5 |

The grid is faster on average even after charging misses, but misses MORE
moderate and negative changes by the deadline. A mean-based pass must not hide
this. Do not deploy based on this result. Freeze a fresh deadline-miss
noninferiority comparison before adopting a replacement. Weak shifts remain
largely undetected; none of these variants solves that regime within 512 steps.

## False revocations and guarantee boundary

In each of symmetric, sparse, low-variance boundary and dependent null scenarios,
all gates had 0/512 alerts; per-scenario Wilson95 upper is 0.745%. In the
high-variance boundary null, fixed/grid had 2/512 (upper 1.413%) and
adaptive/hedge 1/512 (upper 1.098%). All pass the frozen 2% empirical upper gate.
These intervals are per scenario, not simultaneous coverage over the table.

The mathematical 1% per-gate control comes from the conditional-mean
supermartingale argument in the protocol, NOT from these finite counts. It
requires bounded differences and predictable pair selection/bets. It does not
establish target-law diameter, causal identification, validity under selectively
missing labels, or simultaneous 1% control after choosing among these gates.

## Verification and cost

Unit/race tests and vet passed. Tests verify baseline agreement with the actual
v3 gate, first-step predictability, positive null factors, start timing, latching
and invalid-input rejection without state mutation. The entire raw experiment
and summaries replayed exactly, with source/protocol hashes verified and all
10,240 seeds distinct. Replay uses the same algorithms, not an independent
statistical implementation.

[Microbenchmark](mmm-gate-v1-benchmark.txt) measures all four gates together after
all eight starts are active, with the output retained. It is not production
latency and excludes forecasts, retrieval, queueing and persistence. Log-space
wealth bounds numerical growth; state is fixed-size for this eight-start
contract. Arbitrary restart schedules require another declared budget.

Research basis: predictable bounded betting and wealth mixtures from
[Waudby-Smith & Ramdas](https://arxiv.org/html/2010.09686v7). Our finite grid and
adaptive-rate heuristic are explicit adaptations, not optimality claims.
