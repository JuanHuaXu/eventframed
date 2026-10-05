# Context reliability v21 results

**FAILED the frozen diagnostic gate.** Exact observed-context conditioning
improves error enrichment relative to the class-level v20 warning, but does not
provide the required error coverage. No production mechanism was changed.

See [protocol](mmm-context-reliability-v21-protocol.md) and
[artifact](mmm-context-reliability-v21.json). Both replay consumed v19 traces;
the partition named confirmation is not new confirmation for this diagnostic.

Clustered majority shift128, confirmation-labelled Post partition:

| Quantity | Result | Gate |
| --- | ---: | ---: |
| Warned frames | 432/4264 = 10.13% | <=50% |
| Errors captured | 64/361 = 17.73% | >=50% |
| Warned error | 64/432 = 14.81% | >=2 times nonwarned |
| Nonwarned error | 297/3832 = 7.75% | Comparator |
| Unknown support | 393/4264 = 9.22% | Reported, not excluded |

The error-rate ratio is approximately 1.91, below the required 2. The more
substantial failure is coverage: most errors remain outside the warned set.
Unknown cases contribute 54 errors, while clear cases contribute 243. Thus low
support alone does not explain the missing error coverage.

This is not a test of whether additional observations improve warned forecasts.
There is no new Brier gain, online policy comparison, latency benchmark, or
validity certificate. Context keys are local acquired fields and values, not
latent regimes or full hidden inputs. The model uses nominal-centered shrinkage,
not ordinary Bayesian evidence and not a time-uniform confidence bound.

Two tests check support, class/value/mask separation, history/key caps, read-only
recency, and exclusion of current/future labels and hidden input. Complete
576-record replay and embedded-source binding are checked against the artifact.
All scenario groups, including stationary and delayed/missing feedback, are
retained rather than screening on the favorable subgroup alone.

## Research decision

Do not tune this gate on v19 or enable a general observation trigger. The narrow
error-enrichment signal could motivate a future costed intervention, but it does
not meet this research screen and does not establish the value of another read.
Further threshold variants are lower priority than the unresolved real-task
representation and actual learner-in-shadow integration requirements. Neither
the seven-direction objective nor the observation research is complete.
