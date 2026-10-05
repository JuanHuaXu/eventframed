# Interaction regression v1: broad rescue fails

All four candidates FAIL the frozen broad gates. This is exploratory evidence
on consumed v120 records, including both historical phases, not fresh confirmation.
The preceding fresh component screen and these full-case forecasts are different
experiments. All seven research directions remain open.

## Results

| Candidate | Non-harm gates | Recovery-gain gates | Verdict |
|---|---:|---:|---|
| Screened64 | 482/840 | 20/128 | FAIL |
| Screened32 | 331/840 | 12/128 | FAIL |
| Full256-feature64 | 109/840 | 8/128 | FAIL |
| Full256-feature32 | 208/840 | 31/128 | FAIL |

Gate counts are correlated checks, not independent experiments. Each comparison
uses all32 trajectories in its case/phase/schedule, with the frozen +/-3.5 SE
exploratory interval. All individual gates and expected/realized Brier and
expected accuracy summaries are in `mmm-spectral-regression-v1-summary.json`.

Historical phase1, delayed/missing, terminal64 expected Brier, 64-label models:

| Case | Linear L2 | Screened L2 | Full L2 | Logistic | Boolean | Markov |
|---|---:|---:|---:|---:|---:|---:|
| Additive stationary | .206976 | .207415 | .276267 | .207774 | .236019 | .221755 |
| Hierarchy gradual | .266970 | .267759 | .295191 | .263783 | .246227 | .244456 |
| Local-table gradual | .276713 | .275839 | .295857 | .272673 | .249170 | .250387 |
| Parity4 stationary | .291024 | .054303 | .215460 | .285755 | .048129 | .049066 |
| Majority to parity | .294204 | .101751 | .232389 | .288538 | .075374 | .060262 |
| Parity to majority | .125312 | .118529 | .232200 | .114788 | .199795 | .102272 |

Screened parity4 reaches94.956% expected accuracy, compared with49.868% for
both linear models, but the existing Boolean/Markov controls reach95% with
better Brier. Thus the interaction representation repairs a genuine restricted
linear-model blind spot; it does not rescue the best existing system. The full
feature dictionary is not enough either. High-dimensional estimation and prior
allocation are plausible causes of its degradation, not established by this
experiment alone. No claim of irreducible failure or exhausted leads follows.

## Verification and cost

- 2688 records,129024 fits,4128768 saved candidate forecasts. Independent
  scoring checks161280 archived control metrics. Every fit's evidence origins
  match independently rebuilt as-of availability. Maximum dual residual4.845e-13.
- Independent pivoted Gaussian elimination and Hamming-distance full kernel
  verify1512 stratified fits,48384 forecasts and504 origin lists. Maximum
  forecast difference7.089e-14. This is not independent reconstruction of all fits.
- Analytic main-effect, parity, duplicate-input, constant-label, invalid-input,
  no-mutation, future-label and unavailable-label isolation tests pass under
  Go's race detector. Poisoned teacher Q and archived forecast fields do not
  affect the new learner. Tests exercise the research adapter, not production.
- Full byte-identical replay passes: initial collection5.60s, replay5.67s Go
  test-body wall time, excluding compilation and prior initialization. These
  are sequential offline research runs, not serving latency.
- Three component benchmark repetitions on Apple M4: linear fit24.90-24.96us,
  screened fit50.54-51.05us, full fit34.36-35.49us. Prediction6.60-6.64ns,
  7.05-7.06ns and167.2-168.1ns respectively, zero prediction allocations.
  The benchmark fixture yields a small screened model; it is not the maximum
  26-feature case. Fits report2552/2568/6392 allocated bytes respectively,
  excluding stack scratch and one-time tables. Fixed tables occupy about5MiB;
  their512-state construction is specific to nine Boolean inputs and does not
  establish scalability to arbitrary event spaces. A source-hash read overlapped
  part of this benchmark session; treat these as indicative component timings,
  not tightly isolated or loaded p99 measurements.

The first collector attempt rejected the source's manifest header as a record
before producing forecasts. The empty `mmm-spectral-regression-v1.jsonl` is
retained; the actual dataset is `mmm-spectral-regression-v1-forecasts.jsonl`.
Header parsing was fixed without changing data or fitting rules. An aging-rule
mismatch inherited from query experiments was also caught and removed before
collection: v120 learner forecasts are held fixed between publications.

## Next discriminating work

Do not repackage a soft mixture over individual parity atoms as a new lead:
`fitBooleanSpecialist` already integrates a Beta agreement rate over512 such
atoms. A genuinely different follow-up must address combinations of weak
interactions or justified pooling across time, rather than rediscovering that
specialist or lowering a threshold until this dataset passes. Degree/group-wise
regularization is still untested here, but needs a frozen data-independent
allocation or independently validated fitting rule, plus all current controls.
Do not promote any candidate or change production/paper based on these negatives.

## Artifact hashes

- Forecasts: `3fb74dbe9595d92a65ec289ae5911d8d4199d1fe048c953c6350e062c4da7bc0`
- Summary: `fd2b1c1827724c4ecd0181f4e7937f3817245e0c5e499f36ab3c126e7504e5df`
- Input: `5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f`
- Collector/component: `52bb60777eecd764f9c0961e42f94f6e7467e6eddd1caaf2651486e30c8cdf22`
- Isolation tests: `47838f2538ab084c5fec1bf98eddb4bbad950788c0073f58ac2823c0d6575555`
- Summary code: `45cbe137e244470c7d11450fd90aceae12b3ef23efaceb0e1f02eb547947c115`
- Independent audit: `bc635ab9b7321776872f5f536dd8b22cf4026724183e2bdc8c5244b7e3539666`
- Protocol: `f17d4c4f314b884e55354c6616166a57fdbad225a8fcacc834a2b67a0594adf1`
