# Known-law diagnostic v16

Analyzed the1728 consumed v15 records without refitting or changing predictions.
This is an offline known-generator/regime reference, not a deployable model or
an independent confirmation run.

Subset64 adaptive arm, clustered shift128 confirmation, post window:

| Family | Noise floor | Observation deficit | Forecast deficit | Expected Brier | Realized Brier |
| --- | --- | --- | --- | --- | --- |
| Majority | .047500 | .027938 | .009694 | .085133 | .081209 |
| Multiplexer | .047500 | .018817 | .045178 | .111494 | .114208 |

The first three columns sum to conditional expected Brier. Realized scores
need not equal that expectation on a finite sample. Majority's observation term
accounts for about74% of excess expected error; multiplexer forecast deficit
accounts for about71%. Forecast deficit combines estimation, calibration,
misspecification and unavailable regime knowledge, not one identified cause.

## Stopping Breakdown

For majority,4444/4608 post frames stop for confidence, versus164 at the budget.
Their observation-deficit sums are128.42374 and .31678: about99.8% of this
deficit lies on confidence-stopped frames. Early stops use2.24 coordinates on
average. Only34 of those frames read all three relevant coordinates, although
reading all three is not always necessary to decide a majority.

For multiplexer,3074 frames stop for confidence and1534 at budget. The budget
frames carry forecast-deficit sum167.57127 versus40.60859 on early stops. More
observations alone therefore cannot be assumed to address that family's main
problem. These stop-reason totals pool frames; the primary table averages
per-stream risks equally.

This supports testing early stopping as a specific majority limitation. It does
not show that forcing extra reads improves learned forecasts: smaller conditional
cells, poor variable choices and uncalibrated estimates could offset information
gain. Test a full-budget control against current stopping on fresh streams, with
the same six-coordinate cap and explicit realized observation costs. Preserve
the known-law reference solely as a diagnostic.

## Verification

Tests cover all complete-input probabilities, correlated conditional examples
and the squared-loss decomposition identity. Streaming analysis verifies source
hashes, observed masks and reconstruction of original realized scores. The
diagnostic source is embedded in the main output; both outputs bind the v15 hash.
No prior experiment or production behavior changed.

Artifacts: [protocol](mmm-error-budget-v16-protocol.md),
[error budgets](mmm-error-budget-v16.json),
[stopping totals](mmm-error-budget-v16-stopping.json).

```sh
python3 -m unittest discover -s research -p test_error_budget.py
python3 research/error_budget.py docs/experiments/mmm-windowbank-v15.jsonl.gz NEW.json
python3 research/stopping_budget.py docs/experiments/mmm-windowbank-v15.jsonl.gz NEW-stopping.json
```

All seven research directions remain open. This diagnostic narrows an acquisition
lead; it does not validate a new policy or real-world learning claim.
