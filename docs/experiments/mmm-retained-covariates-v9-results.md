# Retained challenger: alternative input generators

2026-10-01. The frozen [v9 protocol](mmm-retained-covariates-v9-protocol.md)
reused the v8 four-arm learner and gates without changing its source. A Go
build overlay changed only the two input draws. Two 480-stream studies used
biased independent bits and latent-correlated bits, respectively. Each stream
had 512 frames; the four arms shared inputs, outcomes, audit selection and
feedback. These are new input-generator tests, not new outcome mechanisms.

## Frozen verdict

| Input generator | Replacement | Adaptive retained | Static retained |
| --- | --- | --- | --- |
| Biased independent | FAIL | PASS | PASS |
| Latent correlated | FAIL | FAIL | PASS |

The strict preregistered replication criterion was met by **static retained**
only. This strengthens the earlier synthetic case for retaining a short-count
learner alongside the tree, but does not validate adaptive mixing as necessary
or superior. It does not complete goals 1, 2 or 4.

Confirmation post-change Brier gain relative to fixed MMM, with the frozen
paired 3.6-SE lower bound in parentheses:

| Generator | Arm | Shift 128 | Shift 256 |
| --- | --- | ---: | ---: |
| Biased | Adaptive retained | 0.01672 (0.00917) | 0.01761 (0.00723) |
| Biased | Static retained | 0.01415 (0.00980) | 0.01304 (0.00516) |
| Latent | Adaptive retained | 0.02040 (0.00741) | 0.00552 (-0.00139) |
| Latent | Static retained | 0.01700 (0.00810) | 0.00827 (0.00271) |

Every failed gate is retained: the biased replacement arm breaches the mean
harm floor in recurring/post (-0.01071) and interaction/post (-0.01676) and
the stable05 protection limit (upper harm 0.01257). The latent replacement
arm breaches recurring/post (-0.01289), interaction/post (-0.01183) and the
shift256 gain/lower-bound check (0.00471, lower -0.00439). Latent adaptive
retained fails only shift256 lower-bound positivity (gain 0.00552, lower
-0.00139). The other frozen protection, mean-harm and per-fit sign checks
pass for adaptive and static retained in both generators. No thresholds or
arms were selected after these results.

## Verification and cost

Full journals, summaries, and independent checker outputs are available for
[biased](mmm-retained-covariates-v9-biased-check.json) and
[latent](mmm-retained-covariates-v9-latent-check.json) inputs. The separate
JavaScript checker recomputed all 120 paired comparisons and three verdicts
per generator from the 480-record summaries, checked the untouched original
source hash and confirmed exactly two draw-site substitutions before the
overlay helper. The Go runner checked 983,040 pre-outcome probabilities and
512-frame score/availability/feedback chronology per generator, including
independent recomputation of every arm's Brier, log loss and accuracy. Both
studies completed with no runner failure.
The focused `TestRetainedControlsAndJournal` race test and package `go vet`
passed. A separate full-package race run was intentionally interrupted after
345 seconds when its unrelated research tests reached roughly 2.3 GB; it is
**not** reported as a full-suite pass.

Mean per-512-frame stream cost recorded inside the learner, not serving:

| Generator | Count fits | Tree rebuilds | Tree updates | Max current tree nodes |
| --- | ---: | ---: | ---: | ---: |
| Biased | 4.23 ms | 0.148 ms | 358.1 | 15 |
| Latent | 4.20 ms | 0.148 ms | 358.1 | 15 |

Core run wall times were 11.05 and 11.08 seconds for 480 streams, excluding
some artifact serialization. These timings are single-machine offline test
measurements, not p95/p99 request latency, loaded persistence, or a production
throughput claim. The fixed 64-audit tree rebuild schedule is not an adaptive
window mechanism by itself.

## Limits and next test

The frozen v8 runner uses three base-fit seeds independent of evaluation
RNGs, but reuses each fitted base across design and confirmation. Its earlier
result note has been corrected; these v9 results inherit the same structure.
The 3.6-SE intervals are conditional on that finite fitting set and are not
general population guarantees. The outcome laws, bit semantics, and audit
mechanism are still from v8. In particular, latent correlation is not evidence
of real text or agent-task transfer. A genuinely independent outcome family,
new fitting samples across phases, partial-view dependence stress, and
untouched outcome-labeled agent tasks remain required before broader claims.
No production source or served forecast changed.

Source and result paths for reproduction:

```text
research/build-retained-covariates-v9.mjs
research/retained-covariates-v9/{biased,latent}/{manifest.json,overlay.json,retained.go}
internal/observationlearners/research_covariates_run_test.go
research/check-retained-covariates-v9.mjs
docs/experiments/mmm-retained-covariates-v9-{biased,latent}.json.gz
docs/experiments/mmm-retained-covariates-v9-{biased,latent}-summary.json
docs/experiments/mmm-retained-covariates-v9-{biased,latent}-check.json
```

The builder and runner use exclusive-create outputs. Reproduction must use
fresh overlay and artifact paths; existing evidence must not be overwritten.
