# Failure diagnosis and pointwise safety control

This is post-hoc research on the consumed independent-v1 cohort. It is not a
new confirmation, deployment, or replacement of the seven success criteria.

## What failed

For d=c-b and p=b+w*d, binary expected Brier excess satisfies exactly

    Delta(p,b;q) = w*Delta(c,b;q) - w*(1-w)*d^2.

The all-forecast average weighted raw excess was +.000057226; the mixture
term was -.002428813, giving the observed -.002371587 improvement. Thus the
algebraic gain is primarily diversification, not a better raw challenger.
This decomposition is not a causal attribution experiment.

All34 harmful fixed windows had harmful average raw and unclipped-mixture
excess. Their means were raw+.060103312, unclipped+.018271360, and guarded
+.013553917. The guard improved28 of these windows and worsened6. It generally
mitigated the problem but did not certify conditional expected safety.

## Diagnostic replacement guard

The contract derives a pointwise scalar constraint from both binary outcomes,
using the same .01 allowance and unchanged Fixed Share proposed weights. This
bounds realized and conditional expected excess by .01 in exact arithmetic,
therefore also bounds average excess on every nonempty window. Floating-point
checks allow1e-12 numerical slack. It does not guarantee zero harm, calibration,
learning consistency, faster recovery, or general structured-output safety.

Unlike cumulative comparator constraints studied in
[Conservative Online Convex Optimization, Section3](https://ecmlpkdd-storage.s3.eu-central-1.amazonaws.com/preprints/2021/sub_12975_17.pdf),
this is a per-forecast additive bound. We used a direct binary-Brier derivation,
not that paper's regret theorem or an implementation of its algorithm.

| Proposal and guard | Whole expected Brier | Terminal64 Brier | Harmful32-windows >.01 |
|---|---:|---:|---:|
| Fixed Share / pointwise | .156847006 | .145296192 | 0 |
| Fixed Share / local ledger | .155144958 | .145014962 | 34 |
| Markov | .157516545 | .145520129 | reference |
| Constant half / pointwise | .156957522 | .145505054 | 0 |
| Full challenger / pointwise | .157230280 | .145859317 | 0 |
| Unclipped half | .156312476 | .147335080 | 720 |
| Raw challenger | .164432730 | .157270275 | 1544 |

The pointwise policy changes all172032forecasts, with mean weight.284298;
it is not an exact baseline fallback. Maximum observed conditional excess
was .009477416. Its whole-stream gain versus Markov is .000669539 (~.4251%),
retaining28.23% of the local-ledger gain. Eight-index cluster pointwise95%
interval[-.000751296,-.000582784] describes this consumed-data comparison,
not independent validation or a safety confidence sequence.

Adaptive proposals beat constant-half overall by.000110516, but the changing-
scenario whole-stream difference is only-.000006298, interval crossing zero.
Stationary terminal performance versus Markov is also inconclusive. The
guard has not established faster adaptation or recovered lost learning value.

## Verification and cost

- 15972 endpoint/conditional-risk grid checks passed, including zero budget,
 equal predictions, boundary probabilities, tiny differences, invalid inputs.
- Both full diagnosis and pointwise screen replayed byte-identically; bootstrap
 comparison replay was exact. No model fits or scores were silently retried.
- Warm scalar guard benchmark .0183--.0185 microseconds per forecast over three
 rounds of1720320calls. This is JIT-optimized arithmetic, not end-to-end timing;
 it excludes filtering, fitting, I/O, retrieval, persistence and serving.
- No Go/runtime code changed; prior races remain the last Go evidence. New
 work is evaluator/research JavaScript, not a daemon integration.

Artifacts: `mmm-spike-independent-v1-diagnostic.json`,
`mmm-pointwise-guard-v1-screen.json`, `-comparisons.json`, `-benchmark.json`.
Core SHA256: fd3250406343634cc015fbebb684c9bface7bf1fd3f8b92f4c21d37f92130253.
Screen SHA256: b2fab69f1ebb58c32a4500f221185131d7d609a48610c4f2b53663e5e5bf7f2f.
Diagnosis SHA256:930ca8e073b1232b781f64a9c475d66df268d60c2a23dca142934ff0442d0c82.

## Next investigation

Read the primary [Competitive On-line Linear Regression paper](https://papers.neurips.cc/paper/1419-competitive-on-line-linear-regression.pdf)
before designing a direct mixture-loss learner. Current updates score the
underlying experts, whereas delivered utility depends on their mixture.
Compare directly optimized mixture weights with existing Fixed Share and
constant-mixture controls, preserving as-of delayed feedback and explicit
safety/utility costs. The literature lead has been located but its full
algorithm has not yet been audited or selected. Do not inherit its guarantees
under missing data, changing models, or an added guard without proof.

All seven objectives remain open. Production, whitepaper, commits and pushes
remain untouched.
