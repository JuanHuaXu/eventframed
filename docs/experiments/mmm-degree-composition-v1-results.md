# Fixed-noise composition v1 results

## Verdict

All four candidates FAIL the unchanged broad gates. This is an exploratory
replay of consumed synthetic data, not fresh confirmation, live agent evidence,
or a converged-BFGS experiment. All seven research goals remain OPEN.

| Additional expert / mixer prior | Non-harm | Recovery gain | Result |
| --- | ---: | ---: | --- |
| Fixed-noise64 / .95 generic64 | 799/840 | 53/128 | FAIL |
| Fixed-noise32 / .95 generic64 | 814/840 | 55/128 | FAIL |
| Fixed-noise64 / uniform | 807/840 | 56/128 | FAIL |
| Fixed-noise32 / uniform | 822/840 | 59/128 | FAIL |

The standalone fixed-noise64/32 models passed577/840 and508/840 non-harm
checks. Composition protects against many standalone regressions, but these
counts do not establish a useful increment over the existing mixture.

## Isolate The Increment

Compare each augmented mixer against a four-expert mixer with the SAME prior
convention. The uniform four-expert control is independently reconstructed;
the original .95-prior control matches the stored Go Markov predictions.

| Additional expert / prior | Paired non-harm | Positive .005 gains | Average Brier gain |
| --- | ---: | ---: | ---: |
| 64 / .95 generic64 | 166/168 | 6/168 | +0.00034728 |
| 32 / .95 generic64 | 164/168 | 0/168 | -0.00012732 |
| 64 / uniform | 166/168 | 18/168 | +0.00081035 |
| 32 / uniform | 164/168 | 0/168 | -0.00013110 |

Here168 cells include both scoring spans, both phases,21 cases and2 schedules.
The average is descriptive with overlapping spans, not an independent pooled
confidence claim. This supplementary comparison does not replace the840/128
gates. Uniform changes the original experts' masses too: it is not evidence
that simply adding an expert caused all the improvement against the old mixer.

The822/840 variant still fails18 non-harm comparisons, all terminal-span.
Some are uncertainty failures around near-zero mean differences; others are
measured deterioration. For example, phase0 delayed parity-to-majority has
Brier harm +0.02167 versus original Markov (interval +0.01098 to +0.03237).
Both phases' delayed majority-to-parity also deteriorate relative to Boolean.
See machine-readable failure details rather than treating all18 as proven harm.

## Verification And Cost

-2688 records; four candidates across all256 clocks, plus two mixture controls.
-2016 checkpoint/full-history comparisons and2016 unavailable-label poison tests.
-Independent log-space reconstruction checks1680 stratified predictions;
maximum absolute error1.78e-15. It shares neither checkpoint state nor
probability-space normalization with the collector.
-1376256 copied linear forecasts checked exactly. The unchanged scorer checks
source control metrics and all candidate probabilities.
-Original Go Markov control maximum difference2.28e-15 over every prediction.
-A second full collection is byte-identical. Timing sidecars differ as expected.

Collector timed loop3.595s; replay3.614s. These include all six mixture arms,
evaluation and sampled audits, exclude source parsing/output, and are NOT a
production latency benchmark. No model was refitted in this tape experiment.
An implementation adds eight fits per included window per256-frame episode;
both windows share16 extra fits across variants. The earlier64-label capped
fit fixture was4.147-4.192ms. Do not hide that work behind cheap mixture lookup.
Suffix-filter arithmetic is O(32*5); the offline harness retains full source
tapes. This is not evidence of a deployed bounded-memory implementation.

Output retains the six-arm summary schema (linear64/32 then four candidates)
and its legacy `Fits:48` field for scorer compatibility, not a claim of48 new
fits executed here. Forecast-generation/as-of tests for the underlying models
remain in the learned-degree-v1 experiment; label poison here tests composition.

## Next Discriminating Lead

Do not tune another fixed mixer prior or hazard on these outcomes. The new
64-label expert gives only a small increment, while32 gives none on average.
Investigate whether fitting degree variances to an as-of predictive Brier/LOO
criterion instead of Gaussian training marginal likelihood improves calibration.
This is different from the already-failed v93 expert-stacking experiment, but
it must retain its small-sample/shift warnings. Freeze the new objective and
verify LOO identities against explicit refits before using outcome-based gates.
No claim that this untested lead will succeed; no production or paper adoption.

## Artifacts

- Protocol: `mmm-degree-composition-v1-protocol.md`.
- Collector: `research/degree-composition-v1.mjs`.
- Auditor: `research/degree-composition-v1-audit.mjs`.
- Scorer: unchanged `research/spectral-regression-summary.mjs`.
- Forecasts and replay: `mmm-degree-composition-v1-{forecasts,replay}.jsonl`.
- Sidecars: corresponding `.jsonl.checks.json`, including same-prior losses.
- Summary: `mmm-degree-composition-v1-summary.json`.
- Independent audit/contrasts: `mmm-degree-composition-v1-audit.json`.

SHA256 forecasts/replay:
`f7fc119389f8bbba5132e617aa4df25dd051ef03726cb1bf1e64c51b3ffef604`.
SHA256 summary:
`0bb57970d85a3a57c6fd304cfd283073e4c437283bec56d2ab2bddda4160b4ea`.
Both input hashes match their previously frozen artifacts; full hashes are in
the audit JSON. Original failures and all prior artifacts remain untouched.
