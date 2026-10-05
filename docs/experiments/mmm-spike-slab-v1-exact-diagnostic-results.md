# One-feature original-likelihood posterior diagnostic

Same pi=1/255, unit slab and unit intercept prior as the screened model.
Balanced +/-1 designs, n16/64, with0 or2 flipped labels in each class.
Numerically integrate the spike's one-dimensional null evidence and the
slab's two-dimensional original logistic likelihood on [-10,10]. Two midpoint
resolutions512/1024 agree within1e-8 for log Bayes factor, inclusion and
prediction; complement tests pass. No JJ likelihood surrogate enters this
reference. This is not an exact255-feature posterior audit or a uniform
quadrature-error certificate. Gaussian-tail truncation and finite resolution
remain numerical approximations.

| n | Flips/class | Reference inclusion | VI inclusion | Reference P(+1) | VI P(+1) |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 16 | 0 | .792662804199 | .815824305341 | .790796594517 | .789362074878 |
| 16 | 2 | .0105246102502 | .0088992968393 | .502035241570 | .501583463003 |
| 64 | 0 | ~1 | ~1 | .950338049976 | .951601630573 |
| 64 | 2 | .999999998661 | .999999999984 | .901069168801 | .899919257000 |

Log evidence ratios(slab/null):6.87838574967,.993875721706,
36.8362800084,25.9686533004, respectively. Single-feature predictive errors
are below.00144 in these examples. With16 labels and four total flips, even
the original-likelihood posterior remains near.5: the strong sparse prior,
not merely VI numerics, contributes to this behavior.

Inference: a more accurate optimizer or quadrature alone cannot be assumed
to rescue early predictions. This does not prove that the full screen's
failure is solely prior mismatch; competition/dependence among255 features
and publication cadence remain untested causes. Investigate those separately
before choosing a new inference family or evidence-driven prior adaptation.
Do not tune the fixed pilot prior against these results.

Command: `go test ./internal/observationlearners -run '^TestSpikeExactPosteriorDiagnostic$' -count=1 -v`.
Test0.30s/package0.670s on the research host; a diagnostic, not a production
latency measurement. Component file: `spike_slab_exact_diagnostic_test.go`.
