# Arrival-triggered refitting: exploratory results

The frozen contract is `mmm-spike-cadence-v1-contract.md`. Source index 0,
all 21 scenarios, two phases, two arrival schedules: 84 records, 2,688
forecasts at clocks 128-159 and 2,101 fresh fits. This is a consumed-data
diagnostic, not untouched confirmation or asynchronous serving validation.

| Comparator | Mean expected Brier delta (refit minus comparator) | Records improved | Records harmed by more than .01 |
| --- | ---: | ---: | ---: |
| Frozen32 same model | -.007186895 | 41/84 | 5/84 |
| Generic64 | -.008928246 | 50/84 | 17/84 |
| Boolean64 | -.008693760 | 33/84 | 8/84 |
| Markov incumbent | -.000060372 | 30/84 | 15/84 |

These are descriptive paired differences, not confidence bounds. The average
gain versus Markov is negligible and conceals considerable scenario harm.
Arrival refitting improves all four pooled phase/schedule cells versus the
frozen model, but only two versus Markov. Largest harm versus frozen32 is
.050612632 in phase 1 mux3 with immediate outcomes. Unconditional publication
does not meet the stationary/non-harm requirement. All seven goals remain open.

The independent audit checks source origin availability at every clock,
exact refit/publication decisions, initial equivalence to frozen32, final
factor moments, convergence, evaluator fields and prediction budgets. All
2,101 fits stop by bound-and-state convergence; maximum 113 iterations and
2,366 integrand evaluations per forecast. Maximum reconstructed mean and
variance errors are 8.88e-16 and 3.33e-16. The auditor does not independently
integrate the full 255-factor predictive mixture; separate component tests
cover small-mixture integration. Future-label poisoning and quiet-window
tests pass. The prior checkpoint recorded full spike race PASS (43.915s).

Collection took 22.26s and replay 22.19s, including artifact serialization.
These are research collection durations, not loaded serving latency. The
same-model frozen comparator uses 84 fits; refitting uses 2,101. Comparisons
are not equal-compute. Immediate publication is idealized.

SHA256:

- Raw and replay: `d68e001c9ef4b1f18fbb3662a4a6710b77f1774acda41d5816d7e6cfb6fa4e09`
- Audit: `329fcb82bd3c6c20b12a521bf8c37fc5d416ba7753538842e933c46c7b502df7`
- Summary: `0265f9dad4f74097f9671797c6f74e73869515b2b41748802a4a47949093f06d`
- Cadence test code: `cd9d2eb0c967361fe40d17f443bccda5be8a71d55c1bca3174a96f085cf7ac4f`

Reproduce the paired summary with `node research/spike-cadence-v1-summary.mjs`
followed by the audit JSON path and a new output path. Preserve all artifacts.
Next: test explicit delayed-feedback expert aggregation, not prior tuning.
