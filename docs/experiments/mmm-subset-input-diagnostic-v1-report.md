# Input marginalization: useful but conditional headroom

This diagnostic confirms that uniform averaging over hidden fields can cause
substantial error in the retained subset learner under a simple dependency.
It does not establish an integrated rescue or a software defect.

64 independent training fits, two input laws, n64/n128, 16 fits per cell.
All three arms use the same training labels and subset outcome model. Only
the input distribution changes. Scores integrate exactly over the 512 raw
inputs and label noise, not a sampled evaluation set. Only bit0 is observed;
the label is bit2 with .05 noise. The dependent generator copies bit2 into bit0.

| Input law | Training labels | Uniform Brier | Empirical Brier | Oracle-input Brier |
|---|---:|---:|---:|---:|
| Independent | 64 | .250303 | .253944 | .250303 |
| Independent | 128 | .250408 | .254570 | .250408 |
| Copied field | 64 | .104211 | .049674 | .049675 |
| Copied field | 128 | .100469 | .048129 | .048106 |

For copied inputs, empirical gains are .054537 [.049119,.059955] and
.052340 [.049420,.055259], using paired mean +/-3.5SE over training fits.
Independent-input gains are negative: -.003641 and -.004162. These are
exploratory intervals, not sequential or simultaneous deployment guarantees.
Oracle partial-input floors are .25 and .0475, respectively.

Full-input forecasts agree across all arms within1e-12. Their mean Brier is
.049532/.048539 for independent inputs and .050283/.048207 for copied inputs.
The outcome learner already estimates full-input outcomes well in this fixture;
the large dependent partial-input gap is a marginalization issue. Partial risk
can beat a finite fitted full-input forecast through averaging; no claim that
partial information beats the full-information Bayes optimum is made.

## Decision

Worth an integrated, fresh, equal-observation-cost test of distribution-aware
subset predictions. Do not unconditionally substitute empirical weights:
finite-sample noise hurts independent inputs. Keep the existing selector and
Anti-Pigeon authorization unchanged to isolate this candidate. Check existing
distribution-aware work before inventing another estimator or safeguard.

In particular, `mmm-acquisition-v11-results.md` already rejected an empirical
forest integration: clustered gains were only about .0003, and oracle input
weights did not rescue that workflow. The present fixed-mask subset result is
not a reversal of that failure and could also disappear through observation
selection or low mixture influence. Those mechanisms require direct testing.

## Verification

- Race-enabled diagnostic PASS: 2.51 seconds experiment, 3.981 seconds package.
- Exact population floor and full-input negative-control checks passed.
- Four captured sources verified against both embedded text and current files.
- Fit seeds,64 records,16 per cell, metric ranges and parity verified.
- Summary replay byte-identical. This runtime includes repeated fitting and
  race instrumentation; it is not a serving-performance benchmark.

Raw: `mmm-subset-input-diagnostic-v1.json`.
SHA256: `8253abb7adaf066d7156f1a1bac382b4a6fe619b8b12371bf513760ec0b14175`.
Frozen design: `mmm-subset-input-diagnostic-v1-contract.md`.
Evaluator: `research/subset-input-diagnostic-summary.mjs`.
No production, remote, whitepaper or existing candidate source changes.
All seven goals remain open.
