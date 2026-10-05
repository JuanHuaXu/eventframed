# Alternate-path diagnostic v12: FAILED

192 consumed v9 streams;73,736 supported-fit frames. Each expert chooses its own
observations under the same six-coordinate cap, before outcome feedback.

Clustered shift128, supported-fit post-window Brier:

| Split | Short count | Uniform tree | Empirical tree | Oracle tree |
| --- | --- | --- | --- | --- |
| Design | 0.133472 | 0.177770 | 0.177870 | 0.177770 |
| Confirmation | 0.140166 | 0.194349 | 0.194354 | 0.194349 |

Confirmation accuracy is80.60% for short counts and73.20% for the empirical
tree. Mean observation costs are4.97 and6.00 coordinates respectively. On these
own paths the tree is worse, not merely suppressed by the mixture's choices.

The empirical tree fails32 full/post harm checks plus both clustered shift
gain checks. These are overlapping grouped failures, not34 independent trials.
No criterion changed. All group metrics, including accuracy and cost, remain
in the summary; the table above is not the sole evidence.

This deprioritizes forcing the existing bounded forest into control. It does
not falsify every forest architecture, every sample budget or general MMM.
The trees may be intrinsically weak under64 audit labels and the present fitting
rule. A different small-sample probabilistic challenger is an open lead. V12 is
not independent confirmation and cannot replace a fresh prospective test.

## Verification

Original tree forecasts match the v9 journals before alternate paths are scored.
Tests check own-view values against synthetic hidden inputs, coordinate budgets,
forecast consistency and final-outcome non-leakage. Targeted race test and command
vet pass. The raw artifact binds v9's hash and retains diagnostic source text.
Summary reconstructs scores using equal per-stream weights. No production path
or old artifact was modified.

Artifacts: [protocol](mmm-alternate-v12-protocol.md),
[raw](mmm-alternate-v12.json.gz), [summary](mmm-alternate-v12-summary.json).

```sh
go run ./cmd/eventframe-observation-alternate docs/experiments/mmm-dependent-v9.json.gz NEW.json.gz
python3 research/alternate_summary.py NEW.json.gz NEW-summary.json
go test -race ./internal/observationlearners -run '^TestAlternatePaths$' -count=1
```
