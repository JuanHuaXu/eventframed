# Active research checkpoint: V33 and V34 complete

2026-10-03. Goal ACTIVE; all seven whole directions OPEN. Weekly usage
last read2%, below >80%. No stop condition or blocker. Previous goal turn
was progress (V33 preparation/design); this continuation is progress:
completed V33 confirmation/audits, implemented and tested a new coherent
repeated-trial dispersion model, and completed V34 fresh controlled splits.
No live experiment, benchmark or test remains at this checkpoint.

## Authoritative State

- [V33 result](mmm-curve-v33-results.md):1,536 worlds,53,760 snapshots;
  fixed-model scarcity-only rescue FAIL19/24 design and16/24 confirmation.
  More labels repair much curved calibration, not every frozen gate.
- [V34 result](mmm-dispersion-v34-results.md):1,536 DIFFERENT worlds,
  38,400 snapshots,3,686,400 conditional trials; overall component
  FAIL6/24 design and5/24 confirmation. Aligned/reversed low-variance
  forecast improvements pass, Beta2 protection survives, distribution-shape
  calibration and nonlinear/protection limits remain.
- V33 independent-language tape/math auditor, exact both-split replay,
  eight negative controls, no-future/nonmutating snapshots pass.
- V34 exact both-split RNG/issued-law replay and separate batch beta
  integrals pass; six negative tape controls plus source control,
  no-future/underflow/atomicity and four-module race/vet pass.

V34 model/tests/collector/protocol hashes stay frozen after collection.
Its post-collection auditor hash is recorded separately in both audit JSONs.
Earlier V31/V32/V33 tapes, checks and gates remain unchanged. No consumed
cohort threshold/prior sweep or oracle forecast clipping was performed.

## Full Objective Audit

1. More generators, noise structures and repeated observations are covered;
   temporal/delayed evidence and broader task validation remain incomplete.
2. These stationary studies do not prove adaptive-window recovery.
3. Neither posterior grants AP/source authority or closes split usefulness.
4. Dispersion/LOO components and costs advance the challenger, but broad
   quality gates fail and actual shifted sample efficiency is not established.
5. No untouched outcome-labeled agent run here; the local generation path
   last inspected had an embedding model only. Revalidate before using it.
6. Microbenchmarks/model phase sums do not prove loaded durable freshness
   or serving latency. Existing lifecycle results remain separate evidence.
7. Matched labels do not equalize total compute/acquisition cost, and no
   falsification-over-random/uncertainty result is claimed here.

## Commands

Workspace `<LOCAL_ROOT>`; stale app cwd is not authoritative.
Apple M4, Go1.27.1 darwin/arm64, Nodev26.8.1. Run from repository root:

```sh
node research/curve-v33-verify.mjs
node research/curve-v33-audit-negative.mjs
EVENTFRAME_CURVE_V33_REPLAY=<LOCAL_ROOT>/docs/experiments/mmm-curve-v33-design.jsonl go test ./internal/researchblend -run '^TestCurveReplayV33$' -count=1 -v
EVENTFRAME_CURVE_V33_REPLAY=<LOCAL_ROOT>/docs/experiments/mmm-curve-v33-confirmation.jsonl go test ./internal/researchblend -run '^TestCurveReplayV33$' -count=1 -v
EVENTFRAME_DISPERSION_V34_AUDIT=<LOCAL_ROOT>/docs/experiments/mmm-dispersion-v34-design.jsonl go test ./internal/researchdispersion -run '^TestStudyAuditV34$' -count=1 -v
EVENTFRAME_DISPERSION_V34_AUDIT=<LOCAL_ROOT>/docs/experiments/mmm-dispersion-v34-confirmation.jsonl go test ./internal/researchdispersion -run '^TestStudyAuditV34$' -count=1 -v
go test -race ./internal/researchdispersion ./internal/researchblend ./internal/researchcalibration ./internal/researchpartition -count=1
go vet ./internal/researchdispersion ./internal/researchblend ./internal/researchcalibration ./internal/researchpartition
go test ./internal/researchdispersion -run '^$' -bench . -benchmem -count=3
```

Next investigate coherent member-distribution shapes and nonlinear mean
structure on fresh protocols, preserving fixed2/dispersion controls. Do not
reinterpret a learned variance as a learned full distribution, or repeated
transcript processing as new independent trials. Then broaden delayed/shifted
integration and return to actual-agent/loaded-service work; the seven-goal
objective is not replaced by synthetic component success.

Six preexisting tracked modifications remain unchanged (205 insertions,
5 deletions). New work is isolated/local. No production deployment, install,
whitepaper edit, commit or push was performed by this continuation.
