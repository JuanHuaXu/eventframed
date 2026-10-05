# Online distribution-aware acquisition v11: FAILED

576 paired streams, 294,912 frames across uniform, empirical and diagnostic
oracle modes. Neither empirical retained arm passed the frozen advancement rule.
Both have four failed gain checks, all on clustered shift128 (versus fixed and
uniform, in both splits). No empirical group exceeded the .01 mean-harm limit.

Clustered confirmation, post-change Brier:

| Arm | Uniform acquisition | Empirical acquisition | Oracle acquisition |
| --- | --- | --- | --- |
| Fixed count control | 0.133407 | 0.133407 | 0.133407 |
| Adaptive retained | 0.131980 | 0.131695 | 0.131719 |
| Static retained | 0.136740 | 0.136448 | 0.136253 |

The empirical gain versus uniform is .000286 for adaptive and .000292 for
static retention, not the required .005. Oracle does not supply a large rescue.
These are fresh streams, not a relabeling of v10 fixed-mask diagnostics. The
four policies in each mode use identical input/outcome/audit sequences.

## Path Diagnostic

Post-hoc comparison (not an additional success criterion) shows limited path
movement on clustered shift128. Adaptive empirical differs from uniform on
67/4096 confirmation paths and38 final masks; design differs on only3/4096
paths and masks. Static empirical differs on10 confirmation paths and63 design
paths. These counts exclude probability-only differences.

The change is wired and does affect observations, but usually leaves the path
unchanged. That alone does not prove the mixture rarely selects the tree:
different guides can choose identical views. Guide-selection frequency needs
direct reconstruction before changing the gating policy. Static retention
explicitly uses the count guide for its tied inner bundle, so its path changes
arise indirectly through outer mixture selection, not direct tree guidance.

A defensible next discriminator is to measure which expert actually selects each
view and compare a shared information-value acquisition rule with the existing
winner-guided rule. Forcing tree use without controls could simply discard the
incumbent's useful observation policy. This lead remains untested; no production
adoption follows from small non-harmful mean changes.

## Verification

- Mode0 exactly matches the v9 runner on a common stream.
- Uniform joint observation policy matches all512 inputs' original forest
  acquisition paths on a fitted test model, with numerical forecast tolerance.
- Cross-mode fixed-count forecasts and input/outcome/audit sequences match.
- All576 records replay exactly excluding timing instrumentation.
- Raw audit recomputes Brier and accuracy, verifies source hashes, observed
  values, budgets, forecast bounds and delivery ordering.
- Targeted race tests and command vet pass. Original v9/v10 sources unchanged.

Artifacts: [protocol](mmm-acquisition-v11-protocol.md),
[raw](mmm-acquisition-v11.json.gz), [summary](mmm-acquisition-v11-summary.json),
[post-hoc paths](mmm-acquisition-v11-paths.json).

```sh
go run ./cmd/eventframe-observation-acquisition NEW.json.gz
python3 research/acquisition_summary.py NEW.json.gz NEW-summary.json
EVENTFRAME_ACQUISITION_ARTIFACT=NEW.json.gz go test ./internal/observationlearners -run '^TestAcquisitionArtifactReplay$' -count=1
```

Pilot mean screens do not prove population non-inferiority or calibrated
forecasts. Real agent outcomes, other label families and integrated shadow
learning remain outstanding. The overall research objective remains active.
