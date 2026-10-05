# Generic/Boolean family v91 results

## Verdict

**FAIL the complete frozen advance rule.** All12 targeted excess-risk conditions
pass, but5 of60 case/phase/sample-count cells fail the all-view non-harm gate.
Every failure is at16 labels. The family model greatly reduces the specialist's
non-parity damage at larger sample counts, but sparse-evidence protection is
not established. No delayed-stream rescue or production adoption is validated.

There are1,280 independent fresh training sets and3,840 nested-prefix paired
records. Seed roles are disjoint from v90. The finite generator and disjoint
design/confirmation rule-pool specification are unchanged; this is not new
real-world domain evidence. Scores exactly integrate512 raw inputs and fresh
outcome probabilities. Expected accuracies below are simulator quantities, not
observed agent-answer rates, grokking or inherited94.7% task accuracy.

## Full-input confirmation

| Rule | Labels | Generic Brier | Family Brier | Generic expected accuracy | Family expected accuracy |
| --- | ---: | ---: | ---: | ---: | ---: |
| Parity3 | 32 | 0.07884 | 0.05085 | 91.90% | 95.00% |
| Parity4 | 32 | 0.14601 | 0.05129 | 79.65% | 95.00% |
| Complemented parity4 | 32 | 0.13534 | 0.04853 | 79.54% | 95.00% |
| Parity4 | 64 | 0.07078 | 0.04827 | 92.68% | 95.00% |
| Majority3 | 16 | 0.18659 | 0.19458 | 74.87% | 73.32% |
| Majority3 | 32 | 0.10098 | 0.10375 | 88.99% | 88.83% |
| Majority3 | 64 | 0.0564475 | 0.0564513 | 94.82% | 94.82% |
| Multiplexer3 | 16 | 0.19226 | 0.20771 | 72.94% | 69.49% |
| Multiplexer3 | 32 | 0.10595 | 0.11094 | 89.52% | 88.84% |
| Multiplexer3 | 64 | 0.0583224 | 0.0583253 | 94.71% | 94.71% |
| Null | 64 | 0.27060 | 0.25876 | 50.00% | 50.00% |

Confirmation parity4 at32 labels gains0.09472 Brier, approximate across-fit
mean +/-3.5SE interval[0.07384,0.11560]. Its20%-excess-risk-adjusted gain
has positive interval[0.05855,0.09149]. At64 labels, raw gain0.02252 has
interval[0.01701,0.02802]. The exact noisy-rule full-input risk floor is0.0475.

The family model is not uniformly beneficial. Majority/multiplexer at32 labels
still have statistically visible small regressions, albeit inside the declared
0.01 tolerance. At64 labels the specialist posterior weight is approximately
0.00885% for majority and0.00630% for multiplexer, versus99.993% for parity4.
This is evidence of finite-model discrimination, not universal truth detection.

## Failed protection cells

Here harm is family-minus-generic Brier. A gate fails when the paired upper
mean-plus3.5SE bound exceeds0.01; failure to certify is not always evidence that
the mean harm itself exceeds0.01.

| Phase/case,16 labels | Full-input mean harm | Upper bound on harm | Partial-view gate |
| --- | ---: | ---: | --- |
| Design majority3 | 0.01729 | 0.03177 | Also fails |
| Design multiplexer3 | 0.01071 | 0.01854 | Passes |
| Confirmation parity1 | 0.00205 | 0.01166 | Passes |
| Confirmation majority3 | 0.00799 | 0.01227 | Passes |
| Confirmation multiplexer3 | 0.01545 | 0.02922 | Also fails |

Do not remove these samples or redefine the success gate. The intervals are
approximate across-fit normal bounds, not anytime confidence sequences. The
models' posterior probabilities themselves are not non-inferiority certificates.

## Still missing

With fixed observed mask63, parity4 at64 labels reaches only54.22% expected
accuracy and Brier0.23112, close to this view's0.23102 oracle risk floor. This
view lacks necessary variables for many sampled target masks. An adaptive
observer was not run, so this is not its measured acquisition performance.

Dependent parity4 at64 labels has full Brier0.04832, but partial Brier0.19574
versus its true partial risk floor0.18355. Family averaging does not repair
the uniform input assumption. Delay, missing feedback, selector publication,
residual validity and actual end-to-end serving are not tested by this component.

## Performance

Apple M4, darwin/arm64, Go1.27.1, benchmark suffix10;500ms, three repeats, same
fixtures for generic and family builders, no concurrent experiment.

| Fit size | Generic time | Family time |
| --- | ---: | ---: |
| 16 | 6.135-6.191ms | 6.456-6.486ms |
| 64 | 6.975-6.985ms | 7.538-7.556ms |
| 256 | 10.187-10.575ms | 11.728-11.751ms |

At64 labels the extra fit cost is approximately0.56ms, about8%. Generic
allocation is~651,264 bytes/3 allocations; family is~667,648 bytes/4 allocations.
Both compile one19683-cell conditional snapshot; temporary fitting arrays cause
the extra allocation. Family lookup is7.048-7.253ns, zero allocations. This is
only a supplied-mask probability lookup, not an MMM traversal or daemon latency.
The exponential nine-bit representation and asymptotic costs from v90 remain.

## Verification and next lead

Generation PASS53.436 seconds. Focused race contracts PASS1.626 seconds; vet and
diff checks pass. Evidence integrals are checked independently with Beta
functions, all19,683 partial mixtures match their component mixture, symmetry
and immutable concurrent reads hold, and input/seed bounds pass. Full3,840-record
v91 replay PASS53.504 seconds. Predecessor v90 replay PASS28.301 seconds.
Independent summaries verify14 v91 source/protocol hashes; both v90 and v91
stored summaries reproduce exactly. Benchmark source/output hashes also pass.

Next: [skeptical-prior sensitivity proposal](../../research/family-prior-rescue-proposal.md).
This remains a proposal. The evidence suggests testing prior influence when
data are sparse, not a confirmed implementation bug or an authorization to
weaken acceptance rules. Keep the generic fallback and future delayed-stream
tests in scope. All seven research directions remain open.

Artifacts: [protocol](mmm-family-v91-protocol.md), [records](mmm-family-v91.json),
[summary](mmm-family-v91-summary.json), [benchmarks](mmm-family-v91-benchmarks.txt),
[benchmark metadata](mmm-family-v91-benchmark-metadata.json).
Raw SHA256: `8aebf1ddccaf8d39b72491e582418fd48a4b610ee79a95d191544d3f4d629797`.
No production/OpenClaw changes, commit or push.
