# Conditional next lead: infer which observations belong to the current regime

Initial design written while v119 ran, before its quality results. The
[Go component](segment-posterior-component.md) now passes numerical checks;
fresh quality remains untested. This does not retroactively rescue v119 or
justify tuning its consumed confirmation data.

The current MAP/variational comparison changes coefficient uncertainty, while
both models still fit the same last64/32 eligible labels. Previous Markov
experiments route among fitted experts; they do not infer a posterior over the
training segment itself. A search of the observationlearners implementation and
research notes found no run-length posterior implementation in that path.

Adams and MacKay's [Bayesian Online Changepoint Detection](https://www.cs.princeton.edu/~rpa/pubs/adams2007changepoint.pdf),
Sections1-2 and Algorithm1, marginalizes predictions over possible run lengths
under independent segment parameters. Section2.4 gives growing worst-case cost;
truncating probability tails is an approximation, not a hard constant-cost
guarantee. Its ordered-feedback formulation does not automatically cover our
delayed labels. The adaptation below must be verified separately.

## Possible delayed-evidence formulation

At decision clock t, freeze a bounded history [l,t-1] and the label-origin set
I_t actually eligible under the common evidence budget. Features are already
observed. For an interval [a,b], define its marginal likelihood

M_t(a,b) = sum_k prior(k) integral product_{j in I_t intersect [a,b]}
 p_k(y_j | x_j,theta) prior_k(dtheta).

An interval with no available labels has likelihood1. Do not insert predicted
labels, treat missing labels as negatives, or multiply them twice after arrival.
Initially use only models with genuine computable marginal likelihoods; a MAP
training objective or variational bound is not interchangeable with M_t.

For a declared constant hazard h in(0,1), put F(l-1)=1 and

F(b) = sum_{a=l}^b F(a-1) c(a) (1-h)^(b-a) M_t(a,b),
c(l)=1, c(a)=h for a>l.

The normalized last-segment weights at t-1 are the corresponding terms of
F(t-1), divided by F(t-1). The next-label predictive is h times the new-segment
prior predictive plus (1-h) times the weighted existing-segment predictive.
This convention places a new boundary BEFORE the next observation. Spell it out
to avoid the off-by-one ambiguity between run-length conventions.

Recompute from the as-of eligible set when delayed labels arrive. This is a
bounded retrospective calculation for a prospective forecast, not a claim that
the usual one-step recursion remains exact under arbitrary late feedback.
Older labels excluded by the cap are intentionally outside the retained model.
Selection depending on labels would need a selection model; it is not supplied
by this formula. A forced left boundary is also a declared approximation.

## Required tests before a quality run

- Compare partition sums and next forecasts to exhaustive cut-pattern enumeration
  on tiny histories, including empty evidence, unequal gaps and delayed labels.
- Verify missing/unrevealed label changes cannot alter issued forecasts; duplicate
  arrivals cannot increase likelihood or support. Preserve original forecast IDs.
- Use log-space arithmetic, explicit positive priors and finite resource caps.
  Do not claim exact unbounded-history inference after truncation.
- Match the same eligible labels and acquisition charges across controls. A cap
  on elapsed frames and a cap on observed labels are not equivalent under delays.
- Benchmark marginal-likelihood construction as well as dynamic programming;
  segment costs must not be hidden in a nominal quadratic recurrence.
- Retain stationary, gradual, abrupt, Boolean and null controls. A piecewise-
  stationary model may improve abrupt recovery and still fail gradual changes.

If v119 shows useful uncertainty gains, test incumbent-preserving composition
first. If stale-evidence recovery remains the dominant deficit, this is a distinct
mechanism lead for direction2, not permission to tune the consumed confirmation
data or drop failed families. Actual implementation and fresh validation remain
required. No effect on production, current scored laws or the whitepaper.

## Tiny-model mathematical check

The [standalone reference script](segment-posterior-reference.mjs) checks the
recurrence against exhaustive cut-pattern enumeration for every outcome and
available-label mask at lengths0..6, three hazards and two query contexts:
32766 comparisons. The toy segment prior mixes a pooled Beta-Bernoulli model
and two context-specific Beta-Bernoulli models. The independent reference uses
closed half-integer beta integrals rather than sequential likelihood products.
Maximum evidence difference1.23e-15; prediction difference1.89e-15. Changing
all unavailable labels leaves every result identical. Invalid bounds reject.

[Recorded output](../docs/experiments/mmm-segment-posterior-reference.json)
includes the source hash and result digest. This checks a finite formulation,
not predictive improvement, the intended richer learners, delayed journal
ownership, bounded live truncation or serving performance. The subsequent Go
component is a separate implementation checkpoint; v119 remains unchanged.
