# v5 verification notes

Date: 2026-09-12. Scope: research-only learner experiment, not production audit.

## Verified

- Unit/race checks passed before evaluation; vet passed.
- Seeds enumerate disjoint fitting/evaluation domains; confirmation trajectories
  are distinct from design. The same three frozen fitting models are deliberately
  reused across splits, not claimed to be fresh confirmation training samples.
- The journal contains 480 streams of 512 steps and seven paired forecasts.
- All source/protocol hashes in the evidence match the actual files.
- All forecasts, full/post scores, comparisons and verdicts reproduce exactly
  in a complete rerun, excluding nondeterministic fit wall-clock measurements.
- Delayed labels are delivered after the due forecast, never before. Missing
  labels do not reach training, monitoring or mixture weights. Updates use stored
  original expert forecasts. The replay test verifies no duplicate deliveries.
- Independent audits and forest node limits have explicit accounting checks.
- Simple-rule learning and the bounded monitor have separate unit controls.

Command: `go test -v ./internal/observationlearners -count=1`.
TestEvidenceRecomputesAndReplays passed, along with all unit checks.

## Limitations and repair

Recomputation shares scoring code with generation: it verifies artifact
consistency, not an independent statistical proof. Approximate paired intervals
are conditional on the fixed fitting seeds. No confidence-sequence theorem is
claimed for the prototype monitor or repeated tree split checks.

Review confirmed that the initial prediction benchmark discarded its return
value. The benchmark now stores an observable result; its rerun has a separate
filename. No experiment code, protocol, seed, or prediction artifact was changed
by this benchmark-only repair. Microbenchmarks exclude I/O and concurrency.

Known attribution gap: tree and count-model update cadences differ. Known scope
gap: all coordinates are observed, not selected by MMM. These are declared
limitations requiring later experiments, not grounds to relabel v5 as a pass.
