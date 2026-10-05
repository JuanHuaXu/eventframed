# Mixture-aware stopping v19: modest gain, screen FAILED

1728 fresh streams,884,736 frames. The mixture gate fails five gain checks:
majority clustered shift128 versus current and fixed in both splits, plus
multiplexer clustered design versus fixed. No mean-harm or cost criterion fails.
Thresholds were not changed after evaluation.

Clustered shift128 confirmation, adaptive arm:

| Family | Current Brier | Mixture-stop Brier | Full-budget Brier | Current reads | Mixture-stop reads |
| --- | --- | --- | --- | --- | --- |
| Majority | .084722 | .081799 | .079833 | 2.42 | 2.53 |
| Multiplexer | .108455 | .107387 | .108211 | 3.71 | 3.92 |

Majority mixture-minus-current difference is -.002922, descriptive six-fit95%
interval[-.004911,-.001532]. This is a promising modest benefit, but below the
predeclared .005 magnitude. It uses about4.3% more acquired coordinates versus
current, rather than always spending6. These coordinate costs are not measured
service latency. Multiplexer difference is -.001068 with interval spanning zero.

Even full-budget majority gain is .004889 in this fresh confirmation cell,
slightly below .005. V17's targeted mean pass therefore should not be read as a
robust minimum-gain guarantee. Preserve both experiments and their uncertainty.

## Verification

Always-accept and always-deny callbacks reproduce original and full-budget
observation results. Nil/error gates and snapshot changes reject. Every gated
confidence stop matches the actual emitted mixture's confidence region. Controls,
unchanged arms0/1/3 and delayed feedback remain paired. All1728 records replay
exactly apart from timing (127.16s). Source/metric/mask/cost audits, targeted race
tests and command vet pass. An initial experiment-adapter naming error failed
compilation and was corrected before any artifact run.

Artifacts: [protocol](mmm-mixturestop-v19-protocol.md),
[raw](mmm-mixturestop-v19.jsonl.gz), [summary](mmm-mixturestop-v19-summary.json).

```sh
go run ./cmd/eventframe-observation-mixturestop NEW.jsonl.gz
python3 research/mixturestop_summary.py NEW.jsonl.gz NEW-summary.json
EVENTFRAME_MIXTURESTOP_ARTIFACT=NEW.jsonl.gz go test ./internal/observationlearners -run '^TestMixtureStopArtifactReplay$' -count=1
```

## Next Questions

Retain this as a low-cost research candidate, not the completed rescue. V18's
confident shared mistakes remain outside its trigger. Investigate a prequential
calibration check using only already-realized, journaled predictions before
altering stopping again. Separately test real-task feature representations and
actual worker integration; synthetic stopping gains cannot discharge those
requirements. All seven roadmap directions remain open and production unchanged.
