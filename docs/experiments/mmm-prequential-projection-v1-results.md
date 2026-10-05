# Prequential error evidence: component boundary passes

This completes the [projection protocol](mmm-prequential-projection-protocol.md)
before fitting an error-informed critic. It does not establish acquisition
efficacy, statistical calibration, or completion of any whole research goal.

## Evidence chain

The source stores channel10 forecasts from the64-label segmentation model before
reading the current frame's label. For origins128..159, those predictions come
from its128-clock fit. The new Go verifier reconstructs that fit independently
of the saved prediction table using the batched fitter and the correct as-of
label support, then looks up all32 issued inputs.

All2688 records and86016 issued probabilities pass. Maximum discrepancy from
the source's saved predictions is1.8108e-13, below the frozen1e-10 tolerance.
Twelve poisoned-fit fixtures give unchanged forecasts when excluded outcomes
are flipped. The twelve-fixture race suite checks384 issued values and passes
in39.847s package time. Full replay takes34.91s wall,136.24s user,1.61s system
with four workers; it is verification work, not serving latency.

## Projection checks

Across all2688 records:

- 86016 issued records are projected, but only received labels are exported.
- Complete-delivery runs contain43008 observed labels,32 per record, and no
  unresolved candidates in the recent pool.
- Delayed runs contain18826 observed labels,7-22 per record, and9277 candidate
  queries in total. Coverage accompanies all error summaries.
- All92770 candidate feature values exactly match a separately implemented
  reference calculation on this input (maximum arithmetic discrepancy0).
- 5376 whole-record poisoning/delivery-equivalence checks pass: future/unarrived
  Y, all Q, future X/predictions, and identity metadata do not affect the view.
- Changing an actually received label changes the features in all1344 delayed
  runs tested, ruling out a vacuous always-constant projection.
- Hand-calculated, empty-view, exact arrival-boundary, strict field-access,
  ownership, invalid-view and duplicate-label tests pass.
- Projection replay is byte-identical. `go vet ./internal/observationlearners`
  passes. Source and code hashes are recorded in the artifacts.

The adapter uses retrospective delivery metadata only to reconstruct receipt
status, then removes it. A runtime implementation would need actual receipt
records. It must not let a policy distinguish an unresolved permanently missing
packet from a packet that will arrive later based on hidden generator metadata.

These errors describe an auxiliary previously issued model, not the current
refitted query law. They are observed-feedback diagnostics, not unbiased risk
estimates under arbitrary selective/missing feedback. The Hamming locality and
empty-evidence defaults are declared research choices, not validated semantics
for general text/vector memory.

## Next experiment

The planned state-information versus nonlinear-model factorial can now use
these bounded inputs without relying solely on an assumed as-of boundary.
Freeze its model capacities and comparisons before evaluation, retain all
random/entropy controls and query costs, and keep generator truth out of inference.
No acquisition improvement is claimed until that test runs.

Artifacts: `mmm-prequential-projection-v1.json`,
`mmm-prequential-projection-v1-replay.json`, `mmm-prequential-issued-v1.json`,
`mmm-prequential-issued-contracts.txt`, and `mmm-prequential-issued-v1-run.txt`
in this directory. All seven goals remain open; production and paper unchanged.
