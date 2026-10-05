# Full-budget v17: targeted gain, overall FAILED

1152 fresh streams,589,824 frames. The majority stopping hypothesis passes its
finite mean-gain checks in both splits, but the overall protocol fails two
multiplexer gain checks versus fixed counts. No mean-harm check fails. Do not
report this as a universal rescue or change the criteria after seeing results.

Clustered shift128 post Brier and observation cost:

| Family/split | Current stop | Full budget | Current mean coordinates | Full coordinates |
| --- | --- | --- | --- | --- |
| Majority design | .079964 | .074792 | 2.36 | 6 |
| Majority confirmation | .086478 | .080902 | 2.45 | 6 |
| Multiplexer design | .111324 | .110393 | 3.54 | 6 |
| Multiplexer confirmation | .110713 | .108794 | 3.39 | 6 |

Majority confirmation paired difference is -.005576, descriptive six-fit95%
interval[-.010609,-.001481]. The point gain meets .005, but that interval does
not establish a population gain of at least .005. Extra information costs about
2.44 times as many acquired coordinates in that cell, not zero runtime overhead.

Multiplexer confirmation improves only .003729 versus fixed counts, below .005;
its improvement versus current stopping is .001919 with interval spanning zero.
This matches v16's warning that multiplexer was mainly forecast-limited rather
than observation-limited. No result here identifies a universal stopping rule.

## Verification

Original-stop mode matches the previous subset64 runner exactly. Unchanged
arms0/1/3 and input/outcome/delivery sequences remain paired. The full-budget arm
spends exactly6 on complete synthetic readers. All1152 records replay exactly
apart from timing (78.27s). Streaming metric/source/mask/delay audit, targeted
race tests and command vet pass. Production remains unchanged.

Artifacts: [protocol](mmm-fullbudget-v17-protocol.md),
[raw](mmm-fullbudget-v17.jsonl.gz), [summary](mmm-fullbudget-v17-summary.json).

```sh
go run ./cmd/eventframe-observation-fullbudget NEW.jsonl.gz
python3 research/fullbudget_summary.py NEW.jsonl.gz NEW-summary.json
EVENTFRAME_FULLBUDGET_ARTIFACT=NEW.jsonl.gz go test ./internal/observationlearners -run '^TestFullBudgetArtifactReplay$' -count=1
```

## Next Lead

Seek selective stopping that preserves the majority information gain without
always spending6. First check whether a confident observation guide stops while
the emitted mixture remains uncertain. That mismatch would motivate testing
mixture-aware stopping, not unconditional extra reads. Independently refine
forecast quality for multiplexer and the real-task representation; those gaps
cannot be declared solved by this observation result. All seven directions remain
in scope, with no deployment or broad claim promotion.
