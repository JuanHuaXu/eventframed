# Retained challenger under alternative input generators

Frozen 2026-10-01 before execution. This is a research-only covariate-generator
stress test of v8, not an independent outcome-mechanism generator or agent-task
validation. Preserve the v8 output and source unchanged. Its
`internal/observationlearners/retained.go` SHA-256 is
`ecbc832eb2ea3d71957fe0ecd0da226325c60c5b4e247cc7ab245ddcdda86370`.

Use a Go build overlay to replace exactly the two calls drawing uniform
`uint16(rng.Intn(512))`: one for the 4096-sample frozen base fit and one for
each online frame. No other source expression may change. The outcome laws,
scenarios, fitting seeds, audit probability, missingness, delay, four arms,
same-frame read paths, score windows, and frozen v8 `SummarizeRetained` gates
remain unchanged. Fit and evaluation use independent seeds, but the same
declared input generator. All arms share each frame and feedback tape.

Input bits are numbered 0 through 8. Two generators are fixed:

1. `biased`: independent bits 0-2 have probability 0.75 of one; bits 3-5
   probability 0.5; bits 6-8 probability 0.25.
2. `latent`: draw a fair latent bit Z. Bits 0, 1, 6, and 7 independently
   copy Z with probability 0.85 and flip it with probability 0.15. Bits
   2, 3, 4, 5, and 8 are independent fair bits.

Run every original v8 scenario with three independent 4096-label base fits,
eight streams per fit, and both design and confirmation phases: 480 streams
per generator, 960 total. Each stream is 512 frames. Reuse the existing
confirmation gates verbatim, including the 3.6-SE paired intervals,
`>=0.005` primary shift gains with positive lower bound and all fit signs,
stationary protection, and `>=-0.01` all-scenario mean-harm bound. A strong
input-generator replication requires at least one retained arm to pass all
frozen gates in **both** generators. Report every failed cell, not only
aggregate verdicts. Do not select a new threshold or arm after seeing results.

Verify the overlay changed only the two pinned draw sites and leaves the
original source hash intact. Record each generated source hash, exact input
seed scheme, full forecast journal and summary. Audit pre-outcome predictions,
delivered-origin ordering, equal audit/available counts, summary score
recomputation, and static-vs-adaptive differences. Record tree fit/update
nanoseconds, node counts, and process wall time; these are offline experiment
costs, not daemon serving latency. Preserve failures and do not publish an
algorithm change from this screen alone.

This test deliberately changes P(X) and hence the frequency of contexts. It
does not test a new P(Y|X) family or real text. Distribution-shift evaluation
is motivated by [Gama et al. (2013)](https://link.springer.com/article/10.1007/s10994-012-5320-9),
but no theorem from that work is inherited here.
