# Prequential reliability v20 results

**Outcome: FAILED the predeclared diagnostic screen.** This is a replay of
consumed v19 traces, not a fresh confirmation or a tested online intervention.

The [protocol](mmm-calibration-v20-protocol.md) uses only previously delivered
labels, separately for each predicted class. The last 64 confidence-stop
outcomes supply a smoothed working reliability estimate; at least 32 are
required. A warning means empirical correctness is more than .03 below mean
nominal confidence. It is not a correctness certificate or confidence sequence.

## Target result

Clustered majority, shift128, confirmation-labelled partition, post-change:

| Quantity | Observed | Required |
| --- | ---: | ---: |
| Warning fraction | 1817/4264 = 42.61% | At most 50% |
| Errors captured | 145/361 = 40.17% | At least 50% |
| Warned error rate | 145/1817 = 7.98% | At least twice nonwarned |
| Nonwarned error rate | 216/2447 = 8.83% | Comparator |

The warning budget passes, but both targeting requirements fail. Warnings do
not enrich for errors in this target. No threshold retuning or production
promotion follows from this consumed-data diagnostic.

The [machine-readable artifact](mmm-calibration-v20.json) retains all 576
mixture-stop records' aggregate groups, including stationary and delayed/missing
feedback scenarios. In clustered majority delayed/missing confirmation Post,
the warned error rate is 68/897 = 7.58%, versus 176/1899 = 9.27% nonwarned.
This also does not support interpreting warnings as a useful error locator.

## Verification and interpretation

Two unit tests cover bounded history, minimum support, class separation, and
decision timing before current or future label delivery. A complete replay
checks all aggregate outputs and the embedded analysis source against the saved
artifact; the input artifact's embedded source hashes are also checked.

No online forecast or observation policy changed, so this experiment establishes
neither Brier improvement nor a serving-latency result. It tests error targeting,
not the value of extra observations after a warning. Serial dependence and
already-consumed data preclude interpreting these fractions as fresh population
guarantees.

The narrow next question is whether *context-specific* observable structure can
predict the value of an additional read. Recent class-level reliability alone
is unsupported for that purpose. Any such replacement needs a separately frozen
policy and fresh streams; delayed feedback, selection, and real-task transfer
remain required checks. The seven-direction research objective remains open.
