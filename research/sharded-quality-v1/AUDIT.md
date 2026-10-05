# Quality Audit And Limits

Classification: future-ID exclusion is confirmed in the older loaded fixture;
counterfactual packet/law equality was previously missing evidence. Sharding's
effect on ANN/quantization/packing needs investigation, not an assumed benefit.
This experiment adds a bounded observation layer and exact nomination reference;
it changes no runtime policy, storage source, held-out labels or default config.

Confirmed evaluator defect: the first cutoff-only tie calculation let a boundary
tie replace a STRICTLY better required ID without losing recall. A new regression
failed before correction (evaluator-before.txt), then the calculation was fixed:
required IDs above cutoff+1e-6 count individually; boundary ties fill only the
remaining declared slots. Thresholds, raw Go measurements and protocols stay
unchanged. The original V1 evaluation is retained as evaluation-original-metric.json,
not the authoritative corrected report. V2 had no evaluation yet at discovery.

Confirmed verifier entry defect: negative-controls' macOS scratch path used
/var while ESM canonicalized its own filename to /private/var. The literal
main-file guard skipped verification and exited0 with empty stdout, allowing
the first fabricated report (before the expected rejection assertion stopped
the harness). cli-alias-before.json preserves the witness. All three CLI tools
now compare real paths; positive controls MUST emit verified:true, and tampered
reports must produce AssertionError rather than setup errors or empty execution.
The direct /Volumes/Data evaluation and independent arithmetic did execute;
all negative controls are rerun after this repair. No Go measurements change.

Alternative causes: changed graph traversal, score quantization, ties, latent
learner motion, packet confidence modulation or availability filtering. To
separate them, hydrate/recompute all past vectors, compare exact nomination with
the SAME service, use unchanged repeats before/after future captures, retain
native nominations and complete common-ID cold forecasts. Exact-reference
instability would falsify an ANN-only explanation. Actual frontier/packet motion
would falsify strict numerical preservation even with no future ID returned.

The reference is test-only. It never supplies outcomes or trains the ordinary
arm. All public source/expected-answer metadata stays outside the indexed turn;
only DESIGN question text is used. Artificial as-of times test the declared
availability contract, not chronology-faithful factual correctness. Evidence
reaches nomination, durable pre-packing journal/frontier and returned packet;
there is no agent/model invocation or proof of improved downstream answers.

The ordinary and exact services are COLD: belief/residual/expert flags must stay
inactive. The comparison reports all three Bernoulli branches and full point/
forecast semantics on common IDs, not a joint probability law over changing
candidate sets. No outcomes means no empirical Brier/calibration claim. The
2|delta-p| Brier bound is an elementary worst-case bound for a COMMON binary
forecast, not a scored experiment or a bound for dropped/gained candidates.

Tie tolerance1e-6, finite mean screens and strict equality1e-9 were frozen before
collection. Strict ID recall/overlap remains visible; tied substitutions are not
called errors in exact-score quality. No confidence interval or confidence
sequence is claimed for these fixed96 design queries. Vectors use the hash
development embedder, not production semantic embeddings. No learned-layer
quality, reference instantiation of the full paper, or whole-goal closure.

Confirmed implementation limit after V1: HNSW's exactCutoff is max(2*EfConstruction,
EfSearch,k). With400construction floor and ~250records/shard, V1's candidate
uses exact local scans, not large-index ANN traversal. The separately frozen
V2 scales to4000past and source-derived1000records/shard so nominal150/600 calls
leave that fallback. Expanded availability probes can still return to exact;
actual branch counts are not instrumented. Do not attribute results to sharding
alone independent of search effort/fallback policy.

Quiet timing is descriptive and interleaved with exact reference work; it is
NOT the matched loaded pilot, nor a quiet non-harm screen. Candidate fan-out can
cost more without capture contention. Source pins cover declared inherited
files and new harness; a complete compiler-derived source closure is not
prospectively frozen in this study. No production/cache/global instruction edits.
