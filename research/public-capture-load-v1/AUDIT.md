# Scope, bug and leakage audit

This is a parent-led scoped review, not an independent research review. Source
and protocol hashes are frozen per comparison. Production files, private
session corpora and confirmation labels are untouched.

## Publication Boundary

This public edition preserves the timing data, protocols, test transcripts and
summarized profiler reports from local checkpoint `f7b7bb1`. The three raw
`.pprof` captures remain local; `PUBLICATION.json` retains their original hashes.
The public verifier checks the available reports but explicitly reports that it
cannot independently verify or regenerate those raw profiles from this edition.
No private session data or unfinished shard-rescue changes are included. This
publishes research evidence, not a production promotion or deployment.

## Data boundary

The fixture decodes only capture envelopes and optional question strings from
DESIGN rows; oracle answers, relevant/obsolete IDs and source URLs are not passed
to serving. All selected messages deliberately precede an artificial cutoff.
That makes this a timing workload, NOT original-history replay or accuracy
testing. The 1,000 stored rows are replicas of 288 DESIGN capture templates,
not independent observations, sources or facts. Both controls receive identical
raw public text. Future captures have fresh keys and all clocks after the last
query. Every pre-packing frontier and packed packet is checked for future IDs.

**Confirmed metadata error:** primary `run.mjs`/`results.json` names the template
count `independentFactCount: 288`. That independence claim is incorrect. Retain
the frozen runner and original output for traceability; interpret the field as
DESIGN capture-template count ONLY. An independent fact count is not established.
The rescue runner uses `designCaptureTemplates: 288` and explicitly marks fact
independence as not established. No statistical inference uses 288 as a sample
size. Repeated physical records must never become extra independent evidence.

## Serving and learner boundary

Recall measures the complete service call, including nomination, forecast/rank,
journal persistence and packing. The tap checks all 50/200 selected candidates;
packing 6-10 records is not pre-scoring truncation. No runtime feature, forecast
guard or durable write is bypassed to reach a timing target. Source inspection
of the guard confirms journal/query/event/baseline/feature validation, not just
matching a stored journal ID. Preview predictions are research-worker records;
these synthetic labels do NOT establish agent utility or modification of the
consumer's scored forecast. All ordinary candidate forecasts still run.

Live publication age begins BEFORE the feedback validation gate, not on return
or after Close. The observer confirms successful processing and all 64 labels
are complete before closing the durable learner. All 128 admission/feedback rows
are decoded and checked against unique service bindings and timestamps; replay
must reconstruct 64 completions. This is same-epoch replay, not crash recovery,
cross-epoch transfer or hidden learning progress.

Capture overlap is recomputed from monotonic intervals intersecting Recall, not
a flag spanning both recall and feedback. All 256 offered writes must complete.
The serial one-ms ticker is closed loop: a slower writer changes instantaneous
load, and a cap can change that load as well. All samples, bytes and overlap are
reported; there is no equal-open-loop-load or sustained overload claim.

## Functional versus timing verdict

Timing failures remain nonzero exits even after all functional checks succeed.
The verifier independently recomputes per-call durations, overlaps, quantiles
and paired ratios from transcripts. It does not infer population p99 from 64
calls. At that count nearest-rank p99 equals the maximum. The full protocol asks
for failure/retry accounting: test exits and full successful call counts are
captured, but internal service retry counts are **not instrumented**. Do not
infer zero internal retries from a successful Recall. Peak RSS and per-phase
exclusive wall time are also not captured.

The diagnostic profiler covers the entire workflow including seed initialization;
whole-process CPU totals are not exclusive Recall cost. Block totals include
idle monitors/test goroutines and cannot all be counted as serving delay. Only
stacks scoped to Recall and lock ownership support the proposed bottleneck.
Profiling changes timing and cannot replace ordinary samples.

**Confirmed validation-harness gap:** the first race command exited green but
its opt-in `TestResearchDurableIdentityAudit` was SKIPPED. Preserve that transcript
and do not treat `validation-results.json`'s aggregate `passed` as execution of
the identity audit. It establishes vet and the temporal guard suites only.
The corrected `validation-v2.mjs` uses the existing non-optional persistent
guard, batch-parity and temporal suites and requires explicit PASS entries plus
no SKIP lines. This repairs coverage reporting, not production behavior. The
old identity audit is an unsafe-composition characterization, not a safety
guarantee. Full public-load race coverage remains untested.

Worker-cap configuration is isolated and deliberately selected after diagnosis.
It preserves source and durability but changes ANN construction scheduling;
functional availability/ledger checks do not prove bit-identical rankings or
agent quality. Even a successful finite pilot leaves corpus-dependent index
work, visible mutations, cross-epoch transfer, real outcomes, sustained load and
the complete seven goals unresolved.
