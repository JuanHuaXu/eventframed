# Active research checkpoint: V31 and V32 complete

2026-10-03. Goal ACTIVE; no stop condition reached. Last weekly usage
readback1%, below the >80% stopping threshold. All seven whole goals OPEN.
Previous research goal turn: progress (V31 candidate/protocol/auditor
became runnable). This goal turn: progress (two new controlled studies,
verified negatives and a tested exact LOO replacement). No live experiment,
test or benchmark process remains at this checkpoint.

## Authoritative Outcomes

- [V31 report](mmm-stack-v31-results.md):1,536 worlds/23,040 arms;
  overall FAIL20/24 design and21/24 confirmation cells.
- [V32 report](mmm-crossfit-v32-results.md):1,536 DIFFERENT worlds/26,112
  arms; overall FAIL20/24 design and21/24 confirmation cells. Tight-curved
  confirmation affine-relative forecast/retrieval improvements are
  component evidence only; packed calibration and protection still fail.
- Both independent full tape reconstructions, both-split exact Go replays,
  negative auditor controls, module race tests and vet pass. All sealed
  hashes remain unchanged after reporting.

V31 implementation uses retained original issued rows and bounded fixed-size
incremental weight fitting. V32 uses exact omitted-own-label rows over
currently arrived evidence; O(N+LH) per refit and O(LN+L^2 H) cumulatively,
with N<=200 and H=54. Do not describe the latter as constant-time streaming.
No future data, known true rate, consumed-cohort prior/threshold retuning,
production deployment or whitepaper change is part of this checkpoint.

## Scope Audit

1. Twelve fixed-world families broaden synthetic evaluation, but changing
   regimes, delayed evidence and agent/task generality remain incomplete.
2. Neither study is a temporal adaptive-window recovery experiment.
3. Neither supplies new valid AP authority or useful split integration.
4. Exact LOO arithmetic and measured limits advance component work; broad
   quality/calibration gates still fail.
5. No untouched outcome-labeled agent experiment was performed here.
6. Offline model timings are not loaded durable-learning freshness/latency.
7. Fixed32 labels do not equalize total acquisition and compute cost.

## Replay and Next Action

Host: Apple M4; Go1.27.1 darwin/arm64; Nodev26.8.1. Workspace is
`<LOCAL_ROOT>`, not the stale app cwd. Run audits from that root:

```sh
node research/stack-v31-verify.mjs
node research/crossfit-v32-verify.mjs
node research/stack-v31-audit-negative.mjs
node research/crossfit-v32-audit-negative.mjs
EVENTFRAME_STACK_V31_REPLAY=<LOCAL_ROOT>/docs/experiments/mmm-stack-v31-design.jsonl go test ./internal/researchblend -run '^TestStackReplayV31$' -count=1 -v
EVENTFRAME_STACK_V31_REPLAY=<LOCAL_ROOT>/docs/experiments/mmm-stack-v31-confirmation.jsonl go test ./internal/researchblend -run '^TestStackReplayV31$' -count=1 -v
EVENTFRAME_CROSSFIT_V32_REPLAY=<LOCAL_ROOT>/docs/experiments/mmm-crossfit-v32-design.jsonl go test ./internal/researchblend -run '^TestCrossfitReplayV32$' -count=1 -v
EVENTFRAME_CROSSFIT_V32_REPLAY=<LOCAL_ROOT>/docs/experiments/mmm-crossfit-v32-confirmation.jsonl go test ./internal/researchblend -run '^TestCrossfitReplayV32$' -count=1 -v
go test -race ./internal/researchblend ./internal/researchcalibration ./internal/researchpartition -count=1
go vet ./internal/researchblend ./internal/researchcalibration ./internal/researchpartition
```

Next prepare/freeze matched label-budget curves on new seeds, retaining
all model/regularizer constants. Check nomination beyond32 carefully:
the old fixed32 `stratumV30` helper is not a multi-pass selector; use
proper unseen-member bookkeeping rather than repeating raw bucket draws.
Separate persistent calibration failure from finite-label scarcity before
adding another learner. Preserve both studies' negatives. Later broaden
temporal/delayed recovery and loaded/agent evidence; do not replace the
full seven-goal objective with this synthetic subproblem.

Six preexisting tracked modifications are preserved. All new research is
local and isolated. No commit, push, installation or production change
was made or authorized by this goal continuation.
