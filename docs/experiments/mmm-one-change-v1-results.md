# Bounded one-change snapshot results

Status: **FAIL** under the [frozen screen](mmm-one-change-protocol.md). Do not
promote the candidate or silently tune its priors/support. All seven full goals
remain open. The earlier full-segmentation and no-change failures are retained.

## Results

All1152 snapshots collected:384 schedule trajectories (192 paired latent
trajectories), three publication windows each. Both phases are consumed data.
Each cell uses32 independent trajectory indices; schedules and snapshots of the
same latent trajectory are paired, not extra independent samples.

| Requirement | No-change64 control | Existing full segment64 control | Combined |
| --- | ---: | ---: | ---: |
| Non-harm | 35/36 | 30/36 | 65/72 |
| Contaminated-window gains | 12/12 | 0/12 | 12/24 |
| Stationary non-harm subset | 12/12 | 12/12 | 24/24 |

The one-change candidate fixes much of the no-change model's mixed-regime error,
but does not improve reliably on segmentation that already existed. It preserves
the stationary placebo within the declared tolerance. That is not a broad
stationary guarantee: only case0 and the stated snapshots were screened here.

Phase1 expected Brier (lower is better):

| Feedback / case / clock | One-change | No-change | Full segmentation |
| --- | ---: | ---: | ---: |
| Delayed / stationary /160 | .217290 | .217734 | .216656 |
| Immediate / majority to parity /160 | .065269 | .230033 | .065642 |
| Delayed / majority to parity /160 | .254710 | .338282 | .229705 |
| Delayed / majority to parity /192 | .054924 | .144384 | .055702 |
| Immediate / parity to majority /160 | .111214 | .191607 | .109833 |
| Delayed / parity to majority /160 | .221965 | .347784 | .207284 |
| Delayed / parity to majority /192 | .089432 | .159083 | .093007 |

For the phase0 delayed majority-to-parity160 cell, full-segmentation minus
candidate gain is -.022070 [-.040482,-.003657], evidence of regression under the
exploratory interval. Other failed non-harm bounds need not prove harm: most
cross zero. No individual successful cell replaces the combined FAIL.

The earlier count-matched current-only reference at phase1 delayed160 achieved
.203094/.194155 for the two transitions, but receives the true change boundary.
It is diagnostic headroom, not a deployable control. The candidate's restriction
changes boundary prior/support and removes full-model hazard smoothing; this
experiment does not isolate those changes individually or prove one-change
models can never work. It rejects this frozen candidate as the proposed rescue.

## Verification and artifacts

- [Raw records](mmm-one-change-v1.jsonl),
  [summary](mmm-one-change-v1-summary.json),
  [run log](mmm-one-change-v1-run.txt),
  [race contracts](mmm-one-change-contracts.txt),
  [benchmark](mmm-one-change-benchmark.txt).
- Five tiny support/evidence fixtures and15 direct predictive-ratio tests;
  four detached concurrent fits; invalid-input controls: PASS under race.
- Two full64 snapshots: current/future/unavailable-label and Q poisoning,
  as-of origin order, and source ownership: PASS under race. Package39.733s.
- Independent JavaScript direct Beta-integral audit checks all57600 posterior
  weights, every relevant marginal log, and all36864 issued forecasts. Maximum
  numerical discrepancy5.69e-14. This audits issued query values, not every unused
  entry in each512-entry forecast table. H0 matches original v120 forecasts.
- Five corrupted-record tests rejected altered origin, cut, posterior weight,
  marginal likelihood and issued forecast. Independent reference also checks
  six support sizes and current/future-label/Q independence.
- Summary replay: byte-identical. Raw header pins source/protocol hashes;
  summary pins source tape, raw artifact and reference/scorer hashes.
- Collection17.74s wall/60.19s user with four workers. This is not serving latency.
- Three isolated fit64 benchmarks:37.52-37.63ms, about476880B and37 allocations
  per fit. No queueing, persistence, ingestion or concurrent-serving costs included.

Intervals are fixed-sample mean +/-3.5SE, not simultaneous or anytime coverage.
No daemon behavior, production, OpenClaw, dependencies or whitepaper changed.
No commit or push. The four preexisting tracked edits remain untouched.

## Next lead, not another prior sweep

Earlier predictive-query v120 only reweighted a fixed expert tape; its protocol
explicitly excluded expert retraining. Later acquisition-training tests did refit
experts, but used random, entropy and disagreement policies rather than a joint
boundary/evidence predictive acquisition model. Investigate that intersection
before proposing a new experiment: paid evidence should change the actual fitted
regime model, not just expert-selection weights.

Required design checks are substantive. A query needs the joint distribution of
its label and future targets under the regime hypothesis, including whether its
origin predates a boundary. Query selection cannot use hidden outcomes, Q or the
true boundary. A capacity-induced dropped observation changes the conditioning
set; one must not call that ordinary posterior conditioning. Define equal total
label cost, charge hypothetical refits and delay, preserve random/entropy controls,
and check prior work for an equivalent experiment before implementing. Existing
failure evidence does not establish this remaining lead's success.
