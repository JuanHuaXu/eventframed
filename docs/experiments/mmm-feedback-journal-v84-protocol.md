# Bounded forecast journal v84

Research lifecycle prerequisite for delayed/missing-feedback experiments, not
a new accuracy claim. Preserve the frozen v82 predictor and feedback transition.
Separate strictly increasing issuance IDs from arbitrary-order feedback IDs.
Retain at most64 origin-indexed forecast snapshots in a bounded ring. Backpressure
must occur before reading input if a live slot would be overwritten.

Each entry contains the emitted probabilities, inner expert probabilities and
availability, model publication version, and pooled/local revocation generation.
Feedback uses that entry, never recomputes a forecast on its revealing label.
The single owner serializes calls; this object is not a concurrent service or
durable database. Publication atomically replaces immutable model references
and requires a strictly increasing version. No production code is changed.

Feedback for an older model version or pre-split generation is consumed as
stale, without updating current selector weights or authorizing a split.
Observed labels may still be supplied separately to a training owner under its
audit/eligibility policy; this journal neither invents nor automatically trains
on them. This conservative policy can discard useful delayed weight evidence;
measure that cost in a later experiment rather than assuming it is free.

Missing feedback requires explicit censoring or expiry by origin cutoff. Neither
operation applies a negative label or any Bayesian update. Duplicate, unknown,
future and expired IDs fail without mutation. No cutoff may refer to unissued
origins. The owner chooses expiry; there is no hidden wall-clock timer.

Tests must establish immediate-feedback parity for enabled and disabled subset
models, snapshot correctness, out-of-order arrival, pending bounds/backpressure,
ring reuse, publication and split invalidation, failed-read atomicity, duplicate/
expiry isolation and accounting: issued=applied+stale+censored+pending. A frozen
20-seed2048-issuance stress schedule exercises delays0..31, independent20% missing
labels, version publication every64 issues and explicit48-origin retention.
Drain remaining entries; replay hashes and counters must match. This is lifecycle
stress, not a valid power/calibration experiment. Run race/vet and microbenchmarks.
