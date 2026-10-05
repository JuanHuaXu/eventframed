# Broader subset validation v14: FAILED

1152 streams,589,824 frames; majority and multiplexer targets replace the
previous label family. Six fitting groups per comparison, two streams per fit.
Both retained variants fail the predeclared screen. The successful v13 pilot is
preserved, but does not establish broad robustness.

Clustered shift128 confirmation, adaptive post-change Brier:

| Family | Fixed count | Retained forest | Retained subset |
| --- | --- | --- | --- |
| Majority | 0.090530 | 0.090444 | 0.088714 |
| Multiplexer | 0.115250 | 0.116188 | 0.112476 |

Subset-minus-forest paired differences and descriptive fitting-group95% intervals:

- Majority: -.001730 [-.003369,-.000440].
- Multiplexer: -.003712 [-.007093,-.001097].

These suggest a benefit on the sampled fits, but neither observed gain reaches
the frozen .005 requirement. Six-group bootstrap intervals are not simultaneous
coverage certificates. They do not justify redefining the threshold after seeing
results or treating589,824 frames as independent fitting replications.

Adaptive retention has eight failed gain checks and no failed mean-harm checks.
Static retention has eight failed gain checks and four forest-harm failures.
The latter all involve biased inputs with delayed/missing feedback, post window:

| Family | Design harm | Confirmation harm |
| --- | --- | --- |
| Majority | .015185 | .013950 |
| Multiplexer | .011677 | .017969 |

Failures count grouped criteria, not independent experiments. All full/post
metrics and per-fit means remain in the summary. Static substitution is therefore
not a supported universal rescue. Adaptive retention is the more credible branch,
but still lacks the required improvement magnitude on correlated new targets.

## Verification

Truth-table tests cover all512 inputs for both families across change boundaries.
Seed tests exclude cross-family/split collisions. All1152 records replay exactly
apart from timing. Streaming audit verifies raw metric reconstruction, pairings,
budgets, observed fields, source hashes and delayed-label availability. Targeted
race tests and command vet pass. Compressed JSON-lines avoids loading the full
trace corpus into memory; first line contains source/header metadata.

Artifacts: [protocol](mmm-breadth-v14-protocol.md),
[raw](mmm-breadth-v14.jsonl.gz), [summary](mmm-breadth-v14-summary.json).

```sh
go run ./cmd/eventframe-observation-breadth NEW.jsonl.gz
python3 research/breadth_summary.py NEW.jsonl.gz NEW-summary.json
EVENTFRAME_BREADTH_ARTIFACT=NEW.jsonl.gz go test ./internal/observationlearners -run '^TestBreadthArtifactReplay$' -count=1
```

## Next Discriminator

Return to window adaptation (direction2) with the stronger probabilistic
challenger, preserving incumbent/count and forest evidence rather than forcing
static replacement. Separate insufficient recent-data adaptation from an
intrinsically weak predictor, with an unchanged-label-volume control. This is a
new lead, not an assertion that window changes will rescue the failures.
Real-task representation, agent outcomes and loaded shadow integration remain
independently necessary; all seven directions remain active.
