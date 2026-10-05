# Retained window banks v15: FAILED

1728 streams,884,736 frames, six fitting groups. Both banks fail eight gain
criteria (four versus fixed counts, four versus subset64); no mean-harm
criterion fails. No threshold or prior was changed after the run.

Clustered shift128 confirmation, adaptive arm post Brier:

| Family | Fixed count | Subset64 | Window bank | Forest-retaining bank |
| --- | --- | --- | --- | --- |
| Majority | 0.081107 | 0.081209 | 0.081159 | 0.081497 |
| Multiplexer | 0.118639 | 0.114208 | 0.113979 | 0.114169 |

Versus subset64, multiplexer differences are -.000229 (window bank) and
-.000040 (retaining bank). Their descriptive fitting-group95% intervals are
[-.001584,.001092] and[-.001304,.001250], respectively. These are not established
improvements and are far below the required .005. Majority retaining-bank
performance is slightly worse, although within the .01 allowance.

This does not establish that every adaptive window is ineffective. It shows
that adding this16-label subset expert to64-label retention, with the frozen
online weighting and observation policy, is not a sufficient rescue. Preserving
the forest likewise does not supply the missing gain in this experiment.
These modes add fitting work; lack of measured quality benefit argues against
adoption rather than assuming additional complexity is harmless.

## Verification

Exact subset64 control, unchanged outer arms0/1/3, paired inputs/outcomes and
delayed feedback, duplicate-weight guide selection, targeted race tests and
command vet pass. Streaming audit reconstructs metrics, checks source hashes,
field budgets and label timing. All1728 records replay exactly apart from timing
instrumentation (replay test passed in170.51s).

Artifacts: [protocol](mmm-windowbank-v15-protocol.md),
[raw](mmm-windowbank-v15.jsonl.gz), [summary](mmm-windowbank-v15-summary.json).

```sh
go run ./cmd/eventframe-observation-windowbank NEW.jsonl.gz
python3 research/windowbank_summary.py NEW.jsonl.gz NEW-summary.json
EVENTFRAME_WINDOWBANK_ARTIFACT=NEW.jsonl.gz go test ./internal/observationlearners -run '^TestWindowBankArtifactReplay$' -count=1
```

## Implication for Research

The strongest remaining model candidate is still adaptive subset64 retention,
with its v13 success and v14 limitations intact. Before adding more window
variants, distinguish finite-label learning error from observation-policy error
against an explicitly nondeployable known-law reference. Separately move the
best current candidates into real-task representation tests; synthetic tuning
cannot resolve the existing semantic-retrieval regression. Directions3,5,6,7
also retain their independent unfinished requirements. No production change.
