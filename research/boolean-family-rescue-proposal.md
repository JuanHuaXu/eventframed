# Proposed generic/Boolean family rescue after v90

Status: proposal only. The standalone specialist is implemented, but the v90
advance conjunction failed and its non-parity regressions rule out replacement.
Do not relabel v90 as a pass. This proposal needs fresh tests before integration.

## Evidence and falsifiable response

The specialist fits compact noisy parity well, but cannot represent majority or
multiplexer conditional laws in its individual rule hypotheses. The existing
subset family can. Retain both through one latent model-family choice rather
than a fixed average or a target-name switch. This is Bayesian model averaging,
not two independent observations of the same training evidence.

Let F be either G (generic conditional-cell family) or P (noisy parity family).
Declare P(F=G)=P(F=P)=1/2. Within each family retain the existing normalized
inclusion1/3 subset prior pi_S and proper Beta(1/2,1/2) parameter priors.
Both families use the same uniform input law; their label likelihoods differ.
For a supplied eligible training sequence D define

$$
Z_F(D)=\sum_S \pi_S\int L_F(D\mid S,\theta_F)
\,d\Pi_F(\theta_F),
\qquad
w_F(D)=\frac{Z_F(D)}{Z_G(D)+Z_P(D)}.
$$

In G, theta_F contains one Bernoulli parameter per subset assignment, and Z_G
uses the product of per-cell Beta integrals already used in fitSubset. In P,
one agreement-rate parameter per rule supplies the Beta integral from v90.
Both are sequence likelihoods with no binomial coefficient. The common input
likelihood cancels in family odds. Evaluate normalizers with log-sum-exp.

$$
Q(Y=1\mid x,D)=w_G(D)Q_G(Y=1\mid x,D)
+w_P(D)Q_P(Y=1\mid x,D).
$$

Compile the resulting full law into the same conditional representation. Since
the input law is shared and independent of family, partial marginalization
commutes with this mixture. If future families have different input laws, the
observed input also changes the family weights; this formula cannot be reused
unchanged. A sliding audit window is a declared local working model, not an
exact posterior over a nonstationary lifetime history.

## Required checks before a new experiment

- Verify each marginal likelihood against independent Beta integrals, family
  weight normalization, and full/partial mixture identity. Preserve v90 replay
  by adding separate sources rather than changing its frozen constructors.
- Test identical evidence exactly once in each alternative family. Family
  averaging is not multiplication of their likelihoods as independent sources.
- Test sparse/impoverished support, dependent inputs, non-parity rules and null
  labels. The input-law mismatch is not repaired by averaging outcome families.
- Use new fit seeds and record all variants. Do not reuse consumed v90 fits as
  confirmation or use the generating rule to select a family.
- Preserve original generic, age and incumbent learners during any later
  delayed-stream test; ordinary posterior motion still needs its publication
  and residual-certificate treatment.
- Charge both fits and compilation; shared tables or algebraic optimizations
  need numerical parity tests, not assumed equivalence.

## Avoid repeating the protocol mistake

The fixed0.01 improvement threshold in v90 is impossible on its n64 parity3
cells: the control's entire excess over the exact simulator risk is below0.01.
Keep that test failure in the record. A separately frozen experiment can judge
excess-risk reduction where headroom exists and non-inferiority near the noise
floor, using a declared rule and fresh data. Do not change the existing v88
delayed-stream or whole-direction success criteria on this basis.

The next component test must also constrain the large majority/multiplexer
regressions seen in v90. If family evidence still overweights the wrong model,
investigate held-out/prequential family selection rather than adding a rule-name
exception. Component success alone cannot authorize delayed-stream adoption.
