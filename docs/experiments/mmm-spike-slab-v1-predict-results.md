# Mixture prediction and first pilot checkpoint

## Verified numerical component

Actual variational mixture integration is implemented in
`internal/observationlearners/spike_slab_predict_test.go`; derivation and
numerical contract are in `research/spike-slab-predictive-derivation.md`.
It preserves conditional-intercept dependence rather than substituting a
moment-matched Gaussian. Independent Gaussian/small-mixture integrations
differ by at most 1.284039541360471e-11 in the component grid. Maximum grid
cost: 5,851 integrand evaluations; full fitted fixture: 1,969. The asymmetric
zero-mean test gives 0.3420349397747625, not 0.5.

Characteristic-function enumeration, complement symmetry, fitted query mean
AND variance, query-law ownership, invalid inputs and evaluation caps pass.
The complete spike suite passed under race in 6.526 seconds before the pilot
adapter was added. The adapter's future/evaluator poisoning and admitted-label
positive-control tests passed under race in 4.154 seconds.

Analytic omitted-tail bound is below 1.45e-14. Finite-interval Simpson error
is an estimate, NOT a uniform numerical certificate or calibration guarantee.
These tests establish numerical consistency, not quality on agent tasks.

Apple M4, prediction only, three repetitions of ten operations:

| ns/op | B/op | allocations/op |
| ---: | ---: | ---: |
| 3175917 | 6144 | 3 |
| 3170038 | 6144 | 3 |
| 3174725 | 6144 | 3 |

Command: `go test ./internal/observationlearners -run '^$' -bench '^BenchmarkSpikeMixturePrediction$' -benchtime=10x -count=3 -benchmem`.
Fit preparation is excluded. This is not serving tail latency or a full
learning-cycle cost; earlier fitting cost is separately reported.

## Pilot stopped: confirmed representation saturation

The frozen pilot contract declares pi=1/255, slab variance 1, all 255 masks,
clock128/window64/index0 across 84 phase/case/schedule combinations. This
expects one active coefficient and prior query-logit variance two. It is an
explicit modeling assumption, not an outcome-selected optimum. Source SHA256:
`5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f`.

Collection stopped after 78 completed fits; it did NOT pass. Initial run
11.39 seconds; diagnostic replay11.32 seconds, same failure and byte-identical
partial artifacts. Raw diagnostic:

```
fit 78 phase1 case18 schedule0: fast spike coordinate overflow:
coordinate=149 inclusion=1 logOdds=36.745813783531418
mean=-2.8035886764902735 variance=0.090377188582484225
```

The finite log-odds rounds to probability one in the current sigmoid
representation. This violates the fitter's strict-interior check. The
variance and mean are finite and valid; this is confirmed numerical
representation saturation, not evidence that the model's quality fails.
No records were skipped, no fallback used, no prior or tolerance retuned.
The diagnostic edit only names values in the existing error.

Both partial JSONL files have SHA256
`c66c9289d0aaef86465e047246ff4bd1689526ad30a2aac9c2e309cab8578e24`.
Files: `mmm-spike-slab-v1-pilot.jsonl` and `mmm-spike-slab-v1-pilot-replay.jsonl`.
Do not report their successful subset as the 84-fit quality result.
The new audit script intentionally requires all84 and has not been run on
these incomplete artifacts.

NEXT: preserve finite inclusion log-odds and stable complementary mass through
factor entropy, moments and prediction; independently test extreme odds
before rerunning the unchanged pilot into new artifacts. Simply permitting
g=1 makes the current KL contain 0*log(0); silently clipping discards the
tail and changes the declared inference. Dense and fast paths must agree.
All seven research goals remain open. Production and whitepaper untouched.
