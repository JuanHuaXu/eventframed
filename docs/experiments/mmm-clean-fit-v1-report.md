# Count-matched clean evidence: useful diagnostic headroom

## Finding

Existing subset models improve substantially when old-regime evidence is removed
at the same sample count. This supports investigating training-window
contamination before adding another model family. It is not a successful
deployable reset: the evaluator knows the true boundary and the current gate
does not. All seven research goals remain open.

Full-input expected Brier at checkpoint480, subset model (lower is better):

| Case / schedule | Cohort | Actual last64 | Matched mixed | Clean, same count |
| --- | ---: | ---: | ---: | ---: |
| Majority to parity / immediate | 1 | .14763 | .17503 | .09031 |
| Majority to parity / immediate | 2 | .14278 | .18462 | .08708 |
| Majority to parity / delayed | 1 | .24588 | .26382 | .15642 |
| Majority to parity / delayed | 2 | .22849 | .25075 | .13784 |
| Parity to majority / immediate | 1 | .14291 | .15942 | .07419 |
| Parity to majority / immediate | 2 | .12310 | .15292 | .07442 |
| Parity to majority / delayed | 1 | .19704 | .21427 | .11425 |
| Parity to majority / delayed | 2 | .19194 | .20623 | .10462 |

At checkpoint384 delayed majority-to-parity, clean subset risk is still .25403
and .25878, with only about15 eligible labels. Cleaning is not sufficient when
support is too sparse. The count model remains near .24-.25 under full-input
evaluation because these sample sizes populate few of its512 exact input cells.
That observation is not a general criticism of its partial-view forecasts.

## Design

[Contract](mmm-clean-fit-v1-contract.md): all128 trajectories, both schedules,
checkpoints128/256/384/480. Each checkpoint uses the latest fit published
strictly before that clock. Actual takes its last64 origins. Clean takes up to64
already-arrived post-boundary origins. Matched takes a seeded random subset of
actual, of exactly clean's size. All origins must be audited, nonmissing, and
available at the original fit clock. No future outcomes or extra labels are
borrowed. Both existing model classes are fitted without prior changes.

The diagnostic enumerates all512 full inputs with5% noise to remove observation
choice and mixture-weight effects. It measures expected simulator risk, not
agent accuracy or actual served Brier. Before a change and in stationary cases,
clean and actual predictions are exactly equal. At179 of1024 checkpoints no
model was published yet; all three arms remain neutral, and these cases stay
in the results. Empty post-change sets are likewise retained.

Matched controls use one seeded subset per cell. The comparison isolates sample
count but changes temporal composition; means do not establish simultaneous
coverage or a universal benefit. An estimated change boundary could discard
useful observations, retain stale ones, or trigger falsely. Those costs remain
untested here.

## Verification

-1024 records collected in16.75s and replayed in16.73s; raw artifacts are
byte-identical. These times are offline diagnostic execution, not serving tails.
-Independent scorer verifies909 captured source hashes and1,572,864 count-model
probabilities; recomputes all full-input risk values, sample eligibility,
matched counts, stationary equality and neutral states. Summary replay is exact.
-Focused race contracts pass (final1.225s package time); vet passes. Full
experimental collection was not race-instrumented.
-Initial collection stopped on a valid no-publication state. The collector was
corrected to retain that state as neutral. Its failed header artifact remains
`mmm-clean-fit-v1.jsonl`; use the `-verified` file, not that partial output.

Raw SHA256: `ba5e0fcdd9e28a0cca2e027c9fdc69bd2945eed24a0a8906683c61fa2ce1d090`.

Artifacts: [verified raw](mmm-clean-fit-v1-verified.jsonl),
[native replay](mmm-clean-fit-v1-replay.jsonl),
[summary](mmm-clean-fit-v1-summary.json),
[replay summary](mmm-clean-fit-v1-replay-summary.json),
[scorer](../../research/clean-fit-summary.mjs),
[collector](../../internal/observationgate/clean_fit_diagnostic_test.go).

## Next Candidate

Inspect earlier reset/epoch experiments before implementing a duplicate. Test
a fit-origin boundary derived solely from the existing accepted split and
available evidence, preserving false-split controls and explicit behavior when
little post-split evidence exists. First quantify how much clean evidence that
actual gate retains compared with this hindsight diagnostic. A detector can
identify a change too late to retain enough useful data, so blindly discarding
everything before the detection clock is not yet justified.

No production source, paper, remote, commit or push changes.
