# Adaptive routed observation v103

**Overall FAIL: 106/108 frozen quality gates pass.** The two failures are the
additional claim that lookahead improves late parity4 expected Brier over the
cheaper coherent one-step observer by at least .005 with positive lower bound.
Do not relabel this as a full pass or erase those gates.

The routed learner's improvements do survive adaptive observation in this
finite immediate-feedback screen: all 48 non-harm gates against generic64,
all 48 against routed one-step, all six interaction gains against generic64,
and all four regime-recovery gains against generic64 pass. The second set
protects against a declared .01 harm allowance, not against every possible harm.

## Artifacts and protocol

- [Frozen protocol](mmm-adaptive-v103-protocol.md)
- [Complete step-level artifact](mmm-adaptive-v103.json): 768 trajectories,
  196,608 scored steps; 45,824,234 bytes, created mode 0600.
- [Independent reconstruction and gates](mmm-adaptive-v103-summary.json)
- [Evaluator](../../research/adaptive-v103-summary.mjs)
- [Driver](../../internal/observationlearners/adaptive_v103_test.go)
- [Paired benchmark output](mmm-adaptive-v103-benchmarks.txt)

Artifact SHA-256:
`173e2f1027b79455211f9eff313225f117f58890a03c424244f049d64fd265f0`.
All 21 source/protocol/evaluator hashes checked. Summary regenerates byte for
byte; all step-level scores and paid costs reconstruct. Full deterministic
replay of every trajectory passed. No interim quality inspection occurred.
Generation took 123.57 seconds; replay took 125.08 seconds.

The four consumed v102 fixed-mask checks reproduce both all-step and late
metrics exactly. They are compatibility tests, not fresh quality evidence.
The new 3,072 role-specific effective seeds are unique and disjoint from the
audited v90-v102 learner and null allocations. This does not establish
independence from every historical study or an independent task generator.

## Confirmation results

Expected Brier, lower is better. Each row averages 32 trajectories. All routed
arms use the same model families and full-frame training audits; their online
weights can diverge because each scores its own acquired observations.

| Case and segment | Generic observer | Routed, matched masks | Routed one-step | Routed lookahead | Random views | Fixed mask63 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Parity4, all | .098058 | .089474 | .087471 | .085520 | .222339 | .221513 |
| Parity4, late | .069626 | .058861 | .058131 | .057423 | .214549 | .215503 |
| Majority to parity, late | .206222 | .183270 | .182948 | .181409 | .260415 | .259293 |
| Parity to majority, late | .173148 | .166955 | .166558 | .167110 | .196250 | .195935 |

Late parity4 expected accuracy is 92.80% generic, 94.74% routed one-step and
94.80% routed lookahead. These are finite synthetic expected accuracies, not
real-agent accuracies or uniform guarantees across regimes. After changes,
lookahead's late accuracies are 74.30% and 77.77%, respectively.

Lookahead gains over generic64 on the two confirmation recovery cases:

- Majority to parity: .024814, interval [.011948, .037679].
- Parity to majority: .006038, interval [.000929, .011147].

The frozen failure, lookahead gain over routed one-step on late parity4:

| Phase | Mean gain | Approximate interval | Required |
| --- | ---: | --- | --- |
| Design | -.000709 | [-.002006, .000587] | mean >=.005 and lower >0 |
| Confirmation | .000708 | [-.001008, .002423] | mean >=.005 and lower >0 |

Intervals are paired mean +/-3.5 standard errors over trajectories, as frozen.
They are fixed-sample approximate intervals, not exact confidence sequences.
The small opposite-signed means do not support a general lookahead advantage.

## Acquisition and compute

On confirmation late parity4, mean paid coordinates are 5.636 for routed
one-step and 6.000 for lookahead. On majority-to-parity they are 5.726 and 5.983;
on parity-to-majority, 5.542 and 5.965. The planner's terminal-entropy objective
does not reward early information per coordinate; a higher acquisition cost is
therefore possible even with the same cap. No cost-efficiency gain was shown.
The matched arm shares the baseline's physical reads, not duplicate reads.

Every arm also receives 272 full nine-field training audits, including initial
training, per trajectory. This extra 2,448-coordinate training bill is separate
from foreground cost. Fixed mask63 and random-view quality are poor partly
because they can miss relevant fields; their results do not falsify the learned
full-input models or constitute a fair unlimited-observation baseline.

Lookahead evaluated a mean 1,009.75 distinct memoized states and 4,006.77 branch
transitions per frame; maxima were 1,103 and 4,702. These are planning operations,
not observed evidence. Hypothetical branches never update the posterior.

Paired known-parity microbenchmark on Apple M4, Go darwin/arm64, -10, three
500ms repetitions after replay completed:

| Policy | Time per observation | Bytes | Allocations |
| --- | --- | ---: | ---: |
| Coherent one-step | 3.378-4.788 us | 528 | 5 |
| Exact lookahead | 149.080-149.267 us | 180,752 | 6 |

This is approximately 31-44 times the component time on this fixture. It excludes
fitting, publication, routing feedback, SQLite, retrieval and daemon serving.
It is not a latency distribution, a production regression, or a sub-100ms
serving guarantee. The scratch arrays are bounded by the nine-bit model but
allocated per observation in this prototype.

Benchmark source SHA-256:
`2775f1031ae3aeeb06015b65f2ce72e4610f57fc723307499cf10d0a3585870d`.
Benchmark output SHA-256:
`e0277949f61cd88b1c35ef92792dd9656b6a259474af1256b6f59feda12dd574`.

## Implementation checks

The planner matches independent literal full-assignment enumeration at every
visited state for total budgets 1-4, with uniform and nonuniform common input
laws. A separate cross-scope parity example has zero one-view gain but positive
joint gain: all 512 inputs retain .5 with the greedy observer and recover the
supplied .05/.95 conditional law with lookahead within six coordinates. This is
a structural planner test with a known model, not a learned-model accuracy test.

Hidden unobserved suffix changes preserve the trace. Reader failure, epoch
mutation, duplicate feedback and reentrant feedback are rejected without partial
state changes. Race-enabled planner, observation, routing, bank and comparative
checks pass; compatibility and seed checks also pass under race instrumentation.
Vet and whitespace checks pass. Archived v102 kernels/artifacts are unchanged.

## Decision and next boundary

Retain the coherent one-step observer as the economical candidate. Keep exact
lookahead as a tested structural reference, not a default or a demonstrated
quality rescue. Its advantage on the constructed zero-gain example did not
translate into the predeclared learned-stream gain. No threshold change or
larger confirmation run is proposed to make this failed claim pass.

Next, test the cheaper candidate in the missing/delayed-feedback and member-split
setting, with original forecast snapshots and publication versions preserved.
The existing one-pending-outcome API cannot simply accept delayed labels; that
requires an explicit journal/version contract before a fresh integrated test.
The earlier delayed-feedback and member-split failures remain standing until
that evidence exists. Independent generators, scarce labels, real agent tasks,
poisoning/source assumptions and persistence-tail requirements remain open.

All seven research directions remain open. No production changes, whitepaper
edits, commits or pushes were made for this run.
