# Parity acquisition trace: more than premature stopping

Consumed-data diagnostic, not a new rescue. Replay all32 parity4 trajectories
from the empirical-input integration experiment, preserving original metrics
and all three forecast tapes exactly. Record8192 post-change observations before
their revealing labels. Rerun the chosen observer against the same immutable
reader; require exact final mask, values and acquisition cost.

## Findings

| Post-change quantity | Design | Second cohort |
|---|---:|---:|
| Frames |4096|4096|
| Control observes all four relevant fields |970|1022|
| Coupled candidate observes all four |639|671|
| Control complete, candidate incomplete |340|387|
| Candidate complete, control incomplete |9|36|
| Candidate subset-guided frames |104|232|
| Subset-guided incomplete confidence stops |21|23|
| Subset-guided incomplete budget stops |27|59|

Among the340/387 losses of complete information,225/279 terminate at the full
budget, versus115/108 at confidence. Most budget losses use count-model guides,
not the subset observer:204/231. Therefore disabling the empirical subset
observer's confidence stop cannot by itself account for most lost information.
Forecast changes alter later mixture weights and which model guides retrieval.
This path-level evidence does not isolate the first harmful decision or prove
that a particular alternative observation policy will improve predictions.

Under this known uniform-input noisy-parity generator, missing ANY of the four
relevant bits leaves conditional outcome probability .5; complete information
gives .05 or .95. The information-limited minimum expected Brier therefore rises
from .202045 to .218409 (design), and .199474 to .216827 (second), solely from
changed masks. Expected Brier of the actual forecasts rises .243478->.254920
and .240016->.255799. These oracle calculations are evaluator-only, not signals
available to the learner and not an additive causal decomposition of harm.

## Next Decision

No confidence-threshold or pseudocount tuning. A future rescue must address
budgeted view selection and guide authority jointly, with partial observations
and total cost accounted for. Inspect prior joint-lookahead and acquisition
experiments before proposing a new policy; a known target mask must never enter
the deployed observer. The fixed-observation branch is an important control,
not a substitute for integrated success.

## Verification and Parser Correction

Race-enabled32-trajectory replay passed:16.31s experiment,17.592s package;
vet passed. Summary replay is byte-identical. Missing-mask negative control
is rejected before summary publication. No running processes remain.

The first summary mistakenly read uppercase prediction keys despite the Go
JSON tags being lowercase. That produced invalid null expected losses and
undefined cost counts. `mmm-subset-input-trace-v1-summary.json` is INVALID and
retained only as a failed diagnostic artifact. Use
`mmm-subset-input-trace-v1-summary-corrected.json` and matching
`mmm-subset-input-trace-v1-summary-replay.json`. Raw data was not changed.
The corrected parser validates mask/value/cost/probability shapes before
bitwise operations or arithmetic; this is a project-local lesson, no global
instruction change. No conclusions use the invalid summary.

Raw:`mmm-subset-input-trace-v1.jsonl`.
SHA256:`b0c9e5a5c770a7ed8a49ddebd379ecc51871f6c97af1e15bd424d4dce397a855`.
Parent archive:`24a5a70f080f92a40bd51308a23b051dfec3cec0ef824e0bd2dfede4929749ee`.
Source archive hashes and two trace-source hashes verified.
All seven goals remain open; no production, whitepaper or remote changes.
