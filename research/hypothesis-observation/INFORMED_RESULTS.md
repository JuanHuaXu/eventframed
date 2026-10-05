# Informed source-mode priors v4: FAILED overall screen

Completed 1,792 synthetic episodes: seven cases, two disjoint evaluation splits,
128 episodes per cell, four paired arms (7,168 arm trajectories). Protocol,
source hashes, calibration counts and raw acquisition/forecast traces are in
`INFORMED_PROTOCOL.md`, `informed-v4.json.gz` and `informed-v4-summary.json`.
No production change, real-source authentication or real-agent experiment occurred.

## Confirmation outcomes

Lower multiclass Brier is better. These are means over 128 episodes per row.

| Case | Fixed joint final Brier | Informed final Brier | Fixed accuracy | Informed accuracy |
| --- | ---: | ---: | ---: | ---: |
| Independent, 5% noise | 0.000226 | 0.000145 | 100% | 100% |
| Independent, 20% noise | 0.188769 | 0.167410 | 88.28% | 88.28% |
| Copied, 5% noise | 0.126845 | 0.125138 | 91.41% | 91.41% |
| Copied, 20% noise | 0.408615 | 0.438963 | 68.75% | 67.97% |
| Mixed, 20% noise | 0.329346 | 0.316164 | 77.34% | 78.13% |
| Mixed, random signal | 0.254033 | 0.310216 | 79.69% | 75.78% |
| Mixed, misleading signal | 0.275650 | 0.343978 | 81.25% | 78.13% |

Three frozen conditions failed:

1. Independent20 curve-Brier gain was 0.015994, but its paired z=3.3 interval
   [-0.009168, 0.041155] crossed zero. The mean passed the 0.01 threshold; the
   uncertainty condition did not. Final-Brier gain was 0.021358 with interval
   [0.000760, 0.041957], but that was not the required recovery statistic.
2. Copied20 curve-Brier harm was 0.020981, above the 0.01 ceiling.
3. Copied20 final-Brier harm was 0.030349, above the 0.01 ceiling. Its interval
   spans zero: this is failed mean non-harm screening, not established population
   harm at the stated descriptive confidence level.

Random and misleading signals increased mean final Brier by 0.056183 and
0.068328. For the misleading case, the harm interval was [0.007400, 0.129256].
These stress cases were declared diagnostics, not added to the pass gate afterward.
Confidently-wrong counts in mixed20 were 4/128 fixed versus 5/128 informed;
misleading signals gave 3/128 versus 6/128. Zero counts in other cells do not
establish a zero failure probability. The artifact's degenerate normal intervals
on all-zero paired differences are not useful rare-event confidence bounds.

The cheaper one-source control scored 0.408023 on copied20, versus 0.438963
informed. It used 16 total units (eight signal checks plus eight observations),
versus 24 for the other arms. Extra source reads have not earned their cost here.

## Verification and interpretation

`python3 -m unittest test_informed_v4 -v` passed all three tests, including full
replay of all episodes and source hashes. The factorized update agrees with
enumeration of all 4,096 hypothesis/mode states after each of five observations,
including unequal priors and contradictory reports. Changing only calibrated
priors leaves all non-informed arms unchanged. Forecast normalization, unique
acquisitions and charged query counts are verified in the full replay.

Disjoint calibration yielded independent-mode priors 0.098133 and 0.902893 for
negative and positive signals. This is a synthetic 90%-accurate signal, not an
empirically authenticated provenance service. Although the likelihood for the
signal is calibrated separately, fixed all-independent/all-copied populations
do not match the model's equal mode prior. Finite calibration uncertainty is
also not integrated over shared signal parameters. Neither caveat licenses
discarding the failures: they are relevant operating mismatches.

## Next discriminating experiment

Do not retune the prior or replay this confirmation as new evidence. Separate
acquisition from forecast weighting on new seeds: fixed/informed acquisition
crossed with fixed/informed scoring, all scorers receiving the exact same tape
within each acquisition policy. That can distinguish a poor choice of what to
observe from a poor inference from already-observed evidence. Include an explicit
matched generative-mode population and retain all-copy stress. A cost-sensitive
stop rule remains a separate lead; no claim of lead exhaustion is warranted.
