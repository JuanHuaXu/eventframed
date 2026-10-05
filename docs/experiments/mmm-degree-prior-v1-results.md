# Degree-prior v1: fixed allocation does not rescue learning

All four fixed-prior candidates FAIL the frozen broad gates. The full2688
consumed-record collection and byte-identical replay each completed129024 fits
in5.94s test-body wall time. These are not untouched confirmation or loaded
serving timings. No production or whitepaper changes; all seven goals open.

## Broad and matched controls

| Candidate | Non-harm | Recovery gain | Verdict |
|---|---:|---:|---|
| Degree64 | 145/840 | 6/128 | FAIL |
| Degree32 | 304/840 | 38/128 | FAIL |
| Scalar64 | 119/840 | 8/128 | FAIL |
| Scalar32 | 206/840 | 34/128 | FAIL |

The five broad controls remain same-window linear L2, logistic, generic,
Boolean, and Markov, with Boolean omitted from the recovery-gain requirements.
The tables preserve all thresholds from the frozen protocol. Counts are
correlated checks, not independent replications or a probability of success.

| Explicit contrast | Non-harm | Recovery gain |
|---|---:|---:|
| Degree64 vs old full64 | 113/168 | 10/32 |
| Degree32 vs old full32 | 101/168 | 12/32 |
| Scalar64 vs old full64 | 168/168 | 2/32 |
| Scalar32 vs old full32 | 168/168 | 0/32 |
| Degree64 vs scalar64 | 98/168 | 10/32 |
| Degree32 vs scalar32 | 88/168 | 12/32 |

Scalar shrinkage stays within the declared .01 non-harm allowance against the
weak old full model everywhere; this is not zero-harm certification. Equal
degree budgets can improve particular low-order tasks more, but lose that
protection elsewhere. Neither meets the broader comparison. All individual
contrasts, intervals and gates are in `mmm-degree-prior-v1-contrasts.json`.

Historical phase1 delayed/missing terminal64 expected Brier, 64-label models:

| Case | Degree | Scalar | Logistic | Markov |
|---|---:|---:|---:|---:|
| Additive stationary | .243641 | .275141 | .207774 | .221755 |
| Additive gradual | .281286 | .292096 | .244261 | .239065 |
| Hierarchy gradual | .294568 | .289702 | .263783 | .244456 |
| Local-table gradual | .299137 | .293781 | .272673 | .250387 |
| Parity4 stationary | .265371 | .216538 | .285755 | .049066 |
| Majority to parity | .275814 | .232205 | .288538 | .060262 |
| Parity to majority | .159447 | .233358 | .114788 | .102272 |

Same total prior variance does not imply the same inductive bias. The equal-
degree prior redistributes variance toward low-order terms, so its lower-order
gains and high-order losses are consistent with that change. This experiment
does not establish the optimal allocation, or show that all fixed priors fail.
Do not continue a score-driven fixed-weight grid and call it confirmation.

## Verification

- All pairwise feature-sum/kernel identities checked over512x512 inputs for
  degree, scalar and unit-prior audit modes (786432 pair checks). Both candidate
  kernel diagonals are37. Unit-prior predictions match the earlier full model.
- Independent pivoted Gaussian elimination with combinatorial Hamming kernels
  reconstructs1512 stratified fits and48384 forecasts, maximum difference
  1.999e-14, with504 independently reconstructed evidence-origin lists.
- All1376256 linear-control forecasts agree exactly between experiments.
  The reused summary verifier checks4128768 candidate probabilities and161280
  archived control metrics; all models' origins match the as-of source journal.
- Constant-label, duplicate-input, invalid-input, unavailable/future-label and
  evaluator-metadata poisoning tests pass under the race detector. All2688 raw
  forecast records replay byte-for-byte. This verifies research arithmetic and
  data flow, not a live asynchronous serving implementation.

Artifacts: `mmm-degree-prior-v1-summary.json`, `-contrasts.json`, `-audit.json`,
`-contracts.txt`, `-forecasts.jsonl`, `-replay.jsonl`, and `-bench.txt`.

## Component cost

Three repetitions, Apple M4, 64-label fixture: degree fitting46.22-46.28us;
scalar fitting39.14-39.46us. Both allocate6392 bytes per fit (10 allocations),
excluding stack scratch and shared initialization. Prediction167.5-168.4ns,
zero allocations. The fixture retains all256 features. These timings exclude
I/O, queuing, persistence, publication and partial-input marginalization.
This implementation reuses the previous approximately5MiB Boolean tables even
though the new kernel only needs ten Hamming-distance values per fit. That
research reuse is not a claim of necessary production memory or general-domain
scalability. No concurrent experiment or source hashing during this benchmark.

## Next distinct lead

The source paper actually learns interaction-order weights from training
evidence. This experiment deliberately did not. A bounded, as-of-only learned
group-weight model is therefore a distinct untested lead, but must declare its
objective, regularization, bounds, optimizer, approximation semantics, and
holdout gates before execution. An empirical Gaussian likelihood fitted to
binary labels remains a working model, not an ordinary Bernoulli posterior.
Keep both fixed priors and all stronger controls; no future labels, repeated
overlapping-window evidence multiplication, or whole-goal completion by fiat.

## Hashes

- Forecasts: `02c76d2956c402b88584aaa88666980e6e418449ec24840eaa9c8ad38674dd27`
- Summary: `9c44def4cf9dc8293f3e1ed12acd1c2f9be59f2a707552add4c31b1792b1dfaa`
- Contrasts: `433a0cd3f49fd2d4dd2bdf2f81308f584df6faa30d110a2fbc47fdc1be8d9edf`
- Component: `06dbd5b40d3d977aaaf1bd6f16df90521a64a3e0614f7048df93fd8d773b68e3`
- Collector/isolation: `796ba7c1efa68ea25f02820a502c842e7ae19432746a472d9fcab699653c85c7`
- Independent audit: `9c354cd3d707f03d808ec9c0ab4b176dcb002affc4e1a1bfeb3a1872bf29d9c3`
- Contrast code: `b0fd4d14b90793cfef42f30561bfbd8ec535c368ec69738a2686c2979259deef`
- Protocol: `d9113cab79e6ace08c1a28d4553547bc382da60123785979fcc16057164a91fa`

Input and reused summary-code hashes are recorded in the preceding spectral
regression results. No archived input or predecessor algorithm was edited.
