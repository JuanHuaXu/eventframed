# Preserved-incumbent MMM v3 audit

## Scope and invariant

The v2 failure was traced to surprise-driven revocation and replacement of a
well-supported incumbent with a small fit. v3 is a separately frozen research
composition, not a patch to those historical results or production defaults.
The invariant is that sharing review cannot erase the incumbent or allow an
outcome to modify its own forecast. Competing explanations remain tested via
no-AP, local-only, breadth and random controls.

## Verification

- Unit tests and race checks passed for the new package and command before
  running design and confirmation. They cover disjoint seed domains, bounded
  two-sided sequential evidence, observed-mask validation, preservation of the
  incumbent, duplicate/out-of-order feedback, and next-step fitting visibility.
- Every recorded score and mixture forecast recomputed from raw traces.
  Observed masks, six-coordinate budget, fit/audit timing and sharing transitions
  were checked. Source/protocol hashes match the frozen implementation.
- All 320 streams and six policy histories replayed deterministically; equality
  excludes only measured fitting wall time. This full replay ran without the race
  detector; the separate unit/one-stream race check passed before evaluation.
- All 100 aggregate contrasts and verdict flags reproduced. The primary pass is
  not substituted for the unresolved AP ablation or the small stationary harm.
- No production package imports the new research state. The one additive helper
  in the existing observation prototype exposes only validated observed-mask
  forecasting; v1/v2 hashed source files were not rewritten.
- Three serial microbenchmark repetitions completed after experiment and replay,
  with no concurrent test/build workload. `go vet` passed for the new package
  and command. Timing exclusions are recorded alongside results.

## Limits

Small fixed-schema generator, one fitting seed, supplied temporal views, full
outcome availability, independent reference stream, and only 32 confirmation
streams per scenario. Complete audit pairs are assumed; no missing-outcome or
delayed-feedback experiment. Weights are performance weights of adaptive expert
algorithms, not truth probabilities. The sharing test addresses a reliability
proxy conditional-mean null, not the whitepaper's full target-law diameter.
There is no source-provenance, causal or production-publication validation here.
The synthetic command is not a live asynchronous scheduler or end-to-end test.
