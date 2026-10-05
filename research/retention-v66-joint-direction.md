# Next Lead: Calibrated Observation Value And Regime Adaptation

PROPOSAL ONLY. Do not change the current model based on this document alone.

V63 solves the 200-member constructor allocation failure without shrinking
the frontier. V64's full actual-mixture diagnostic improves over rejected V60
but fails all scientific gates except runtime/memory. V65 identifies a separate
mathematical obstacle: p*q1+(1-p)*q0 is not the current mixture forecast, with
maximum defect .007763602 in the small public-API diagnostic. A single-expert
posterior-variance formula therefore cannot be inherited as a guaranteed
proper-risk gain for this adaptive ensemble.

## Research Grounding

Raftery, Karny and Ettler's [Dynamic Model Averaging](https://pmc.ncbi.nlm.nih.gov/articles/PMC2895940/)
combines changing models and recursive parameter estimates, with a simplifying
independent-model update approximation. It is relevant to bounded online
aggregation, not evidence that our delayed measurements and window forgetting
share a coherent joint law or that their guarantees transfer.

Van Erven, Grunwald and de Rooij's [switch distribution](https://arxiv.org/abs/0807.1005)
addresses Bayesian catch-up through switching among predictive models. Its
efficient prequential construction motivates explicit sequence-law accounting;
its consistency/rate results are not a guarantee for this finite, selectively
observed, misspecified EventFrame mixture.

## Falsifiable Next Steps

1. Build a SMALL independently enumerable joint sequence model over regime,
   member rate, noise, first/paired evidence and future prediction target. Bind
   observation nomination, hypothetical branches and scored outputs to the SAME
   model. Verify joint normalization and tower identities before optimization.
   Keep it bounded and include delayed original-position factor replacement.
2. Compare that reference with a computationally bounded filtering/mixture
   approximation. Explicitly measure approximation/calibration defects rather
   than claiming equality from separate declared kernels. If coherence cannot
   be retained within the budget, label acquisition as an empirical utility
   heuristic and test it against random/uncertainty at equal TOTAL cost.
3. Isolate regime/noise/prior misspecification from ensemble retention. Use
   incumbent-preserving negative controls and no-observation ablations; do not
   attribute V64 harm uniquely to selector switching. Do not pick another
   window/forgetting factor from these outcomes and call the next seed fresh.
4. Freeze the surviving candidate and repeat across independently generated
   fitting/noise/shift/delay populations before any reserved confirmation.
   Preserve all old failures, unchanged gates and outcome-volume/cost accounting.

This lead concerns parts of goals 1/2/4/7, not the complete objective. Useful
externally valid Anti-Pigeon decisions (3), untouched labeled agent utility (5)
and loaded durable freshness/serving (6) remain required. No private or sealed
labels, production changes, whitepaper edits or publication are authorized by
this proposal. All seven whole goals remain OPEN and viable leads remain.
