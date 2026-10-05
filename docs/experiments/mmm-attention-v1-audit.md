# MMM prototype audit

2026-09-12. Scope: new isolated observation controller and experiment only.
Existing serving code and prior neural-pilot files were not edited.

## Round 1: information boundaries and integration

The reader interface exposes only selected scope/depth prefixes, with no outcome
or task-family argument. The controller's model is fitted only on earlier
simulator samples. Generator/scorer functions stay in the experiment package.
Raw text and inferred Why fields cannot supply observed bits. Prediction uses
the same fitted table across controls. Repeated inspection consumes only novel
coordinate budget and never changes any conditional count. No production claim,
causal assertion, new evidence source, or publication authority is created.

## Round 2: chronology and representation guards

Added rejection of reversed prior/current timelines and reused coordinate-source
IDs before the frozen experiment. The finite fixture assumes distinct sources;
general overlapping views require a separately specified model. Tests reject
future-available evidence, cross-tenant records, inferred fields, undeclared
reader coordinates, and changing epochs. Missing reads do not become zeros and
cannot be retried for free. The snapshot is expected immutable; this prototype
does not provide a concurrent live-database snapshot implementation.

## Round 3: evidence and performance accounting

All rows and unsuccessful scenarios retained. Count of independent evaluation
trajectories is 3584; policy repetitions do not multiply sample size. Paired
intervals are fixed-sample approximate bounds, not confidence sequences. A
separate post-shift summary prevents aggregate gains from hiding harm. Strong
fixed-order and random controls are included alongside restrictive ablations.
Exhaustive inspection has more budget and sparser fitted conditionals, so it is
neither a matched control nor an oracle. Fitting-count support and the full table
remain unchanged under concurrent inference tests. Final artifact tests rehash
sources, recompute scores and summaries, and replay every trajectory exactly.

The array model costs about 2 MiB per known task and scales exponentially with
feature count. Selection enumerates at most nine views and eight outcomes per
view, under a six-coordinate inspection budget. Timing includes the current
implementation's RNG allocation even on deterministic policies. No serving
latency, model-training cost, queue behavior, or arbitrary corpus scaling is
established by this in-memory microbenchmark.

The confirmed shift failure is a research limitation, not silently repaired by
tuning on confirmation. Keep the artifact and freeze a new rescue protocol.
