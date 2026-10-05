# Skeptical family prior v92 results

## Verdict

**FAIL the complete frozen gate.** The skeptical prior passes all12 targeted
excess-risk conditions and59/60 case/phase/sample-count non-harm conditions.
The equal-prior comparison passes12/12 and52/60 respectively on these same
fresh samples. Better protection is not complete protection; no adoption.

All30 skeptical confirmation cells pass, but design majority3 at16 labels
fails both full and partial non-harm bounds. Both phases are required. Do not
discard the design failure, retune the prior, or call the confirmation pass an
overall rescue. v90/v91/v88 verdicts remain unchanged.

The experiment has1,280 fresh independent training sets with nested16/32/64
prefixes and3,840 three-arm records. Role seeds are disjoint from v90/v91.
Expected scores integrate512 raw inputs and fresh simulator outcome probabilities;
they are not chatbot accuracy, grokking, or production MMM acquisition results.

## Confirmation examples

| Rule | Labels | Generic Brier | Skeptical Brier | Generic expected accuracy | Skeptical expected accuracy |
| --- | ---: | ---: | ---: | ---: | ---: |
| Parity4 | 16 | 0.26076 | 0.17675 | 51.60% | 72.40% |
| Parity4 | 32 | 0.14909 | 0.05350 | 78.95% | 95.00% |
| Parity4 | 64 | 0.07225 | 0.04829 | 92.69% | 95.00% |
| Majority3 | 16 | 0.18762 | 0.18950 | 73.72% | 73.29% |
| Majority3 | 32 | 0.09705 | 0.09894 | 90.45% | 89.78% |
| Multiplexer3 | 16 | 0.18589 | 0.18743 | 74.42% | 74.30% |
| Multiplexer3 | 32 | 0.11105 | 0.11149 | 86.67% | 86.58% |
| Null | 64 | 0.26883 | 0.26416 | 50.00% | 50.00% |

Small regressions remain inside tolerance on these confirmation examples.
Parity4 at32 labels gains0.09559 Brier, approximate mean +/-3.5SE interval
[0.07872,0.11245]. At64 labels the gain is0.02396,[0.01842,0.02949].
These are across-fit normal bounds, not sequential coverage guarantees.

The prior changes posterior influence, not its maximum: confirmation n32
parity4 still gives the specialist93.98% average weight. At n16, majority and
multiplexer specialist weights fall to5.49% and6.00%. At64 labels, skeptical
null Brier0.26416 is worse than the equal-prior mixture's0.25756, though both
beat generic0.26883. Skepticism has a cost as well as a protective effect.

## The remaining failure is meaningful

Design n16 majority mean full-input harm is0.00401, but its upper3.5SE bound
is0.01337, above0.01. Partial mean harm0.00363 has upper bound0.01272.
The means alone do not exceed tolerance; the experiment fails to establish
the required bound. This is not grounds to ignore a heavy harmful tail.

A post-hoc examination of the recorded fits finds index37, rule mask266:
skeptical specialist weight0.83918, generic Brier0.23695, candidate0.40899,
and expected accuracy55.80% to50.00%. Its full-input harm0.17204 accounts for
about67% of this cell's total signed harm0.25661. The next-largest harm is
0.00707. This is a consumed diagnostic, not fresh confirmation and not a
reason to remove the observation or hard-code its mask.

Thus a skeptical prior can still be overwhelmed by misleading finite evidence.
Correct posterior arithmetic is not a guarantee of non-inferiority. The next
lead is predictive validation, not another prior fitted to these outcomes.

## Scope and cost

The fixed partial view still limits parity4 n64 expected accuracy to57.03%,
with Brier0.21842 near its view-specific oracle0.21836. Dependent-input n64
partial Brier0.19386 remains above its true floor0.18039. No change repairs
observation coverage or the uniform input assumption. Delayed learning and
real-world serving remain untested for this candidate.

Apple M4, darwin/arm64, Go1.27.1, benchmark suffix10; same fixtures,500ms, three
repeats, no concurrent experiment:

| Fit size | Equal prior | Skeptical prior |
| --- | ---: | ---: |
| 16 | 6.563-6.702ms | 6.566-6.793ms |
| 64 | 7.706-8.112ms | 7.704-7.729ms |
| 256 | 11.890-11.963ms | 11.927-12.078ms |

Both use about667,648 bytes and4 allocations per fit. Lookup is7.058-7.398ns,
zero allocations, only for a supplied-mask probability. These overlapping
ranges do not establish a meaningful speed advantage; prior changes leave the
same asymptotic fitting/compilation work. This is not daemon tail latency.

## Verification and next lead

Generation PASS82.536 seconds; focused race PASS1.832 seconds; vet and diff
checks pass. Tests cover prior endpoints/equal-prior recovery, analytic posterior
odds, all19,683 partial mixtures per tested prior, symmetry, immutable concurrent
reads, evidence/prior bounds and seed isolation. Full3,840-record replay
PASS82.729 seconds.
The independent evaluator verifies17 source/protocol hashes and all records.
The final race sweep including predecessor evidence contracts PASS1.955 seconds;
v90/v91/v92 summaries and source hashes reproduce, and benchmark hashes pass.

Next proposal: [predictive stacking with exact LOO shortcuts](../../research/predictive-stacking-proposal.md),
grounded in Yao et al.'s primary paper and explicitly subject to its small-sample
limitations. The proposed binary-score formula and conjugate shortcut need
independent tests; neither is implemented or validated yet. All seven research
directions remain open. No production/OpenClaw changes, commit or push.

Artifacts: [protocol](mmm-prior-v92-protocol.md), [records](mmm-prior-v92.json),
[summary](mmm-prior-v92-summary.json), [benchmarks](mmm-prior-v92-benchmarks.txt),
[benchmark metadata](mmm-prior-v92-benchmark-metadata.json).
Raw SHA256: `fa583cd78260deb9cd86a42f85404b8564d14fb84d822d7d9b4bc6fb3d0997ac`.
