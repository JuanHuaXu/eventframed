# Online Brier aggregation after v93

Proposal only, not implemented or validated. Static Bayesian odds and LOO
stacking can both endorse a misleading finite sample. Instead of another
static prior/penalty sweep, test cumulative realized-loss control with honest
online feedback. This changes the learning protocol; it does not retroactively
rescue v92/v93's static held-out-risk failures.

## Primary sources and normalization

Vovk and Zhdanov,
[Prediction With Expert Advice For The Brier Game](https://www.jmlr.org/papers/volume10/vovk09a/vovk09a.pdf),
give a strong aggregating algorithm and cumulative-loss guarantee. Their
multiclass Brier loss sums over outcomes. For a binary outcome it is twice our
single-coordinate loss (p-y)^2. Do not copy their rate or optimality constant
without this conversion. The simple convex-mixture construction below is a
more conservative specialization, justified directly rather than claimed to
be their optimal substitution function.

## Proposed immediate-feedback construction

Let each expert issue p_{k,t} in[0,1] before y_t is observed. The generic expert
has initial mass w_{G,0}=0.95; remaining bounded challenger experts share0.05.
For eta=1/2, issue and update

$$
p_t=\sum_k w_{k,t-1}p_{k,t},\qquad
w_{k,t}=\frac{w_{k,t-1}\exp[-\eta(p_{k,t}-y_t)^2]}
{\sum_jw_{j,t-1}\exp[-\eta(p_{j,t}-y_t)^2]}.
$$

These are online expert weights, not posterior truth probabilities. Predictions
may come from evolving algorithms, but must be available before their outcomes.
Do not reset weights on every fit or reuse future validation outcomes.

For fixed y, the second derivative of exp[-eta(p-y)^2] equals

$$
\exp[-\eta(p-y)^2]\{4\eta^2(p-y)^2-2\eta\}\le0
\quad (0<\eta\le1/2).
$$

Concavity and Jensen give the one-step potential inequality. Multiplying those
inequalities and lower-bounding the final potential by the generic expert's
term yields the pathwise bound

$$
\sum_{t=1}^{N}\big[(p_t-y_t)^2-(p_{G,t}-y_t)^2\big]
\le\frac{\log(1/w_{G,0})}{\eta}=0.1025865888\ldots.
$$

This proof requires the exact sequential updates just defined. It bounds
cumulative realized loss against the included generic expert, not every
individual prediction, future population risk or classification accuracy.
At N=16 the average bound is about0.006412. The0.95 mass is chosen to place
that finite-prefix bound below0.01, not selected by re-scoring v93 outcomes.
This is a proposed testable invariant, not an empirical success result.

## Boundaries that cannot be skipped

- The first experiment needs immediate complete feedback and common declared
  observations. It does not prove the guarantee for an independently acquiring
  generic policy if the expert only sees a different candidate-chosen mask.
- Unreceived labels cannot update weights. The earlier journal's stale-feedback
  skips, delays and censoring invalidate the simple telescoping proof as stated.
- [Joulani, Gyorgy and Szepesvari, Online Learning under Delayed Feedback](https://proceedings.mlr.press/v28/joulani13.pdf)
  provide explicit delayed-feedback reductions and distinguish adversarial from
  stochastic costs. Their results motivate a separate derivation/implementation;
  they are not inherited by the current journal automatically.
- Stable expert identities can denote evolving predictable algorithms. Actual
  resets or expert replacement must carry an explicit new comparison and regret
  budget. Repeated restarts cannot claim the single0.10259 lifetime bound.
- Adapting acquisition changes costs and the comparator. Measure shared-view
  quality first, then observation-policy comparisons with all reads charged.
- A cumulative bound can still permit an unacceptable one-off event. Keep
  high-priority admission, Anti-Pigeon, provenance and abstention protections.

## Next falsifiable experiment

Implement and independently verify the potential inequality and prefix bound
on adversarial sequences, ties, extreme probabilities and changing predictable
experts. Include a test that demonstrates why delayed/skipped updates are not
covered. Freeze fresh prospective streams, compare generic-only, skeptical BMA
and the online mixture, and measure recovery after changes and false switches.
No known rule, simulator oracle, future outcome or lost label may enter updates.

Preserve static v93 failures and the consumed tail. Online gains must include
the observation/feedback cost and cannot be substituted for the unchanged
whole-direction, delayed-learning or real-world success requirements.
