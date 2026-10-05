# Boolean specialist v90 results

## Verdict

**FAIL the frozen advance conjunction:10/12 required cells pass.** Both n64
parity3 cells fail the0.01 mean-gain threshold. Preserve this classification.
The model has strong finite parity-learning evidence, but large non-parity
regressions falsify using it as a replacement for the generic learner.
No delayed-stream integration or adoption has been validated.

The experiment uses1,280 independent training sets with nested16/32/64 prefixes,
producing3,840 paired per-fit records. Confirmation generator-mask pools are
disjoint from design pools. All512 raw test inputs are integrated exactly with
fresh-outcome probabilities under a known finite simulator. The numbers below
are expected simulator scores, not observed agent-answer rates or grokking.

## Full-input confirmation

| Rule | Labels | Subset Brier | Boolean Brier | Subset expected accuracy | Boolean expected accuracy |
| --- | ---: | ---: | ---: | ---: | ---: |
| Parity3 | 32 | 0.08414 | 0.05115 | 90.98% | 95.00% |
| Parity3 | 64 | 0.05732 | 0.04826 | 94.65% | 95.00% |
| Parity4 | 32 | 0.15995 | 0.05512 | 77.35% | 94.30% |
| Parity4 | 64 | 0.07565 | 0.04870 | 92.02% | 95.00% |
| Complemented parity4 | 32 | 0.15112 | 0.05122 | 78.28% | 95.00% |
| Majority3 | 64 | 0.05489 | 0.19061 | 95.00% | 74.26% |
| Multiplexer3 | 64 | 0.05619 | 0.18790 | 94.91% | 73.92% |
| Dependent parity4 | 64 | 0.06741 | 0.04839 | 92.47% | 95.00% |
| Null | 64 | 0.26956 | 0.25326 | 50.00% | 50.00% |

For noisy deterministic rules, optimal full-input Brier is0.0475 and maximum
expected accuracy95%; null optimal Brier is0.25. The specialist does not create
null skill. The full-input evaluation isolates model fit, not the cost or ability
of an actual observer to obtain every input.

Confirmation parity4 at32 labels gains0.10483 Brier, approximate across-fit
mean +/-3.5SE interval[0.08398,0.12568]. At64 labels its gain0.02695 has
interval[0.02015,0.03375]. Majority3 at64 labels instead loses0.13571,
gain interval[-0.14247,-0.12896]; multiplexer loses0.13171,
interval[-0.13819,-0.12523]. These are fitted-sample intervals, not confidence
sequences, real-world certificates or uncertainty over sampled test labels.

## Protocol defect, not a retroactive pass

The n64 parity3 control's full possible gain above the simulator floor is
0.00967 in design and0.00982 in confirmation. Requiring0.01 is unattainable on
these cells even for an optimal law. Actual gains are0.00862 and0.00906; the
confirmation interval[0.00529,0.01283] remains positive but the threshold fails.
This is a defect in the advance rule we froze, not an arithmetic/model failure
or grounds to change its recorded verdict. A future experiment should declare
headroom-sensitive criteria before using fresh evidence. The v88 delayed-stream
failure is unchanged, and no whole research direction is complete.

## Observation and dependence limits

Under the common fixed cost6 mask63, confirmation parity4 at64 labels has
Boolean Brier0.20899 and accuracy59.14%, despite its95% full-input accuracy.
The exact conditional risk floor for this fixed view averages0.20887: many
target masks need unobserved coordinates. This demonstrates an observation
limit in this fixture, not failure of an adaptive MMM observer we did not run.

The dependent-input control's partial Brier is0.20563 versus its true floor
0.19621, despite full-input Brier0.04839. Our uniform marginalization misses
some input dependence. Good full-input fitting does not validate that assumed
input law. Neither family learning nor additional confidence can remove that
modeling gap by declaration.

## Cost

Apple M4, darwin/arm64, Go1.27.1, benchmark suffix10. Same input fixtures for both
builders,500ms per benchmark, three repeats, no experiment running concurrently.

| Component | Time range | Allocation per fit |
| --- | ---: | ---: |
| Subset,16 labels | 6.110-6.167ms | ~651,264 bytes,3 allocations |
| Boolean,16 labels | 0.343-0.356ms | 335,872 bytes,2 allocations |
| Subset,64 labels | 6.949-6.980ms | ~651,264 bytes,3 allocations |
| Boolean,64 labels | 0.507-0.533ms | 335,872 bytes,2 allocations |
| Subset,256 labels | 10.243-10.387ms | 651,264 bytes,3 allocations |
| Boolean,256 labels | 1.509-1.524ms | 335,872 bytes,2 allocations |
| Boolean compiled forecast | 6.995-7.495ns | 0 bytes,0 allocations |

The lookup microbenchmark is only a probability lookup on a supplied mask,
not an MMM traversal, complete forecast bundle, daemon request or tail latency.
Compiled cells occupy19683*16=314,928 payload bytes before allocator rounding;
fit allocation totals include temporary model arrays. Evidence isO(n*2^d),
full predictive compilationO(4^d) and ternary marginal compilationO(d*3^d)
with d=9 fixed. This is not a scalable high-dimensional parity solver.

## Verification and next action

Unit/race contracts pass: Beta sequence evidence, normalization, complement and
permutation symmetry, sample-order invariance, all19,683 analytic partial
marginals, immutable concurrent reads and input bounds. One initial constant
test expected unjustified certainty from impoverished input support; it was
corrected and retained as an uncertainty negative control. Generator checks
cover disjoint pools, masks within budget, oracle score bounds and null scoring.

Generation: PASS27.860 seconds. Focused race: PASS1.486 seconds. Vet and diff
checks pass. Full3,840-record artifact replay: PASS27.941 seconds.
The final race sweep including unchanged subset evidence/marginal regressions
also passes,1.547 seconds.
The independent evaluator verifies11 source/protocol hashes and all records.

Next lead: [coherent generic/Boolean family averaging](../../research/boolean-family-rescue-proposal.md).
This is proposed, not implemented or validated. It must protect non-parity
cases on fresh data before a separately frozen delayed-stream integration.

Artifacts: [protocol](mmm-boolean-specialist-v90-protocol.md),
[raw records](mmm-boolean-specialist-v90.json),
[summary](mmm-boolean-specialist-v90-summary.json),
[benchmark output](mmm-boolean-specialist-v90-benchmarks.txt),
[benchmark command and source hashes](mmm-boolean-specialist-v90-benchmark-metadata.json).
Raw SHA256: `bbc1dd6609fcd027b3fc5897239177b29b8cf61a268e7b22010a3cc2f006c51f`.

No production/OpenClaw changes, no commit or push. All seven directions remain open.
