# Finite contextual expert model

Research reference, not a novelty claim or production replacement. The causal
effect, Anti-Pigeon admissibility and semantic-group guarantees of EventFrame
are not established by this conditional predictor.

Let H_t contain only admitted outcomes from previously issued forecasts. For
each input mask m, and each projected cell c, draw an independent expert index
k_c with prior pi. One global mask is drawn with prior a. The predictable expert
probabilities p_{k,j} are recorded at issue time. Each observed binary outcome
y_j has likelihood l_{j,k}=p_{k,j}^{y_j}(1-p_{k,j})^{1-y_j} conditional on its
cell's expert. The model conditions on the as-of inputs and issued forecasts,
not on future forecast values.

For c=x_j AND m, define

`S_{m,c,k}(t) = sum_{j in H_t: x_j AND m=c} log(l_{j,k})`.

Then `Z_{m,c}=sum_k pi_k exp(S_{m,c,k})`, `Z_m=product_c Z_{m,c}`,
`w_m=a_m Z_m / sum_n a_n Z_n`, and
`v_{m,c,k}=pi_k exp(S_{m,c,k}) / Z_{m,c}`.
Unseen cells have Z=1 and v=pi. The next Bernoulli probability is

`p_t = sum_m w_m sum_k v_{m,x_t AND m,k} p_{k,t}`.

This uses the same likelihood for evidence and prediction; the expert index
is not fitted on unrelated evidence and attached to an arbitrary forecast law.
Each observation enters once within any fixed mask hypothesis. Alternative
hypotheses each evaluating that observation is Bayesian model comparison,
not extra independent evidence. This does not make the model correctly
specified for the external generator or establish calibration.

The global control restricts m=0. The contextual prior places.95 on m=0 and
spreads.05 over nonempty masks using the existing1/3 inclusion convention,
conditioned on being nonempty. The expert prior is shared between control and
candidate. Both use the same past issued law tape and feedback schedule.
Initial16 training examples belong to the source expert fits; neither router
has issued-law records for those examples, so neither uses them as extra
meta-training evidence. No teacher probabilities or simulator regimes enter
the update.

## Implementation invariants

Store log likelihoods separately from normalized cell weights. A tiny weight
may numerically underflow for prediction, but its log evidence is retained so
later contradictory observations can revive it. Update only the one cell
matching an outcome under each mask. The change in that cell's log evidence
updates the mask's log evidence. Log-sum-exp normalizes both levels.

Issue records own their input/probability copy. Deliveries are idempotent for
the same outcome and reject conflicts or expired records. Missing outcomes
never become negative observations. A fixed set of delivered packets yields
the same mathematical state regardless of delivery order; floating-point
summation need not be bit-identical across different orders. As-of predictions
must still respect actual arrival times.

The bounded research implementation has at most256 issues, eight experts and
nine input bits. Storage is O(K*3^d + 4^d + T*K): the 4^d term is a shared
bit/mask-to-cell index. Per delivery and prediction cost is O(K*2^d); metadata
construction costs O(d*4^d). This is much more work than global expert weights.
Do not describe it as a constant-cost production hot path. No serving timing
has been established.

## What would falsify the lead

If conditional routing does not improve on the matched global predictor while
preserving established Markov performance, the extra context structure is not
justified on this workload. A large posterior non-global mass by itself is not
success: it can reflect spurious specialization. Preserve stationary controls,
delayed switches and every transfer family. Even a consumed-data pass needs
fresh confirmation and subsequent partial-observation and real-agent tests.
